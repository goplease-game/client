// Package game ...
package game

import (
	"github.com/google/uuid"
	"github.com/goplease-game/client/config"
	"github.com/goplease-game/client/ws"
	"github.com/hajimehoshi/ebiten/v2"
)

// Game is the root ebiten.Game implementation.
// It owns shared resources (Server connection, player identity) and delegates
// Update/Draw to the currently active Screen.
type Game struct {
	screen   Screen // active screen
	Server   ws.Client
	PlayerID string // stable UUID for this client session
}

// New creates and initializes a new Game instance with a generated player ID,
// the provided server client, and an initial screen.
func New(server ws.Client, s Screen) *Game {
	g := &Game{
		PlayerID: uuid.NewString(),
		Server:   server,
	}

	g.screen = s
	return g
}

// SwitchTo replaces the currently active screen with the provided one.
func (g *Game) SwitchTo(s Screen) { g.screen = s }

// Update updates the game state by delegating the logic to the active screen
// and handles transitions to the next screen if necessary.
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

// Draw renders the game graphics by delegating the drawing operations
// to the currently active screen.
func (g *Game) Draw(screen *ebiten.Image) {
	g.screen.Draw(screen)
}

// Layout accepts the outside window dimensions and returns the logical
// game screen dimensions retrieved from the configuration.
func (g *Game) Layout(_, _ int) (int, int) {
	conf := config.Get()
	return conf.WindowW, conf.WindowH
}
