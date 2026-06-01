package arena

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/sfx"
)

const moveDuration = 30 // frames

// moveUnitAnim creates a movement action using the unit's current position as the starting point.
// Note: This function only constructs the action object; it does NOT play the animation.
// To play the animation, you must add it to the queue: s.addMoveAnim(s.moveUnitAnim(u, to))
func (s *Screen) moveUnitAnim(u *ds.Unit, to ds.HexCoord) unitMoveAnimAction {
	return unitMoveAnimAction{
		anim:   newMoveAnim(unitImage(u.TemplateID), s.cellCentrePx(u.Pos), s.cellCentrePx(to)),
		unitID: u.ID,
		from:   u.Pos,
		to:     to,
	}
}

type unitMoveAnimAction struct {
	anim   *unitMoveAnim
	unitID string
	from   ds.HexCoord
	to     ds.HexCoord
}

// unitMoveAnim holds the state of an in-progress unit movement animation.
type unitMoveAnim struct {
	img     *ebiten.Image
	fromPx  image.Point
	toPx    image.Point
	tick    int
	useLift bool
}

// newMoveAnim creates a unitMoveAnim and decides whether to use the lift arc.
func newMoveAnim(img *ebiten.Image, fromPx, toPx image.Point) *unitMoveAnim {
	return &unitMoveAnim{
		img:     img,
		fromPx:  fromPx,
		toPx:    toPx,
		tick:    0,
		useLift: fromPx.Y == toPx.Y,
	}
}

// active reports whether the animation is still in progress.
// Safe to call on a nil receiver.
func (a *unitMoveAnim) active() bool {
	return a != nil && a.tick <= moveDuration
}

// update advances the animation frame, capping it at moveDuration + 1
func (a *unitMoveAnim) update() {
	if a.tick <= moveDuration {
		a.tick++
	}
}

// isDone checks if the animation has safely completed its full duration cycle
func (a *unitMoveAnim) isDone() bool {
	return a.tick > moveDuration
}

func (s *Screen) updateMoveAnimations() {
	if len(s.unitMoveAnimQueue) == 0 {
		return
	}

	currentGroup := s.unitMoveAnimQueue[0]
	allDone := true

	if len(currentGroup) > 0 && currentGroup[0].anim.tick == 0 {
		sfx.Play(moveSound)
	}

	for i := range currentGroup {
		if !currentGroup[i].anim.isDone() {
			allDone = false
		}
		currentGroup[i].anim.update()
	}

	if allDone {
		// 1. First loop: ONLY clear the "from" positions for all units in the group.
		// This ensures we don't accidentally wipe out a newly rendered unit on the next iteration.
		for _, action := range currentGroup {
			if fromW := s.boardCellWidgets[action.from]; fromW != nil {
				s.removePulseWidget(fromW)
				s.restoreSafeZoneCell(action.from)
				fromW.SetColor(boardCellBgColor)
				fromW.RemoveChildren() // Safely clear before anyone lands here
			}
		}

		// 2. Second loop: Apply logic updates and render units on their "to" positions.
		for _, action := range currentGroup {
			u := s.unitByID(action.unitID)
			s.moveUnit(u, action.to) // Update unit position logically

			if s.selectedUnitID == u.ID || !u.IsOpponent {
				s.activeUnitMoved = true
				s.updateActiveUnitStatusLabel()
				s.updateNextActionLabel()
			}

			// Render the unit on its new destination cell widget safely
			if toW := s.boardCellWidgets[action.to]; toW != nil {
				targetBg := unitFriendlyBgColor
				if u.IsOpponent {
					targetBg = unitEnemyBgColor
				}
				toW.SetColor(targetBg)
				toW.RemoveChildren() // Clear anything old inside the target cell

				buildBoardCard(toW, u, false) // Draw the unit card

				if !u.IsOpponent {
					s.pulseHexWidgets = append(s.pulseHexWidgets, toW)
				}
			}
		}

		sfx.Play(moveSound)
		s.unitMoveAnimQueue = s.unitMoveAnimQueue[1:]
	}
}

// addMoveAnim schedules a group of animations to be executed simultaneously.
func (s *Screen) addMoveAnim(anims ...unitMoveAnimAction) {
	if len(anims) == 0 {
		return
	}

	group := make([]unitMoveAnimAction, len(anims))
	copy(group, anims)

	s.unitMoveAnimQueue = append(s.unitMoveAnimQueue, group)
}

// moveUnitForced plays the movement animation immediately without using a lift arc.
// This is useful for displaying movement animations received from the server or an opponent.
// If you need scheduled animations (e.g., to support simultaneous animations), see moveUnitAnim and addMoveAnim.
func (s *Screen) moveUnitForced(u *ds.Unit, to ds.HexCoord) {
	act := unitMoveAnimAction{
		anim: &unitMoveAnim{
			img:     unitImage(u.TemplateID),
			fromPx:  s.cellCentrePx(u.Pos),
			toPx:    s.cellCentrePx(to),
			useLift: false,
		},
		unitID: u.ID,
		from:   u.Pos,
		to:     to,
	}

	s.unitMoveAnimQueue = append(s.unitMoveAnimQueue, []unitMoveAnimAction{act})
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
func (a *unitMoveAnim) currentPos() (x, y float64) {
	t := float64(a.tick) / float64(moveDuration)
	if t > 1.0 {
		t = 1.0
	}

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
