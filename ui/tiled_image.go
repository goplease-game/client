package ui

import (
	"image"
	"image/color"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// TiledBackground is a widget that fills its rect with a solid base
// color, then draws src tiled on top as a repeating pattern. Its own
// PreferredSize mirrors sizeSource's, so an AnchorLayout container
// using it as its first child reports the size of the actual content
// stacked behind it.
type TiledBackground struct {
	src        *ebiten.Image
	baseColor  color.Color
	sizeSource widget.PreferredSizer
	widget     *widget.Widget
}

// NewTiledBackground creates a TiledBackground filling its rect with
// baseColor and drawing src tiled on top. sizeSource's PreferredSize is
// reported as this widget's own, so it should be the sibling widget
// whose size the background must match.
func NewTiledBackground(src *ebiten.Image, baseColor color.Color, sizeSource widget.PreferredSizer, opts ...widget.WidgetOpt) *TiledBackground {
	return &TiledBackground{
		src:        src,
		baseColor:  baseColor,
		sizeSource: sizeSource,
		widget:     widget.NewWidget(opts...),
	}
}

// GetWidget implements widget.HasWidget.
func (t *TiledBackground) GetWidget() *widget.Widget {
	return t.widget
}

// PreferredSize implements widget.PreferredSizer.
func (t *TiledBackground) PreferredSize() (int, int) {
	if t.sizeSource == nil {
		return 0, 0
	}
	return t.sizeSource.PreferredSize()
}

// SetLocation implements widget.PreferredSizeLocateableWidget.
func (t *TiledBackground) SetLocation(rect image.Rectangle) {
	t.widget.Rect = rect
}

// Validate implements widget.PreferredSizeLocateableWidget.
func (t *TiledBackground) Validate() {}

// Render implements widget.Renderer.
func (t *TiledBackground) Render(screen *ebiten.Image) {
	t.widget.Render(screen)

	rect := t.widget.Rect
	w, h := rect.Dx(), rect.Dy()
	if w <= 0 || h <= 0 {
		return
	}

	if t.baseColor != nil {
		vector.FillRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(w), float32(h), t.baseColor, false)
	}

	sw, sh := t.src.Bounds().Dx(), t.src.Bounds().Dy()
	if sw <= 0 || sh <= 0 {
		return
	}

	clip := screen.SubImage(rect).(*ebiten.Image) //nolint:forcetypeassert
	for y := 0; y < h; y += sh {
		for x := 0; x < w; x += sw {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(rect.Min.X+x), float64(rect.Min.Y+y))
			clip.DrawImage(t.src, op)
		}
	}
}
