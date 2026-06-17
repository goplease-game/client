package arena

import (
	"fmt"

	"github.com/goplease-game/client/sfx"
	"github.com/goplease-game/client/ui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/colornames"
)

type newRoundBannerState struct {
	round    int
	tick     int
	duration int // total duration in frames
}

// showNewRoundBanner displays an animated "ROUND X" overlay centered on screen.
// The banner fades in and zooms in, holds, then fades out.
func (s *Screen) showNewRoundBanner(round int) {
	sfx.Play("new_round.ogg")
	s.roundBanner = &newRoundBannerState{
		round:    round,
		tick:     0,
		duration: 2 * 60, // 1 seconds
	}
}

// updateNewRoundBanner advances the round banner animation.
func (s *Screen) updateNewRoundBanner() {
	if s.roundBanner == nil {
		return
	}
	s.roundBanner.tick++
	if s.roundBanner.tick >= s.roundBanner.duration {
		s.roundBanner = nil
	}
}

// drawRoundBanner renders the round announcement overlay onto screen.
// Uses zoomIn + fadeIn for the first 0.5s, holds, then fadeOut for the last 0.5s.
func (s *Screen) drawRoundBanner(screen *ebiten.Image) {
	b := s.roundBanner
	if b == nil {
		return
	}

	t := float64(b.tick) / float64(b.duration)
	fadeInEnd := 0.2
	fadeOutStart := 0.8

	var alpha, scale float64
	switch {
	case t < fadeInEnd:
		p := t / fadeInEnd
		alpha = p
		scale = 0.5 + 0.5*p
	case t > fadeOutStart:
		p := (t - fadeOutStart) / (1 - fadeOutStart)
		alpha = 1 - p
		scale = 1.0
	default:
		alpha = 1.0
		scale = 1.0
	}

	tf := ui.TextFaceBold(120)
	label := fmt.Sprintf("ROUND %d", b.round)

	// Measure text bounds.
	w, h := text.Measure(label, tf, 0)

	sw, sh := float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy())

	op := &text.DrawOptions{}
	op.GeoM.Translate(-w/2, -h/2)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(sw/2, sh/2)
	op.ColorScale.ScaleAlpha(float32(alpha))
	op.ColorScale.ScaleWithColor(colornames.Gold)

	text.Draw(screen, label, tf, op)
}
