package arena

import (
	"github.com/ebitenui/ebitenui/image"
	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

// highlightAbilityRange is called on hover over an ability card.
// It clears movement selection, then highlights the ability's range zone
// and valid targets within it.
func (s *Screen) highlightAbilityRange(ab ability.Ability) {
	// Passive abilities have no targeting — nothing to highlight.
	if ab.IsPassive {
		return
	}

	// Clear movement selection first — two highlights at once is noisy.
	s.deselectUnit()

	caster, ok := s.unitByID(s.activeUnitID)
	if !ok {
		return
	}

	from := [2]int{caster.Row, caster.Col}
	cells := cellsInRange(from, ab.Range, s.board)

	s.abilityHighlightCells = cells

	for _, pos := range cells {
		r, c := pos[0], pos[1]
		w := s.boardCellWidgets[r][c]
		if w == nil {
			continue
		}

		cell := s.board[r][c]

		switch {
		case cell == nil || cell.Unit == nil:
			// Empty cell — highlight as range zone.
			w.SetBackgroundImage(image.NewNineSliceColor(abilityRangeCellColor))
		case s.isValidTarget(ab, caster, *cell.Unit):
			// Valid target — highlight as targetable.
			w.SetBackgroundImage(image.NewNineSliceColor(abilityTargetCellColor))
			// Ally — leave untouched.
		}
	}
}

// clearAbilityHighlight restores all ability-highlighted cells to their original colours.
func (s *Screen) clearAbilityHighlight() {
	for _, pos := range s.abilityHighlightCells {
		r, c := pos[0], pos[1]
		w := s.boardCellWidgets[r][c]
		if w == nil {
			continue
		}

		cell := s.board[r][c]
		bg := boardCellBgColor
		if cell != nil && cell.Unit != nil {
			if cell.Unit.IsOpponent {
				bg = unitEnemyBgColor
			} else {
				bg = unitFriendlyBgColor
			}
		}
		w.SetBackgroundImage(image.NewNineSliceColor(bg))
	}

	s.abilityHighlightCells = nil
}

// isValidTarget checks whether `target` is a valid target for `ab`
// according to the ability's TargetMode.
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
