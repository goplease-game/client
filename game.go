// Package game ...
package game

import (
	"github.com/goplease-game/client/config"
	"github.com/hajimehoshi/ebiten/v2"
)

// Game is the root ebiten.Game implementation.
// It delegates Update/Draw to the currently active Screen.
type Game struct {
	screen     Screen
	lastScreen Screen

	lastW, lastH int
}

// New creates a new Game instance with the provided initial screen.
func New(s Screen) *Game {
	return &Game{screen: s}
}

// SwitchTo replaces the currently active screen with the provided one.
func (g *Game) SwitchTo(s Screen) { g.screen = s }

// Update delegates to the active screen and handles screen transitions.
func (g *Game) Update() error {
	next, err := g.screen.Update(g)
	if err != nil {
		return err
	}
	if next != g.screen {
		g.screen = next
	}
	return nil
}

// Draw delegates to the active screen.
func (g *Game) Draw(screen *ebiten.Image) {
	g.screen.Draw(screen)
}

// Layout returns the logical screen dimensions.
func (g *Game) Layout(w, h int) (int, int) {
	conf := config.Get()

	if ebiten.IsFullscreen() {
		if m := ebiten.Monitor(); m != nil {
			w, h = m.Size()
		}
	} else {
		if w < conf.WindowW {
			w = conf.WindowW
		}
		if h < conf.WindowH {
			h = conf.WindowH
		}
	}

	if w != g.lastW || h != g.lastH || g.screen != g.lastScreen {
		if r, ok := g.screen.(Resizable); ok {
			r.Resize(w, h)
		}
		g.lastW, g.lastH = w, h
		g.lastScreen = g.screen
	}

	return w, h
}
