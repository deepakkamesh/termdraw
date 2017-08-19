package main

import (
	"flag"
	"log"
	"time"

	"github.com/deepakkamesh/termdraw"
	termbox "github.com/nsf/termbox-go"
)

func main() {
	flag.Parse()

	speaking, err := termdraw.LoadImages(
		"walle_normal.png",
		"walle_speaking_small.png",
		"walle_speaking_med.png",
		"walle_speaking_large.png",
	)

	if err != nil {
		log.Fatalf("Failed to load images: %v", err)
	}

	blinking, err := termdraw.LoadImages(
		"walle_normal.png",
		"walle_normal_eye_small.png",
	)

	if err != nil {
		log.Fatalf("Failed to load images: %v", err)
	}

	td := termdraw.New()
	if err := td.Init(); err != nil {
		panic(err)
	}
	td.Run()

	td.Animate(blinking, '*', 500*time.Millisecond)

	_ = speaking
	_ = blinking
	// Handle events from termbox.
	for {
		evt := <-td.EventCh
		if evt.Type == termbox.EventKey &&
			evt.Key == termbox.KeyEsc {
			td.Quit()
			break
		}
	}

	// This is needed for termbox to cleanup properly. (not sure why?)
	t := time.NewTimer(1 * time.Millisecond)
	<-t.C
}
