package arena

import (
	"image"
	"image/color"
	"math/rand/v2"
	"strconv"

	"github.com/goplease-game/client/ds"
	"github.com/goplease-game/client/ui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// floatingText holds the state of a floating text animation above a unit.
type floatingText struct {
	text     string
	color    color.Color
	pos      image.Point // starting pixel position (unit centre)
	tick     int
	duration int         // total duration in frames
	delay    int         // frames to wait before animation starts
	coord    ds.HexCoord // used to stagger multiple texts on same unit
}

// showFloatingText displays an animated floating text above the unit at coord.
func (s *Screen) showFloatingText(coord ds.HexCoord, txt string, col color.Color) {
	w := s.boardCellWidgets[coord]
	if w == nil {
		return
	}

	centre := s.cellCentrePx(coord)
	n := rand.IntN(20) + 10 //nolint:gosec

	// Stagger texts on the same unit — each one waits for the previous to finish.
	delay := 0
	for _, ft := range s.floatingTexts {
		if ft.coord == coord {
			delay += 20 // 20 frames between each text on same unit
		}
	}

	s.floatingTexts = append(s.floatingTexts, &floatingText{
		text:     txt,
		color:    col,
		pos:      image.Point{X: centre.X, Y: centre.Y - ui.HexRadius - n},
		tick:     0,
		duration: int(1.2 * 60),
		delay:    delay,
		coord:    coord,
	})
}

func (s *Screen) showFloatingStat(coord ds.HexCoord, val int, labelOpt ...string) {
	col := decreasedStatValueColor
	txt := strconv.Itoa(val)
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
		if ft.delay > 0 {
			ft.delay--
			alive = append(alive, ft)
			continue
		}
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
		if ft.delay > 0 {
			continue // still waiting
		}

		t := float64(ft.tick) / float64(ft.duration)

		alpha := float32(1.0)
		if t > 0.7 {
			alpha = float32(1.0 - (t-0.7)/0.3)
		}

		offsetY := -40.0 * t

		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(ft.pos.X), float64(ft.pos.Y)+offsetY)
		op.ColorScale.ScaleWithColor(ft.color)
		op.ColorScale.ScaleAlpha(alpha)

		w, _ := text.Measure(ft.text, tf, 0)
		op.GeoM.Translate(-w/2, 0)

		text.Draw(screen, ft.text, tf, op)
	}
}
