//go:build ignore

package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

// px sets a pixel with the given alpha (0-255) using the base color.
func px(img *image.RGBA, x, y int, c color.RGBA) {
	if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
		img.Set(x, y, c)
	}
}

func solid(a uint8) color.RGBA { return color.RGBA{0, 0, 0, a} }
func green(a uint8) color.RGBA { return color.RGBA{34, 197, 94, a} }
func gray(a uint8) color.RGBA  { return color.RGBA{160, 160, 160, a} }

// drawClipboard draws a clipboard outline on the image using the given color function.
func drawClipboard(img *image.RGBA, col func(uint8) color.RGBA) {
	f := col(255) // full opacity
	s := col(180) // semi for anti-aliasing

	// Clipboard body: 12w x 14h, offset (5,6) to (16,19)
	// Top edge
	for x := 5; x <= 16; x++ {
		px(img, x, 6, f)
	}
	// Bottom edge
	for x := 5; x <= 16; x++ {
		px(img, x, 19, f)
	}
	// Left edge
	for y := 6; y <= 19; y++ {
		px(img, 5, y, f)
	}
	// Right edge
	for y := 6; y <= 19; y++ {
		px(img, 16, y, f)
	}

	// Clip at top: centered tab
	// Tab body: 6w x 3h at (8, 3)
	for x := 8; x <= 13; x++ {
		px(img, x, 3, f)
	}
	for x := 8; x <= 13; x++ {
		px(img, x, 4, f)
	}
	// Tab sides going down to meet board
	px(img, 8, 5, f)
	px(img, 13, 5, f)
	// Tab connects to board top
	px(img, 8, 6, f)
	px(img, 13, 6, f)

	// Small circle/ring on clip (the metal ring part)
	px(img, 10, 2, s)
	px(img, 11, 2, s)
	px(img, 9, 3, f)
	px(img, 12, 3, f)

	// Content lines on clipboard (representing text)
	for x := 7; x <= 14; x++ {
		px(img, x, 9, col(140))
	}
	for x := 7; x <= 12; x++ {
		px(img, x, 11, col(140))
	}
	for x := 7; x <= 14; x++ {
		px(img, x, 13, col(140))
	}
	for x := 7; x <= 10; x++ {
		px(img, x, 15, col(140))
	}
}

// drawSparkle draws a small 4-point sparkle at (cx, cy).
func drawSparkle(img *image.RGBA, cx, cy int, col func(uint8) color.RGBA) {
	f := col(255)
	s := col(160)

	// Center
	px(img, cx, cy, f)

	// Cardinal points
	px(img, cx, cy-1, f)
	px(img, cx, cy-2, s)
	px(img, cx, cy+1, f)
	px(img, cx, cy+2, s)
	px(img, cx-1, cy, f)
	px(img, cx-2, cy, s)
	px(img, cx+1, cy, f)
	px(img, cx+2, cy, s)

	// Diagonal accents
	px(img, cx-1, cy-1, s)
	px(img, cx+1, cy-1, s)
	px(img, cx-1, cy+1, s)
	px(img, cx+1, cy+1, s)
}

// drawCheck draws a small checkmark at (cx, cy).
func drawCheck(img *image.RGBA, cx, cy int, col func(uint8) color.RGBA) {
	f := col(255)
	px(img, cx-2, cy, f)
	px(img, cx-1, cy+1, f)
	px(img, cx, cy+2, f)
	px(img, cx+1, cy+1, f)
	px(img, cx+2, cy, f)
	px(img, cx+3, cy-1, f)
}

func createIcon(path string, draw func(*image.RGBA)) {
	img := image.NewRGBA(image.Rect(0, 0, 22, 22))
	draw(img)
	f, _ := os.Create(path)
	defer f.Close()
	png.Encode(f, img)
}

func main() {
	os.MkdirAll("icons", 0o755)

	// Normal: black clipboard on transparent (macOS template-style)
	createIcon("icons/icon.png", func(img *image.RGBA) {
		drawClipboard(img, solid)
	})

	// Active: green clipboard with sparkle (just cleaned!)
	createIcon("icons/icon_active.png", func(img *image.RGBA) {
		drawClipboard(img, green)
		drawSparkle(img, 18, 3, green)
	})

	// Disabled: gray clipboard, lower contrast
	createIcon("icons/icon_disabled.png", func(img *image.RGBA) {
		drawClipboard(img, gray)
	})
}
