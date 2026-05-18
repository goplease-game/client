package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

func CreateCircleImage(size int, clr color.Color) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	cx, cy, r := float32(size/2), float32(size/2), float32(size/2)-1
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx, dy := float32(x)-cx, float32(y)-cy
			if dx*dx+dy*dy <= r*r {
				img.Set(x, y, clr)
			}
		}
	}

	return img
}

func TintImage(src *ebiten.Image, iconColor color.Color) *ebiten.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	result := ebiten.NewImage(w, h)

	tr, tg, tb, _ := iconColor.RGBA()

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cr, cg, cb, ca := src.At(x, y).RGBA()
			if ca == 0 {
				continue
			}

			brightness := (float32(cr) + float32(cg) + float32(cb)) / (3 * float32(ca))

			result.Set(x, y, color.NRGBA{
				R: uint8(float32(tr>>8) * brightness),
				G: uint8(float32(tg>>8) * brightness),
				B: uint8(float32(tb>>8) * brightness),
				A: uint8(ca >> 8),
			})
		}
	}

	return result
}
