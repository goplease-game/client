package ui

import (
	"image"
	"image/color"
	"math"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

const HexRadius = 37 // √3 * 37 ≈ 64px

var whiteImage = ebiten.NewImage(1, 1)

func init() {
	whiteImage.Fill(color.White)
}

type HexChild interface {
	GetHexCoord() ds.HexCoord
	GetWidget() *widget.Widget
}

type HexLayout struct {
	HexSize float64
}

func (l *HexLayout) Layout(widgets []widget.PreferredSizeLocateableWidget, rect image.Rectangle) {
	if l.HexSize == 0 {
		l.HexSize = HexRadius
	}

	hexW := int(math.Round(math.Sqrt(3) * l.HexSize))
	hexH := int(math.Round(2 * l.HexSize))

	for _, w := range widgets {
		hc, ok := w.(HexChild)
		if !ok {
			continue
		}

		coord := hc.GetHexCoord()
		x, y := axialToPixel(coord, l.HexSize)

		x += float64(rect.Min.X)
		y += float64(rect.Min.Y)

		rx := int(math.Round(x))
		ry := int(math.Round(y))

		r := image.Rect(rx, ry, rx+hexW, ry+hexH)

		w.GetWidget().SetLocation(r)

		if h, ok := w.(interface{ SetLocation(image.Rectangle) }); ok {
			h.SetLocation(r)
		}
	}
}

func (l *HexLayout) PreferredSize(widgets []widget.PreferredSizeLocateableWidget) (int, int) {
	if l.HexSize == 0 {
		l.HexSize = HexRadius
	}

	maxX := 0.0
	maxY := 0.0

	for _, w := range widgets {

		hc, ok := w.(HexChild)
		if !ok {
			continue
		}

		coord := hc.GetHexCoord()

		x, y := axialToPixel(coord, l.HexSize)

		if x > maxX {
			maxX = x
		}

		if y > maxY {
			maxY = y
		}
	}

	width := int(math.Round(maxX) + math.Sqrt(3)*l.HexSize)
	height := int(math.Round(maxY) + 2*l.HexSize)

	return width, height
}

func axialToPixel(h ds.HexCoord, size float64) (float64, float64) {
	hexW := int(math.Round(math.Sqrt(3) * size))
	hexH := int(math.Round(2 * size))

	x := float64(h.Q)*float64(hexW) + float64(h.R)*float64(hexW)/2
	y := float64(h.R) * float64(hexH) * 3 / 4

	return x, y
}

func HexImage(size int, clr color.Color) *ebiten.Image {
	w := int(math.Sqrt(3) * float64(size))
	h := size * 2

	dst := ebiten.NewImage(w, h)
	dst.Fill(color.RGBA{0, 0, 0, 0})

	cx := float32(w) / 2
	cy := float32(h) / 2
	r := float32(size)

	var path vector.Path

	for i := 0; i < 6; i++ {
		angle := float64(i)*math.Pi/3 - math.Pi/2

		x := cx + r*float32(math.Cos(angle))
		y := cy + r*float32(math.Sin(angle))

		if i == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
	}

	path.Close()

	// ВАЖНО: FillOptions не содержит цвета
	vector.FillPath(
		dst,
		&path,
		&vector.FillOptions{},
		&vector.DrawPathOptions{},
	)

	// цвет делаем через overlay
	overlay := ebiten.NewImage(w, h)
	overlay.Fill(clr)

	out := ebiten.NewImage(w, h)
	out.DrawImage(dst, nil)

	op := &ebiten.DrawImageOptions{}
	op.Blend = ebiten.BlendSourceOver
	out.DrawImage(overlay, op)

	return out
}
