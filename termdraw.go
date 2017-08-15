package termdraw

import (
	"image"
	"image/png"
	"os"
	"time"

	termbox "github.com/nsf/termbox-go"
)

type imageData struct {
	Xmax int
	Ymax int
	data [][]bool
}

type Term struct {
	eventChan chan termbox.Event
	quit      chan struct{}
	curr      uint
	images    []imageData
	ch        rune
	tick      *time.Ticker
}

// New returns an initialized Term.
func New() *Term {
	return &Term{
		eventChan: make(chan termbox.Event),
		quit:      make(chan struct{}),
		tick:      time.NewTicker(200 * time.Millisecond),
		images:    []imageData{},
		curr:      0,
	}
}

func LoadImages(images ...string) ([]image.Image, error) {
	imgList := []image.Image{}

	for _, imgFile := range images {
		f, err := os.Open(imgFile)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		img, err := png.Decode(f)
		if err != nil {
			return nil, err
		}
		imgList = append(imgList, img)
	}
	return imgList, nil
}

func (s *Term) Init() error {

	// Initialize termbox.
	if err := termbox.Init(); err != nil {
		return err
	}
	return nil
}

func (s *Term) Animate(imgs []image.Image, ch rune, d time.Duration) {
	//s.tick = time.NewTicker(d)
	s.ch = ch

	for _, img := range imgs {

		// Allocate Array.
		data := make([][]bool, img.Bounds().Max.Y)
		for j := range data {
			data[j] = make([]bool, img.Bounds().Max.X)
		}

		// Mark X,Y coordinates which are opaque
		for y := 0; y < img.Bounds().Max.Y; y++ {
			for x := 0; x < img.Bounds().Max.X; x++ {
				_, _, _, a := img.At(x, y).RGBA()
				if a > 0 {
					data[y][x] = true
					continue
				}
				data[y][x] = false
			}
		}

		s.images = append(s.images, imageData{
			Xmax: img.Bounds().Max.X,
			Ymax: img.Bounds().Max.Y,
			data: data,
		})
	}
}

func (s *Term) draw() {
	w, h := termbox.Size()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if s.images[s.curr].Ymax > y && s.images[s.curr].Xmax > x && s.images[s.curr].data[y][x] {
				termbox.SetCell(x, y, s.ch, termbox.ColorDefault, termbox.ColorDefault)
			}
		}
	}
	termbox.Flush()
}

func (s *Term) Run() {

	// Start termbox poller.
	go func() {
		for {
			s.eventChan <- termbox.PollEvent()
		}
	}()

	go s.EventLoop()
}

func (s *Term) EventLoop() {
	i := 0

	for {
		select {

		// Handle events from termbox.
		case evt := <-s.eventChan:
			switch evt.Type {
			case termbox.EventKey:
				if evt.Key == termbox.KeyEsc {
					go func() {
						s.quit <- struct{}{}
					}()
				}
			}

		case <-s.tick.C:
			if i == len(s.images) {
				i = 0
			}
			s.curr = uint(i)
			s.draw()
			i++

		case <-s.quit:
			termbox.Sync()
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			termbox.Flush()
			termbox.Close()
			return
		}
	}
}

func (s *Term) Quit() {
	s.quit <- struct{}{}
}
