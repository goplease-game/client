package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// TintImage returns a copy of src tinted with iconColor, preserving the
// original per-pixel brightness and alpha (e.g. for recoloring status icons).
func TintImage(src *ebiten.Image, iconColor color.Color) *ebiten.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	result := ebiten.NewImage(w, h)

	tr, tg, tb, _ := iconColor.RGBA()

	for y := range h {
		for x := range w {
			cr, cg, cb, ca := src.At(x, y).RGBA()
			if ca == 0 {
				continue
			}

			brightness := (float32(cr) + float32(cg) + float32(cb)) / (3 * float32(ca))

			result.Set(x, y, color.NRGBA{
				R: uint8(float32(tr>>8) * brightness),
				G: uint8(float32(tg>>8) * brightness),
				B: uint8(float32(tb>>8) * brightness),
				A: uint8(ca >> 8), //nolint:gosec
			})
		}
	}

	return result
}
