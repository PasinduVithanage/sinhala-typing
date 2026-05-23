//go:build windows

package tray

import (
	"encoding/binary"
	"math"

	"image/color"
)

func iconBytes(state IconState) []byte {
	switch state {
	case IconActive:
		return circleIconDIB(color.RGBA{R: 0x2E, G: 0xAD, B: 0x6E, A: 0xFF}) // green
	case IconError:
		return circleIconDIB(color.RGBA{R: 0xE5, G: 0x53, B: 0x4B, A: 0xFF}) // red
	default:
		return circleIconDIB(color.RGBA{R: 0x80, G: 0x80, B: 0x9A, A: 0xFF}) // gray
	}
}

// circleIconDIB renders a 32×32 circle in the Windows RT_ICON resource format
// (BITMAPINFOHEADER + 32-bpp BGRA XOR mask + 1-bpp AND mask) that
// CreateIconFromResourceEx expects.  PNG does not work with that API.
func circleIconDIB(fill color.RGBA) []byte {
	const sz = 32

	// BITMAPINFOHEADER (40 bytes).
	// biHeight is doubled: upper half = XOR mask, lower half = AND mask.
	hdr := make([]byte, 40)
	binary.LittleEndian.PutUint32(hdr[0:], 40)   // biSize
	binary.LittleEndian.PutUint32(hdr[4:], sz)   // biWidth
	binary.LittleEndian.PutUint32(hdr[8:], sz*2) // biHeight (doubled)
	binary.LittleEndian.PutUint16(hdr[12:], 1)   // biPlanes
	binary.LittleEndian.PutUint16(hdr[14:], 32)  // biBitCount
	// bytes 16–39 stay zero (BI_RGB, no palette)

	// XOR mask: 32-bpp BGRA, bottom-up row order.
	xor := make([]byte, sz*sz*4)
	cx := float64(sz)/2 - 0.5
	cy := float64(sz)/2 - 0.5
	radius := cx - 1.5

	for y := 0; y < sz; y++ {
		for x := 0; x < sz; x++ {
			row := sz - 1 - y // flip to bottom-up
			idx := (row*sz + x) * 4
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist <= radius {
				alpha := uint8(255)
				if dist > radius-1.0 {
					alpha = uint8(255 * (radius - dist))
				}
				xor[idx+0] = fill.B
				xor[idx+1] = fill.G
				xor[idx+2] = fill.R
				xor[idx+3] = alpha
			}
		}
	}

	// AND mask: 1-bpp, DWORD-padded rows, bottom-up.
	// All zeros = show XOR mask pixel (32-bpp alpha handles transparency).
	andRowBytes := ((sz + 31) / 32) * 4 // 4 bytes for sz=32
	and := make([]byte, sz*andRowBytes)

	out := make([]byte, 0, len(hdr)+len(xor)+len(and))
	out = append(out, hdr...)
	out = append(out, xor...)
	out = append(out, and...)
	return out
}
