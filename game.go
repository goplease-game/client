package client

import (
	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenWidth  = 1200
	ScreenHeight = 900
)

// Game is the root ebiten.Game implementation.
// It owns shared resources (Server connection, player identity) and delegates
// Update/Draw to the currently active Screen.
type Game struct {
	screen   Screen // active screen
	Server   *WSClient
	PlayerID string // stable UUID for this client session
}

func NewGame() *Game {
	g := &Game{
		PlayerID: uuid.NewString(),
		Server:   NewWSClient(),
	}
	g.screen = NewMainScreen()
	return g
}

// SwitchTo replaces the active screen.
func (g *Game) SwitchTo(s Screen) { g.screen = s }

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

func (g *Game) Draw(screen *ebiten.Image) {
	g.screen.Draw(screen)
}

func (g *Game) Layout(outsideW, outsideH int) (int, int) {
	return ScreenWidth, ScreenHeight
}
