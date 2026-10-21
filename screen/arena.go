package screen

import (
	game "github.com/goplease-game/client"
	"github.com/goplease-game/client/ds"
	"github.com/goplease-game/client/screen/arena"
	"github.com/hajimehoshi/ebiten/v2"
)

// ArenaScreen wraps the arena package's Screen to satisfy the game.Screen interface.
type ArenaScreen struct {
	arena *arena.Screen
}

// NewArenaScreen creates the arena screen for snap, wiring up exit and
// restart transitions depending on whether this is a practice match
// (against the mock client) or a real one against server.
func NewArenaScreen(ctx *game.Context, snap ds.GameSnapshot, isPractice bool) game.Screen {
	ar := arena.NewScreen(snap, ctx.Server.Get())
	ar.OnExitScreen = func() game.Screen {
		return NewMainScreen(ctx)
	}

	if isPractice {
		ar.OnRestartScreen = func() game.Screen {
			return newScenarioScreen(ctx)
		}
	} else {
		ar.OnRestartScreen = func() game.Screen {
			return NewSearchScreen(ctx, NewMainScreen(ctx))
		}
	}

	s := &ArenaScreen{
		arena: ar,
	}

	return s
}

// Update implements game.Screen by delegating to the underlying arena screen.
func (s *ArenaScreen) Update(g *game.Game) (game.Screen, error) {
	return s.arena.Update(g)
}

// Draw implements game.Screen by delegating to the underlying arena screen.
func (s *ArenaScreen) Draw(screen *ebiten.Image) {
	s.arena.Draw(screen)
}
