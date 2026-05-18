package game

import (
	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ognev-dev/goplease-ebitengine-client/config"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
)

// Game is the root ebiten.Game implementation.
// It owns shared resources (Server connection, player identity) and delegates
// Update/Draw to the currently active Screen.
type Game struct {
	screen   Screen // active screen
	Server   ws.Client
	PlayerID string // stable UUID for this client session
}

func New(server ws.Client, s Screen) *Game {
	g := &Game{
		PlayerID: uuid.NewString(),
		Server:   server,
	}

	g.screen = s
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
	conf := config.Get()
	return conf.WindowW, conf.WindowH
}
