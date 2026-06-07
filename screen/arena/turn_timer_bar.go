package arena

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
)

type timerBarState struct {
	tick     int
	duration int // total duration in frames
}

// startTurnTimer starts the turn timer for the active unit.
func (s *Screen) startTurnTimer() {
	if s.turnTimeSeconds <= 0 {
		return
	}
	s.timerBar = &timerBarState{
		tick:     0,
		duration: s.turnTimeSeconds * 60,
	}
}

// stopTurnTimer stops the turn timer.
func (s *Screen) stopTurnTimer() {
	s.timerBar = nil
}

// updateTurnTimer advances the turn timer and ends the turn when time runs out.
func (s *Screen) updateTurnTimer() {
	if s.timerBar == nil {
		return
	}
	s.timerBar.tick++
	if s.timerBar.tick >= s.timerBar.duration {
		s.timerBar = nil
		s.server.Send(ws.OutMessage{Action: ws.EndTurnAction})
	}
}

// drawTurnTimer renders the turn timer progress bar above the status bar.
func (s *Screen) drawTurnTimer(screen *ebiten.Image) {
	if s.timerBar == nil {
		return
	}

	t := 1.0 - float64(s.timerBar.tick)/float64(s.timerBar.duration)

	sw := float64(screen.Bounds().Dx())
	sh := float64(screen.Bounds().Dy())

	barH := 3.0
	barY := sh - float64(footerH) - float64(statusH) - barH

	// Background.
	bgImg := ebiten.NewImage(int(sw), int(barH))
	bgImg.Fill(color.NRGBA{0x00, 0x00, 0x00, 0x88})
	screen.DrawImage(bgImg, &ebiten.DrawImageOptions{
		GeoM: func() ebiten.GeoM {
			var m ebiten.GeoM
			m.Translate(0, barY)
			return m
		}(),
	})

	// Progress — color shifts from green to red as time runs out.
	// alpha goes from 0 to 255 as time runs out (t goes from 1.0 to 0.0)
	alpha := uint8(255 * (1 - t))
	r := uint8(255 * (1 - t))
	g := uint8(255 * t)

	barW := int(sw * t)
	if barW > 0 {
		fgImg := ebiten.NewImage(barW, int(barH))
		fgImg.Fill(color.NRGBA{r, g, 0x00, alpha})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(sw-float64(barW), barY)
		screen.DrawImage(fgImg, op)
	}
}
