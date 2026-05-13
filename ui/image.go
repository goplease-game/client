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
