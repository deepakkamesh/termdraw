# termdraw

Termdraw displays image files on console.

## Example Usage
```Golang
package main

import (
        "image"
        "image/png"
        "os"

        "github.com/deepakkamesh/termdraw"
)

func main() {

        f2, err := os.Open("walle_normal.png")
        if err != nil {
                panic(err)
        }
        defer f2.Close()

        img2, err := png.Decode(f2)
        if err != nil {
                panic(err)
        }

        f3, err := os.Open("walle_happy.png")
        if err != nil {
                panic(err)
        }
        defer f3.Close()

        img3, err := png.Decode(f3)
        if err != nil {
                panic(err)
        }

        td := termdraw.New()
        td.Animate([]image.Image{img2, img3}, '*', 200)
        if err := td.Init(); err != nil {
                panic(err)
        }
        td.Run()
        for {
        }

}
```
