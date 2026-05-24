package arena

import (
	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

// highlightAbilityRange is called on hover over an ability card.
// It clears any movement selection, then tints cells within the ability's range:
// empty cells get the range tint, valid targets get the target tint.
// Passive abilities have no targeting and are skipped.
func (s *Screen) highlightAbilityRange(ab ability.Ability) {
	if ab.IsPassive {
		return
	}

	// Clear movement highlight — two simultaneous highlights would be confusing.
	s.deselectUnit()

	caster, ok := s.unitByID(s.activeUnitID)
	if !ok {
		return
	}

	cells := cellsInRange(caster.Pos, ab.Range, s.board)
	s.abilityHighlightCells = cells

	for _, pos := range cells {
		w := s.boardCellWidgets[pos]
		if w == nil {
			continue
		}

		cell := s.board.Cells[pos]

		switch {
		case cell == nil || cell.Unit == nil:
			w.SetColor(abilityRangeCellColor)
		case s.isValidTarget(ab, caster, *cell.Unit):
			w.SetColor(abilityTargetCellColor)
		}
	}
}

// clearAbilityHighlight restores all ability-highlighted cells to their default colours.
// Called when the cursor leaves an ability card.
func (s *Screen) clearAbilityHighlight() {
	for _, pos := range s.abilityHighlightCells {
		w := s.boardCellWidgets[pos]
		if w == nil {
			continue
		}

		cell := s.board.Cells[pos]

		bg := boardCellBgColor
		if cell != nil && cell.Unit != nil {
			if cell.Unit.IsOpponent {
				bg = unitEnemyBgColor
			} else {
				bg = unitFriendlyBgColor
			}
		}

		w.SetColor(bg)
	}

	s.abilityHighlightCells = nil
}

// isValidTarget reports whether target is a valid target for ab cast by caster,
// based on the ability's TargetMode.
func (s *Screen) isValidTarget(ab ability.Ability, caster ds.Unit, target ds.Unit) bool {
	switch ab.TargetMode {
	case ability.TargetEnemies:
		return target.IsOpponent != caster.IsOpponent
	case ability.TargetAllies:
		return target.IsOpponent == caster.IsOpponent && target.ID != caster.ID
	case ability.TargetAlliesAndSelf:
		return target.IsOpponent == caster.IsOpponent
	case ability.TargetAny:
		return true
	default:
		return false
	}
}
