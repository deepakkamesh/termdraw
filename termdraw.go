package termdraw

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"time"

	termbox "github.com/nsf/termbox-go"
)

type imageUpdate struct {
	ch         rune
	d          time.Duration
	imagesData []imageData
}

type imageData struct {
	Xmax int
	Ymax int
	data [][]bool
}

type Term struct {
	EventCh    chan termbox.Event // Channel to send termbox event.
	quitLoop   chan struct{}      // Quit channel for main loop.
	quitPoller chan struct{}      // Quit channel for event loop.
	curr       uint               // pointer to the current index of the image displayed.
	images     []imageData        // list of images to cycle through for the animation.
	ch         rune               // Unicode char to render the image.
	tick       *time.Ticker       // ticker which controls the cycle rate of images.
	updateCh   chan *imageUpdate  // channel to send image updates.
	display    bool               // False to not display Animation.
}

// New returns an initialized Term.
func New() *Term {
	return &Term{
		EventCh:    make(chan termbox.Event),
		quitLoop:   make(chan struct{}),
		quitPoller: make(chan struct{}),
		tick:       time.NewTicker(100 * time.Millisecond),
		curr:       0,
		updateCh:   make(chan *imageUpdate),
		display:    true,
	}
}

// Display controls if the animation is displayed or not.
func (s *Term) Display(v bool) {
	s.display = v
}

// LoadImages processes list of png files and loads their data
// as image.Image.
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

// Init initializes termbox.
func (s *Term) Init() error {

	if s == nil {
		return fmt.Errorf("termdraw not initialized")
	}

	// Initialize termbox.
	if err := termbox.Init(); err != nil {
		return err
	}
	termbox.SetInputMode(termbox.InputEsc)
	return nil
}

// Animate processes image data and sends it to the main processing loop to update.
// This is done in the main loop to avoid race conditions; updating image data while
// its being displayed by draw func. Should be called after starting Run().
func (s *Term) Animate(imgs []image.Image, ch rune, d time.Duration) error {

	if s == nil {
		return fmt.Errorf("termdraw not initialized")
	}

	var imagesData []imageData

	for _, img := range imgs {

		// Allocate Array.
		data := make([][]bool, img.Bounds().Max.Y)
		for j := range data {
			data[j] = make([]bool, img.Bounds().Max.X)
		}

		// Mark X,Y coordinates which are opaque from A value.
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

		imagesData = append(imagesData, imageData{
			Xmax: img.Bounds().Max.X,
			Ymax: img.Bounds().Max.Y,
			data: data,
		})
	}

	s.updateCh <- &imageUpdate{
		ch:         ch,
		imagesData: imagesData,
		d:          d,
	}
	return nil
}

// draw updates the terminal with the image currently pointed by s.curr.
func (s *Term) draw() {
	if len(s.images) == 0 {
		return
	}
	w, h := termbox.Size()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if s.images[s.curr].Ymax > y && s.images[s.curr].Xmax > x && s.images[s.curr].data[y][x] {
				termbox.SetCell(x, y, s.ch, termbox.ColorDefault, termbox.ColorDefault)
			}
		}
	}
}

// Run starts the eventpoller and main update loop.
func (s *Term) Run() error {
	if s == nil {
		return fmt.Errorf("termdraw not initialized")
	}

	// Start termbox event poller.
	go func() {
		for {
			select {
			case <-s.quitPoller:
				close(s.EventCh)
				return
			default:
				s.EventCh <- termbox.PollEvent()
			}
		}
	}()

	go s.updateLoop()
	return nil
}

// updateLoop renders the image on the terminal.
func (s *Term) updateLoop() {
	defer termbox.Close()
	i := 0
	for {
		select {
		case upd := <-s.updateCh:
			i = 0
			s.images = upd.imagesData
			s.ch = upd.ch
			s.tick.Stop()
			s.tick = time.NewTicker(upd.d)

		case <-s.tick.C:
			if i == len(s.images) {
				i = 0
			}
			s.curr = uint(i)
			if s.display {
				s.draw()
			}
			termbox.Flush()
			i++

		case <-s.quitLoop:
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			termbox.Sync()
			termbox.Close()
			return
		}
	}
}

// Quit terminates the loop and exits termdraw.
func (s *Term) Quit() {
	// Called in a goroutine because of potential deadlock with
	// poller loop.
	go func() {
		s.quitLoop <- struct{}{}
		s.quitPoller <- struct{}{}
	}()
}
