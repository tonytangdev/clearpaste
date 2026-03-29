//go:build ignore

package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

func createIcon(path string, c color.RGBA) {
	img := image.NewRGBA(image.Rect(0, 0, 22, 22))
	for x := 0; x < 22; x++ {
		for y := 0; y < 22; y++ {
			img.Set(x, y, c)
		}
	}
	f, _ := os.Create(path)
	defer f.Close()
	png.Encode(f, img)
}

func main() {
	os.MkdirAll("icons", 0o755)
	createIcon("icons/icon.png", color.RGBA{100, 100, 100, 255})
	createIcon("icons/icon_active.png", color.RGBA{0, 200, 100, 255})
	createIcon("icons/icon_disabled.png", color.RGBA{180, 180, 180, 255})
}
