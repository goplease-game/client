package arena

import (
	"github.com/goplease-game/client/ds"
	"github.com/goplease-game/client/ui"
	"github.com/goplease-game/client/ws"
)

// selectUnit is called when the player clicks the active unit on the board.
// Tints all reachable cells and stores the selection.
// A second click on the same unit deselects it.
// No-ops if the unit has already moved this turn.
func (s *Screen) selectUnit(u *ds.Unit) {
	if s.activeUnitMoved {
		return
	}
	if s.selectedUnitID == u.ID {
		s.deselectUnit()
		return
	}

	s.selectedUnitID = u.ID
	s.reachableCells = u.ReachableCells(s.board)

	if bc := s.boardCellWidget(u); bc != nil {
		bc.SetColor(unitFriendlyBgColor)
	}

	for _, pos := range s.reachableCells {
		if w := s.boardCellWidgets[pos]; w != nil {
			w.SetColor(unitMoveToCellColor)
		}
	}
}

// deselectUnit clears the current unit selection and restores all
// highlighted cells to their original colours. No-ops if nothing is selected.
func (s *Screen) deselectUnit() {
	if s.selectedUnitID == "" {
		return
	}

	u := s.unitByID(s.selectedUnitID)
	if bc := s.boardCellWidget(u); bc != nil {
		bc.SetColor(unitFriendlyBgColor)
	}

	for _, pos := range s.reachableCells {
		cell := s.board.Cells[pos]

		bg := boardCellBgColor
		if cell != nil && cell.Unit != nil {
			if cell.Unit.IsOpponent {
				bg = unitEnemyBgColor
			} else {
				bg = unitFriendlyBgColor
			}
		}

		if w := s.boardCellWidgets[pos]; w != nil {
			w.SetColor(bg)
		}
	}

	s.selectedUnitID = ""
	s.reachableCells = nil
}

// onReachableCellClicked is called when the player clicks a highlighted reachable cell.
// It starts the movement animation and immediately notifies the server.
func (s *Screen) onReachableCellClicked(to ds.HexCoord) {
	u := s.unitByID(s.selectedUnitID)
	from := ds.HexCoord{Q: u.Pos.Q, R: u.Pos.R}

	s.deselectUnit()

	// Remove the unit icon from the source cell for the duration of the animation.
	if w := s.boardCellWidgets[from]; w != nil {
		w.RemoveChildren()
	}

	s.addMoveAnim(s.moveUnitAnim(u, to))

	// Notify the server immediately — it does not need to wait for the animation.
	s.server.Send(ws.OutMessage{
		Action: ws.UnitMovedAction,
		Data: ds.UnitMovedPayload{
			UnitID: u.ID,
			Coord:  to,
		},
	})
}

// removePulseWidget removes w from the pulse list if present.
func (s *Screen) removePulseWidget(w *ui.HexCellWidget) {
	for i, pw := range s.pulseHexWidgets {
		if pw == w {
			s.pulseHexWidgets = append(s.pulseHexWidgets[:i], s.pulseHexWidgets[i+1:]...)
			return
		}
	}
}

// restoreDropZoneCell resets the safe-zone cell at coord to unoccupied,
// allowing a unit to be placed there again in a future turn.
func (s *Screen) restoreDropZoneCell(coord ds.HexCoord) {
	for _, sc := range s.dropZoneCells {
		if sc.coord == coord {
			sc.occupied = false
			sc.baseColor = boardCellBgColor
			return
		}
	}
}
