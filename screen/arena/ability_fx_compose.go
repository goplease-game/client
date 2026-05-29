package arena

import (
	"log"
	"time"

	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/sfx"
)

// abilityComposerRegistry maps ability IDs to custom fx composers.
// Use for abilities that require complex multi-step visual sequences
// that cannot be expressed with a simple Start/End AbilityFx declaration.
var abilityComposerRegistry = map[ability.ID]AbilityFxComposer{
	ability.ShadowStep:    playShadowStepFx, // TODO fade-in-out
	ability.Translocation: playTranslocationFx,
}

type AbilityFxComposer func(s *Screen, unit *ds.Unit, target ds.HexCoord, onDone func())

// abilityFxComposer plays the Start and End fx of an AbilityFx
// according to its PlayMode. Called by playAbilityFx for abilities
// declared in abilityFxRegistry.
func (s *Screen) abilityFxComposer(ab AbilityFx, unit *ds.Unit, target ds.HexCoord, onDone func()) {
	casterCtx := FxContext{Px: s.cellCentrePx(unit.Pos), Coord: unit.Pos}
	targetCtx := FxContext{Px: s.cellCentrePx(target), Coord: target}

	if ab.Start == fxNone {
		if ab.End != fxNone {
			s.playFxAt(ab.End, targetCtx, onDone)
		} else {
			onDone()
		}
		return
	}

	if ab.End == fxNone {
		s.playFxAt(ab.Start, casterCtx, onDone)
		return
	}

	switch ab.PlayMode {
	case FxSequential:
		s.playFxAt(ab.Start, casterCtx, func() {
			s.playFxAt(ab.End, targetCtx, onDone)
		})

	case FxParallel:
		remaining := 2
		done := func() {
			remaining--
			if remaining == 0 {
				onDone()
			}
		}
		s.playFxAt(ab.Start, casterCtx, done)
		s.playFxAt(ab.End, targetCtx, done)

	case FxDelayed:
		remaining := 2
		done := func() {
			remaining--
			if remaining == 0 {
				onDone()
			}
		}
		s.playFxAt(ab.Start, casterCtx, done)
		s.scheduleDelayed(ab.Delay, func() {
			s.playFxAt(ab.End, targetCtx, done)
		})
	}
}

func playShadowStepFx(s *Screen, unit *ds.Unit, target ds.HexCoord, onDone func()) {
	s.hideUnitOnBoard(unit)
	s.playFxAt(fxTeleportStart, FxContext{
		Px:    s.cellCentrePx(unit.Pos),
		Coord: unit.Pos,
	}, func() {
		s.moveUnit(unit, target)
		time.Sleep(1 * time.Second)
		s.showUnitOnBoard(unit)
		s.playFxAt(fxTeleportEnd, FxContext{
			Px:    s.cellCentrePx(target),
			Coord: target,
		}, onDone)
	})
}

func playTranslocationFx(s *Screen, unit *ds.Unit, target ds.HexCoord, onDone func()) {
	sfx.Play("translocation.ogg")

	opp := s.unitAtCoord(target)
	if opp == nil {
		log.Fatalf("invalid target for translocation, no unit at %s", target)
	}

	from := unit.Pos
	to := opp.Pos

	if w := s.boardCellWidgets[from]; w != nil {
		w.RemoveChildren()
	}
	if w := s.boardCellWidgets[to]; w != nil {
		w.RemoveChildren()
	}

	casterAnim := s.moveUnitAnim(unit, opp.Pos)
	targetAnim := s.moveUnitAnim(opp, unit.Pos)

	s.addMoveAnim(casterAnim, targetAnim)
}

// fxFadeOut gradually hides the unit at the given coord (t: 0=visible, 1=hidden).
func fxFadeOut(ctx ProgramFxContext) {
	if ctx.Widget == nil {
		return
	}
	//ctx.Widget.SetUnitFade(uint8(255 * ctx.T))
}

// fxFadeIn gradually reveals the unit at the given coord (t: 0=hidden, 1=visible).
func fxFadeIn(ctx ProgramFxContext) {
	if ctx.Widget == nil {
		return
	}
	//ctx.Widget.SetUnitFade(uint8(255 * (1.0 - ctx.T)))
}
