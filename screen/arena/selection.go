package arena

import (
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
)

// selectUnit is called when the player clicks the active unit on the board.
// It highlights the unit cell with a border and tints all reachable cells.
func (s *Screen) selectUnit(u ds.Unit) {
	if s.activeUnitMoved {
		return
	}
	if s.selectedUnitID == u.ID {
		// Second click on the same unit — deselect.
		s.deselectUnit()
		return
	}

	s.selectedUnitID = u.ID
	s.reachableCells = u.ReachableCells(s.board)

	// Highlight the unit's own cell with a border.
	if bc := s.boardCellWidget(u); bc != nil {
		bc.SetColor(unitFriendlyBgColor)
	}

	// Tint reachable cells.
	for _, pos := range s.reachableCells {
		if w := s.boardCellWidgets[pos]; w != nil {
			w.SetColor(unitMoveToCellColor)
		}
	}
}

// deselectUnit clears selection and restores all highlighted cells to their original colours.
func (s *Screen) deselectUnit() {
	if s.selectedUnitID == "" {
		return
	}

	u, ok := s.unitByID(s.selectedUnitID)
	if ok {
		if bc := s.boardCellWidget(u); bc != nil {
			bc.SetColor(unitFriendlyBgColor)
		}
	}

	// Restore reachable cells.
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

func (s *Screen) onReachableCellClicked(to ds.HexCoord) {
	u, ok := s.unitByID(s.selectedUnitID)
	if !ok {
		return
	}

	from := ds.HexCoord{Q: u.Pos.Q, R: u.Pos.R}

	s.deselectUnit()

	// Hide the unit icon on the source cell for the duration of the animation.
	if w := s.boardCellWidgets[from]; w != nil {
		w.RemoveChildren()
	}

	s.activeMoveAnim = newMoveAnim(
		unitImage(u.TemplateID),
		s.cellCentrePx(from),
		s.cellCentrePx(to),
		func() { s.finishMove(u, from, to) },
	)

	// Send to server immediately — the server doesn't need to wait for visuals.
	s.server.Send(ws.OutMessage{
		Action: ws.UnitMovedAction,
		Data: ds.UnitMovedPayload{
			UnitID: u.ID,
			Coord:  to,
		},
	})
}

// finishMove is called by the animation's onDone callback.
// It commits board state and updates cell visuals.
func (s *Screen) finishMove(u ds.Unit, from ds.HexCoord, to ds.HexCoord) {
	s.moveUnit(u, to)

	if s.selectedUnitID == u.ID || !u.IsOpponent {
		s.activeUnitMoved = true
	}

	s.activeMoveAnim = nil

	if fromW := s.boardCellWidgets[from]; fromW != nil {
		s.removePulseWidget(fromW)
		s.restoreSafeZoneCell(from)
		fromW.SetColor(boardCellBgColor)
		fromW.RemoveChildren()
	}

	if toW := s.boardCellWidgets[to]; toW != nil {
		targetBg := unitFriendlyBgColor
		if u.IsOpponent {
			targetBg = unitEnemyBgColor
		}

		toW.SetColor(targetBg)
		toW.RemoveChildren()
		buildBoardCard(toW, u, false)
	}

	if !u.IsOpponent {
		if toW := s.boardCellWidgets[to]; toW != nil {
			s.pulseHexWidgets = append(s.pulseHexWidgets, toW)
		}
	}
}

// removePulseWidget removes a specific cell from the pulse list.
func (s *Screen) removePulseWidget(w *ui.HexCellWidget) {
	for i, pw := range s.pulseHexWidgets {
		if pw == w {
			s.pulseHexWidgets = append(s.pulseHexWidgets[:i], s.pulseHexWidgets[i+1:]...)
			return
		}
	}
}

// restoreSafeZoneCell marks the cell at (r, c) unoccupied if it is a SafeZone,
// allowing future unit placement there.
func (s *Screen) restoreSafeZoneCell(coord ds.HexCoord) {
	for _, sc := range s.safeZoneCells {
		if sc.coord == coord {
			sc.occupied = false
			sc.baseColor = boardCellBgColor
			return
		}
	}
}
