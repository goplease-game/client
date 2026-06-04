package arena

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/hex"
	"github.com/ognev-dev/goplease-ebitengine-client/sfx"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
)

// abilityComposerRegistry maps ability IDs to custom fx composers.
// Use for abilities that require complex multi-step visual sequences
// that cannot be expressed with a simple Start/End AbilityFx declaration.
var abilityComposerRegistry = map[ability.ID]AbilityFxComposer{
	ability.ShadowStep:    playShadowStepFx,
	ability.Translocation: playTranslocationFx,
}

type AbilityFxComposer func(s *Screen, unit *ds.Unit, target ds.HexCoord, onDone func())

// abilityFxComposer plays the Start and End fx of an AbilityFx
// according to its PlayMode. Called by playAbilityFx for abilities
// declared in abilityFxRegistry.
func (s *Screen) abilityFxComposer(abFx AbilityFx, abilityID ability.ID, unit *ds.Unit, target ds.HexCoord, onDone func()) {
	casterCtx := FxContext{Px: s.cellCentrePx(unit.Pos), Coord: unit.Pos}

	// Determine End targets — AOE plays on all valid targets, single target plays on target.
	abDef := ability.ByID(abilityID)
	isAOE := abDef.Area != ""

	// playEnd plays the End fx on all valid targets (AOE) or single target.
	// Calls afterDone when all End fx have finished.
	playEnd := func(afterDone func()) {
		if abFx.End == fxNone {
			afterDone()
			return
		}

		if isAOE {
			targets := s.abilityTargets(abDef, unit)
			if len(targets) == 0 {
				afterDone()
				return
			}
			remaining := len(targets)
			done := func() {
				remaining--
				if remaining == 0 {
					afterDone()
				}
			}
			for _, t := range targets {
				s.playFxAt(abFx.End, FxContext{Px: s.cellCentrePx(t), Coord: t}, done)
			}
		} else {
			s.playFxAt(abFx.End, FxContext{Px: s.cellCentrePx(target), Coord: target}, afterDone)
		}
	}

	if abFx.Start == fxNone {
		playEnd(onDone)
		return
	}

	switch abFx.PlayMode {
	case FxSequential:
		s.playFxAt(abFx.Start, casterCtx, func() {
			playEnd(onDone)
		})

	case FxParallel:
		startFinished := false
		endFinished := false
		checkDone := func() {
			if startFinished && endFinished {
				onDone()
			}
		}
		s.playFxAt(abFx.Start, casterCtx, func() {
			startFinished = true
			checkDone()
		})
		playEnd(func() {
			endFinished = true
			checkDone()
		})

	case FxDelayed:
		startFinished := false
		endFinished := false
		checkDone := func() {
			if startFinished && endFinished {
				onDone()
			}
		}
		s.playFxAt(abFx.Start, casterCtx, func() {
			startFinished = true
			checkDone()
		})
		s.scheduleDelayed(abFx.Delay, func() {
			playEnd(func() {
				endFinished = true
				checkDone()
			})
		})
	}
}

func playShadowStepFx(s *Screen, unit *ds.Unit, target ds.HexCoord, onDone func()) {
	unitImg := unitImage(unit.TemplateID, unitIconSize)

	s.clearAbilityHighlight()
	s.setPulseHexTargets(nil)
	s.hideUnitOnBoard(unit)
	sfx.Play("teleport_out.ogg")

	s.activeFxAnims = append(s.activeFxAnims, &ActiveFxAnim{
		pos:             s.cellCentrePx(unit.Pos),
		coord:           unit.Pos,
		programFx:       fxUnitFadeZoomOut(unitImg),
		programDuration: int(0.5 * 60),
		onDone: func() {
			s.moveUnit(unit, target)
			unit.Pos = target
			sfx.Play("teleport_in.ogg")

			s.activeFxAnims = append(s.activeFxAnims, &ActiveFxAnim{
				pos:             s.cellCentrePx(target),
				coord:           target,
				programFx:       fxUnitFadeZoomIn(unitImg),
				programDuration: int(0.5 * 60),
				onDone: func() {
					s.showUnitOnBoard(unit)
					if w := s.boardCellWidgets[unit.Pos]; w != nil {
						s.setPulseHexTargets([]*ui.HexCellWidget{w})
					}
					onDone()
				},
			})
		},
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

// fxUnitFadeZoomOut animates the unit icon fading out and shrinking.
// Draws the icon directly onto screen each frame.
func fxUnitFadeZoomOut(unitImg *ebiten.Image) ProgramFx {
	return func(ctx ProgramFxContext) {
		if unitImg == nil {
			return
		}
		t := ctx.T
		alpha := float32(1.0 - t)
		scale := 1.0 - 0.5*t

		w := float64(unitImg.Bounds().Dx())
		h := float64(unitImg.Bounds().Dy())

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-w/2, -h/2)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(float64(ctx.Px.X), float64(ctx.Px.Y))
		op.ColorScale.ScaleAlpha(alpha)

		ctx.DrawTarget.DrawImage(unitImg, op)
	}
}

// fxUnitFadeZoomIn animates the unit icon fading in and growing.
func fxUnitFadeZoomIn(unitImg *ebiten.Image) ProgramFx {
	return func(ctx ProgramFxContext) {
		if unitImg == nil {
			return
		}
		t := ctx.T
		alpha := float32(t)
		scale := 0.5 + 0.5*t

		w := float64(unitImg.Bounds().Dx())
		h := float64(unitImg.Bounds().Dy())

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-w/2, -h/2)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(float64(ctx.Px.X), float64(ctx.Px.Y))
		op.ColorScale.ScaleAlpha(alpha)

		ctx.DrawTarget.DrawImage(unitImg, op)
	}
}

// abilityTargets returns all valid target coords for the given ability cast by unit.
func (s *Screen) abilityTargets(ab ability.Ability, unit *ds.Unit) []ds.HexCoord {
	var cells []ds.HexCoord

	switch ab.Area {
	case ability.AreaCircle:
		cells = hex.CellsInRange(unit.Pos, ab.AreaRadius, s.board)
	case ability.AreaLine:
		cells = hexAllLines(unit.Pos, ab.AreaRadius, s.board)
	default:
		cells = hex.CellsInRange(unit.Pos, ab.Range, s.board)
	}

	var targets []ds.HexCoord
	for _, pos := range cells {
		cell := s.board.Cells[pos]
		if cell != nil && cell.Unit != nil && s.isValidTarget(ab, unit, *cell.Unit) {
			targets = append(targets, pos)
		}
	}
	return targets
}
