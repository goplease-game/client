package arena

import (
	"fmt"

	"github.com/goplease-game/client/ability"
	"github.com/goplease-game/client/ds"
	"github.com/goplease-game/client/grid"
)

var hexDirections = [6]ds.HexCoord{
	{Q: 1, R: 0}, {Q: -1, R: 0},
	{Q: 0, R: 1}, {Q: 0, R: -1},
	{Q: 1, R: -1}, {Q: -1, R: 1},
}

// highlightAbilityRange is called on hover over an ability card.
// It clears any movement selection, then tints cells within the ability's range:
// empty cells get the range tint, valid targets get the target tint.
// Passive abilities have no targeting and are skipped.
func (s *Screen) highlightAbilityRange(ab ability.Ability) {
	if ab.IsPassive {
		return
	}
	// In targeting mode — don't change highlights.
	if s.selectedAbility != nil {
		return
	}
	s.deselectUnit()

	caster := s.unitByID(s.activeUnitID)
	if caster == nil {
		return
	}

	rangeN := ab.Range
	var cells []ds.HexCoord

	switch ab.Area {
	case ability.AreaCircle:
		cells = grid.CellsInRange(caster.Pos, ab.AreaRadius, s.board)
	case ability.AreaLine:
		cells = hexAllLines(caster.Pos, ab.AreaRadius, s.board)
	default:
		cells = grid.CellsInRange(caster.Pos, rangeN, s.board)
	}

	fmt.Printf("HAB: %s: cells: %d\n", ab.Name, len(cells))

	s.abilityHighlightCells = cells

	for _, pos := range cells {
		w := s.boardCellWidgets[pos]
		if w == nil {
			continue
		}

		cell := s.board.Cells[pos]
		if cell == nil {
			continue
		}
		u := cell.Unit
		if u != nil && u.IsDead {
			continue
		}

		switch {
		case cell.Unit == nil:
			w.SetColor(abilityRangeCellColor)
		case s.isValidAbilityTarget(ab, pos):
			col := abilityAllyTargetCellColor
			if cell.Unit.IsEnemy(caster) {
				col = abilityEnemyTargetCellColor
			}

			w.SetColor(col)
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

// hexLine returns cells in a straight line from `from` in direction `dir`
// up to `length` steps. Only returns cells that exist on the board.
func hexLine(from ds.HexCoord, dir ds.HexCoord, length int, board ds.Board) []ds.HexCoord {
	var result []ds.HexCoord
	cur := from
	for range length {
		cur = ds.HexCoord{Q: cur.Q + dir.Q, R: cur.R + dir.R}
		if _, ok := board.Cells[cur]; !ok {
			// Cell doesn't exist on board — stop this ray.
			break
		}
		result = append(result, cur)
	}
	return result
}

// hexAllLines returns cells in all 6 directions from `from` up to `length` steps.
func hexAllLines(from ds.HexCoord, length int, board ds.Board) []ds.HexCoord {
	var result []ds.HexCoord //nolint:prealloc
	for _, dir := range hexDirections {
		result = append(result, hexLine(from, dir, length, board)...)
	}

	return result
}
