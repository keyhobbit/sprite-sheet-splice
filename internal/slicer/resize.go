package slicer

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
)

// ResizeImage scales src to newWidth x newHeight using nearest-neighbor
// interpolation and returns the result as a PNG-encoded buffer.
func ResizeImage(src image.Image, newWidth, newHeight int) (*bytes.Buffer, error) {
	if newWidth <= 0 || newHeight <= 0 {
		return nil, fmt.Errorf("invalid resize dimensions: %dx%d", newWidth, newHeight)
	}

	sb := src.Bounds()
	srcW := sb.Dx()
	srcH := sb.Dy()

	dst := image.NewNRGBA(image.Rect(0, 0, newWidth, newHeight))

	for y := 0; y < newHeight; y++ {
		srcY := y * srcH / newHeight
		if srcY >= srcH {
			srcY = srcH - 1
		}
		for x := 0; x < newWidth; x++ {
			srcX := x * srcW / newWidth
			if srcX >= srcW {
				srcX = srcW - 1
			}
			r, g, b, a := src.At(sb.Min.X+srcX, sb.Min.Y+srcY).RGBA()
			dst.SetNRGBA(x, y, color.NRGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
		}
	}

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, dst); err != nil {
		return nil, fmt.Errorf("failed to encode resized image: %w", err)
	}
	return buf, nil
}
