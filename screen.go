package game

import "github.com/hajimehoshi/ebiten/v2"

// Screen represents one logical screen in the game.
// The main Game delegates Update/Draw to the active screen.
type Screen interface {

	// Update is called every tick (TPS = 60 by default).
	// Returns the next Screen to display, or itself to stay on the same screen.
	Update(g *Game) (Screen, error)

	// Draw renders the screen onto the provided image.
	Draw(screen *ebiten.Image)
}
