package arena

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

const moveDuration = 30 // frames

// moveAnim holds the state of an in-progress unit movement animation.
// While active, the unit icon is drawn as a floating overlay in Draw(),
// and the board cells are kept empty until the animation completes.
type moveAnim struct {
	img     *ebiten.Image // unit icon to draw
	fromPx  image.Point   // pixel centre of the source cell
	toPx    image.Point   // pixel centre of the destination cell
	tick    int           // frames elapsed
	useLift bool          // whether to apply the lift-travel-land arc
	onDone  func()        // called once when the animation finishes
}

// newMoveAnim creates a moveAnim, automatically deciding whether to use the
// lift arc based on the direction of movement:
//   - horizontal (same row): lift applied.
//   - vertical/diagonal (different rows): no lift (simple ease-in-out).
func newMoveAnim(img *ebiten.Image, from, to image.Point, onDone func()) *moveAnim {
	// Изменено: lift применяется ТОЛЬКО если Y-координаты совпадают (одна строка)
	useLift := to.Y == from.Y
	return &moveAnim{
		img:     img,
		fromPx:  from,
		toPx:    to,
		useLift: useLift,
		onDone:  onDone,
	}
}

// active reports whether the animation is still running.
func (a *moveAnim) active() bool {
	return a != nil && a.tick <= moveDuration
}

// update advances the animation by one frame and calls onDone when finished.
func (a *moveAnim) update() {
	if !a.active() {
		return
	}
	a.tick++
	if a.tick > moveDuration {
		a.onDone()
	}
}

const liftPx = 20.0 // how many pixels the unit rises above the cell

// Motion phases (fractions of total duration):
const (
	liftEnd   = 0.15 // 0.00 → 0.15 : rise straight up
	travelEnd = 0.85 // 0.15 → 0.85 : move horizontally at full height
	// landEnd  = 1.00 // 0.85 → 1.00 : descend straight down
)

// currentPos returns the interpolated pixel position (top-left of the icon)
// for the current frame.
//
// When useLift is true, motion has three phases:
//  1. Lift   (0 → liftEnd):         rise above source cell, no X movement.
//  2. Travel (liftEnd → travelEnd): move to destination at constant height.
//  3. Land   (travelEnd → 1):       descend onto destination, no X movement.
//
// When useLift is false (moving downward), simple ease-in-out across both axes.
func (a *moveAnim) currentPos() (x, y float64) {
	t := float64(a.tick) / float64(moveDuration)

	fx, fy := float64(a.fromPx.X), float64(a.fromPx.Y)
	tx, ty := float64(a.toPx.X), float64(a.toPx.Y)

	var cx, cy float64

	if !a.useLift {
		// Simple diagonal ease for downward moves.
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

// easeInOut is a smooth cubic ease-in-out curve.
func easeInOut(t float64) float64 {
	return t * t * (3 - 2*t)
}

// cellCentrePx returns the pixel centre of the board cell widget at (r, c).
func (s *Screen) cellCentrePx(r, c int) image.Point {
	w := s.boardCellWidgets[r][c]
	rect := w.GetWidget().Rect
	return image.Point{
		X: (rect.Min.X + rect.Max.X) / 2,
		Y: (rect.Min.Y + rect.Max.Y) / 2,
	}
}
