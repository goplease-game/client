package arena

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
)

// floatingText holds the state of a floating text animation above a unit.
type floatingText struct {
	text     string
	color    color.Color
	pos      image.Point // starting pixel position (unit centre)
	tick     int
	duration int // total duration in frames
}

// showFloatingText displays an animated floating text above the unit at coord.
// Positive values show in green, negative in red.
func (s *Screen) showFloatingText(coord ds.HexCoord, txt string, col color.Color) {
	w := s.boardCellWidgets[coord]
	if w == nil {
		return
	}

	centre := s.cellCentrePx(coord)

	// need a little random offset,
	// so text on nearby unit will not overlap
	n := rand.Intn(20) + 10

	s.floatingTexts = append(s.floatingTexts, &floatingText{
		text:     txt,
		color:    col,
		pos:      image.Point{X: centre.X, Y: centre.Y - ui.HexRadius - n},
		tick:     0,
		duration: int(1.2 * 60),
	})
}

func (s *Screen) showFloatingStat(coord ds.HexCoord, val int, labelOpt ...string) {
	col := decreasedStatValueColor
	txt := fmt.Sprintf("%d", val)
	if val > 0 {
		col = increasedStatValueColor
		txt = "+" + txt
	}

	if len(labelOpt) > 0 {
		txt += " " + labelOpt[0]
	}

	s.showFloatingText(coord, txt, col)
}

// updateFloatingTexts advances all floating text animations.
func (s *Screen) updateFloatingTexts() {
	alive := s.floatingTexts[:0]
	for _, ft := range s.floatingTexts {
		ft.tick++
		if ft.tick < ft.duration {
			alive = append(alive, ft)
		}
	}

	s.floatingTexts = alive
}

// drawFloatingTexts renders all active floating text animations onto screen.
func (s *Screen) drawFloatingTexts(screen *ebiten.Image) {
	if len(s.floatingTexts) == 0 {
		return
	}

	tf := ui.TextFace(24)
	for _, ft := range s.floatingTexts {
		t := float64(ft.tick) / float64(ft.duration)

		// Fade out in the last 30% of the animation.
		alpha := float32(1.0)
		if t > 0.7 {
			alpha = float32(1.0 - (t-0.7)/0.3)
		}

		// Float upward — moves up by 40px over the full duration.
		offsetY := -40.0 * t

		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(ft.pos.X), float64(ft.pos.Y)+offsetY)
		op.ColorScale.ScaleWithColor(ft.color)
		op.ColorScale.ScaleAlpha(alpha)

		// Measure to centre horizontally.
		w, _ := text.Measure(ft.text, tf, 0)
		op.GeoM.Translate(-w/2, 0)

		text.Draw(screen, ft.text, tf, op)
	}
}
