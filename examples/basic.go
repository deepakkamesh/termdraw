package main

import (
	"log"

	"github.com/deepakkamesh/termdraw"
)

func main() {

	images, err := termdraw.LoadImages("walle_normal.png", "walle_happy.png")
	if err != nil {
		log.Fatalf("Failed to load images: %v", err)
	}

	td := termdraw.New()
	td.Animate(images, '*', 200)
	if err := td.Init(); err != nil {
		panic(err)
	}
	td.Run()
	for {
	}
}
