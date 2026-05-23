//go:build windows

package tray

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
)

func iconBytes(state IconState) []byte {
	switch state {
	case IconActive:
		return circleIcon(color.RGBA{R: 0x2E, G: 0xAD, B: 0x6E, A: 0xFF}) // green
	case IconError:
		return circleIcon(color.RGBA{R: 0xE5, G: 0x53, B: 0x4B, A: 0xFF}) // red
	default:
		return circleIcon(color.RGBA{R: 0x80, G: 0x80, B: 0x9A, A: 0xFF}) // gray
	}
}

// circleIcon renders a 22×22 NRGBA circle PNG with the given fill colour.
func circleIcon(fill color.RGBA) []byte {
	const sz = 22
	img := image.NewNRGBA(image.Rect(0, 0, sz, sz))
	cx := float64(sz-1) / 2
	cy := float64(sz-1) / 2
	r := cx - 1.0 // leave 1px transparent border

	for y := 0; y < sz; y++ {
		for x := 0; x < sz; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist <= r {
				// Simple anti-alias: reduce alpha near the edge.
				alpha := uint8(0xFF)
				if dist > r-1.0 {
					alpha = uint8(0xFF * (r - dist))
				}
				img.SetNRGBA(x, y, color.NRGBA{R: fill.R, G: fill.G, B: fill.B, A: alpha})
			}
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
