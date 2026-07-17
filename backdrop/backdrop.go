// Package backdrop provides animated background layers (starfield, nebula,
// abstract tree) that can be drawn behind menu or match UI in place of a
// flat dark screen.
package backdrop

import (
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
)

// Backdrop is an animated background layer. Implementations own their own
// state and advance it independently of the game simulation tick.
type Backdrop interface {
	Update()
	Draw(screen *ebiten.Image)

	// Resize updates the layer to a new screen size. Call it whenever the
	// game's Layout reports a different outside size, e.g. on window
	// resize; implementations rescale or reflow their existing state
	// rather than discarding it, so the transition doesn't visibly pop.
	Resize(width, height int)
}

// Kind identifies a Backdrop implementation for explicit selection.
type Kind int

// Kind defines the identifier for a specific type of background effect.
const (
	KindStarfield Kind = iota
	KindNebula
	KindFloatingHexes
	KindHexCrystalField
	KindNebulaWithStars
	KindNebulaWithFloatingHexes
	KindNebulaWithCrystalHexField
)

// MainScreen holds the selection of backdrop kinds available for the main menu or hub screen.
var MainScreen = []Kind{
	// KindNebulaWithFloatingHexes, // disabled for now, maybe can use this somewhere else
	KindNebulaWithCrystalHexField,
}

// ArenaScreen holds the selection of backdrop kinds available for the combat or gameplay arena.
var ArenaScreen = []Kind{
	KindNebulaWithStars,
}

// CombinedBackdrop updates and draws multiple backdrops in order (back-to-front).
type CombinedBackdrop struct {
	layers []Backdrop
}

// Update advances the logical state of all managed background layers.
func (cb *CombinedBackdrop) Update() {
	for _, layer := range cb.layers {
		layer.Update()
	}
}

// Draw renders all managed background layers to the screen in back-to-front order.
func (cb *CombinedBackdrop) Draw(screen *ebiten.Image) {
	for _, layer := range cb.layers {
		layer.Draw(screen)
	}
}

// Resize propagates the new window or screen dimensions to all managed background layers.
func (cb *CombinedBackdrop) Resize(width, height int) {
	for _, layer := range cb.layers {
		layer.Resize(width, height)
	}
}

// New constructs a Backdrop of the given kind sized for width x height.
func New(kind Kind, width, height int) Backdrop {
	if ebiten.IsFullscreen() {
		if m := ebiten.Monitor(); m != nil {
			width, height = m.Size()
		}
	}

	switch kind {
	case KindNebula:
		return NewNebula(width, height)
	case KindFloatingHexes:
		return NewFloatingHexField(width, height)
	case KindHexCrystalField:
		return NewCrystalHexField(width, height)
	case KindNebulaWithStars:
		return &CombinedBackdrop{
			layers: []Backdrop{
				NewNebula(width, height),
				NewStarfield(width, height),
			},
		}
	case KindNebulaWithFloatingHexes:
		return &CombinedBackdrop{
			layers: []Backdrop{
				NewNebula(width, height),
				NewFloatingHexField(width, height),
			},
		}
	case KindNebulaWithCrystalHexField:
		return &CombinedBackdrop{
			layers: []Backdrop{
				NewNebula(width, height),
				NewCrystalHexField(width, height),
			},
		}
	default:
		return NewStarfield(width, height)
	}
}

// RandomOf constructs a Backdrop of a randomly chosen kind sized for
// width x height.
func RandomOf(kinds []Kind, width, height int) Backdrop {
	if len(kinds) == 0 {
		// Fallback to a default kind or handle the error appropriately
		return New(Kind(0), width, height)
	}

	kind := kinds[rand.IntN(len(kinds))]
	return New(kind, width, height)
}
