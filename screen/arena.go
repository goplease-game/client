package screen

import (
	"github.com/hajimehoshi/ebiten/v2"
	game "github.com/ognev-dev/goplease-ebitengine-client"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/screen/arena"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
)

type ArenaScreen struct {
	arena *arena.Screen
}

func NewArenaScreen(snap ds.GameSnapshot, server ws.Client, isPractice bool) game.Screen {
	ar := arena.NewScreen(snap, server)
	ar.OnExitScreen = func() game.Screen {
		return NewMainScreen(server)
	}

	if isPractice {
		ar.OnRestartScreen = func() game.Screen {
			return newPracticeScreen()
		}
	}

	s := &ArenaScreen{
		arena: ar,
	}

	return s
}

func (s *ArenaScreen) OnEnter(_ *game.Game) {}

func (s *ArenaScreen) Update(g *game.Game) (game.Screen, error) {
	return s.arena.Update(g)
}

func (s *ArenaScreen) Draw(screen *ebiten.Image) {
	s.arena.Draw(screen)
}
