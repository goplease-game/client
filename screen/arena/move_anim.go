package arena

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

const moveDuration = 30 // frames

// moveAnim holds the state of an in-progress unit movement animation.
// While active, the unit icon is drawn as a floating overlay in Draw,
// and the board cells are kept empty until onDone is called.
type moveAnim struct {
	img     *ebiten.Image // unit icon to draw
	fromPx  image.Point   // pixel centre of the source cell
	toPx    image.Point   // pixel centre of the destination cell
	tick    int           // frames elapsed since the animation started
	useLift bool          // true for horizontal moves; applies a lift-travel-land arc
	onDone  func()        // called once when the animation finishes
}

// newMoveAnim creates a moveAnim and decides whether to use the lift arc.
// Horizontal moves (same row, to.Y == from.Y) get the arc;
// vertical and diagonal moves use a simple ease-in-out.
func newMoveAnim(img *ebiten.Image, from, to image.Point, onDone func()) *moveAnim {
	return &moveAnim{
		img:     img,
		fromPx:  from,
		toPx:    to,
		useLift: to.Y == from.Y,
		onDone:  onDone,
	}
}

// active reports whether the animation is still in progress.
// Safe to call on a nil receiver.
func (a *moveAnim) active() bool {
	return a != nil && a.tick <= moveDuration
}

// update advances the animation by one frame and fires onDone when complete.
func (a *moveAnim) update() {
	if !a.active() {
		return
	}
	a.tick++
	if a.tick > moveDuration {
		a.onDone()
	}
}

const liftPx = 20.0 // pixels the unit rises above the source cell during the arc

// Motion phase boundaries as fractions of total duration.
const (
	liftEnd   = 0.15 // 0.00–0.15: rise straight up above source cell
	travelEnd = 0.85 // 0.15–0.85: move horizontally at full lift height
	// landEnd = 1.00 // 0.85–1.00: descend onto destination cell
)

// currentPos returns the interpolated pixel position (top-left of the icon)
// for the current animation frame.
//
// When useLift is true the motion has three phases:
//  1. Lift   (0 → liftEnd):         rise above source cell, no X movement.
//  2. Travel (liftEnd → travelEnd): move to destination at constant height.
//  3. Land   (travelEnd → 1):       descend onto destination, no X movement.
//
// When useLift is false (diagonal/vertical move), simple ease-in-out on both axes.
func (a *moveAnim) currentPos() (x, y float64) {
	t := float64(a.tick) / float64(moveDuration)

	fx, fy := float64(a.fromPx.X), float64(a.fromPx.Y)
	tx, ty := float64(a.toPx.X), float64(a.toPx.Y)

	var cx, cy float64

	if !a.useLift {
		e := easeInOut(t)
		cx = fx + (tx-fx)*e
		cy = fy + (ty-fy)*e
	} else {
		switch {
		case t < liftEnd:
			tPhase := t / liftEnd
			cx = fx
			cy = fy - liftPx*easeInOut(tPhase)

		case t < travelEnd:
			tPhase := (t - liftEnd) / (travelEnd - liftEnd)
			cx = fx + (tx-fx)*easeInOut(tPhase)
			cy = fy - liftPx

		default:
			tPhase := (t - travelEnd) / (1 - travelEnd)
			cx = tx
			cy = ty - liftPx*(1-easeInOut(tPhase))
		}
	}

	hw := float64(a.img.Bounds().Dx()) / 2
	hh := float64(a.img.Bounds().Dy()) / 2
	return cx - hw, cy - hh
}

// easeInOut applies a smooth cubic ease-in-out curve to t in [0, 1].
func easeInOut(t float64) float64 {
	return t * t * (3 - 2*t)
}

// cellCentrePx returns the screen pixel centre of the hex cell at coord.
// Returns the zero point if the cell does not exist in boardCellWidgets.
func (s *Screen) cellCentrePx(coord ds.HexCoord) image.Point {
	w := s.boardCellWidgets[coord]
	if w == nil {
		return image.Point{}
	}
	rect := w.GetWidget().Rect
	return image.Point{
		X: (rect.Min.X + rect.Max.X) / 2,
		Y: (rect.Min.Y + rect.Max.Y) / 2,
	}
}
