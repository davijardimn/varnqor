package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

type point struct{ x, y int }

func main() {
	const size = 1024
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: 11, G: 12, B: 16, A: 255}}, image.Point{}, draw.Src)

	// Three routed token units. The geometry is intentionally simple so the
	// mark remains recognisable at 16–32 px in Windows UI.
	fillGradient(img, []point{{220, 303}, {445, 198}, {575, 198}, {804, 303}, {804, 361}, {575, 468}, {445, 468}, {220, 361}},
		color.RGBA{R: 124, G: 92, B: 252, A: 255}, color.RGBA{R: 99, G: 102, B: 241, A: 255})
	fillGradient(img, []point{{220, 436}, {449, 543}, {573, 543}, {804, 436}, {804, 557}, {573, 665}, {449, 665}, {220, 557}},
		color.RGBA{R: 99, G: 102, B: 241, A: 255}, color.RGBA{R: 69, G: 72, B: 198, A: 255})
	fillGradient(img, []point{{313, 650}, {454, 717}, {570, 717}, {716, 649}, {796, 688}, {570, 797}, {451, 797}, {232, 689}},
		color.RGBA{R: 59, G: 130, B: 246, A: 255}, color.RGBA{R: 57, G: 63, B: 175, A: 255})

	if err := os.MkdirAll("build", 0o755); err != nil {
		panic(err)
	}
	file, err := os.Create("build/appicon.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		panic(err)
	}
}

func fillGradient(img *image.RGBA, polygon []point, top, bottom color.RGBA) {
	minX, minY := polygon[0].x, polygon[0].y
	maxX, maxY := minX, minY
	for _, p := range polygon[1:] {
		if p.x < minX { minX = p.x }
		if p.x > maxX { maxX = p.x }
		if p.y < minY { minY = p.y }
		if p.y > maxY { maxY = p.y }
	}
	height := maxY - minY
	if height < 1 { height = 1 }
	for y := minY; y <= maxY; y++ {
		t := float64(y-minY) / float64(height)
		c := color.RGBA{
			R: uint8(float64(top.R)*(1-t) + float64(bottom.R)*t),
			G: uint8(float64(top.G)*(1-t) + float64(bottom.G)*t),
			B: uint8(float64(top.B)*(1-t) + float64(bottom.B)*t),
			A: 255,
		}
		for x := minX; x <= maxX; x++ {
			if inside(point{x, y}, polygon) {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func inside(p point, polygon []point) bool {
	inside := false
	j := len(polygon) - 1
	for i := 0; i < len(polygon); i++ {
		pi, pj := polygon[i], polygon[j]
		if (pi.y > p.y) != (pj.y > p.y) && p.x < (pj.x-pi.x)*(p.y-pi.y)/(pj.y-pi.y)+pi.x {
			inside = !inside
		}
		j = i
	}
	return inside
}
