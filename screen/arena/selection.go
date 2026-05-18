package arena

import (
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
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
		bc.SetBackgroundImage(image.NewBorderedNineSliceColor(
			unitFriendlyBgColor,
			selectedUnitBorderColor,
			3,
		))
	}

	// Tint reachable cells.
	for _, pos := range s.reachableCells {
		if w := s.boardCellWidgets[pos[0]][pos[1]]; w != nil {
			w.SetBackgroundImage(image.NewNineSliceColor(unitMoveToCellColor))
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
			bc.SetBackgroundImage(image.NewNineSliceColor(unitFriendlyBgColor))
		}
	}

	// Restore reachable cells.
	for _, pos := range s.reachableCells {
		cell := s.board[pos[0]][pos[1]]
		bg := boardCellBgColor
		if cell != nil && cell.Unit != nil {
			if cell.Unit.IsOpponent {
				bg = unitEnemyBgColor
			} else {
				bg = unitFriendlyBgColor
			}
		}
		if w := s.boardCellWidgets[pos[0]][pos[1]]; w != nil {
			w.SetBackgroundImage(image.NewNineSliceColor(bg))
		}
	}

	s.selectedUnitID = ""
	s.reachableCells = nil
}

// onReachableCellClicked starts the movement animation toward (toR, toC).
// Board state and visuals are updated only once the animation finishes (onDone).
func (s *Screen) onReachableCellClicked(toR, toC int) {
	u, ok := s.unitByID(s.selectedUnitID)
	if !ok {
		return
	}

	fromR, fromC := u.Row, u.Col

	s.deselectUnit()

	// Hide the unit icon on the source cell for the duration of the animation.
	if w := s.boardCellWidgets[fromR][fromC]; w != nil {
		w.RemoveChildren()
	}

	s.activeMoveAnim = newMoveAnim(
		unitImage(u.TemplateID),
		s.cellCentrePx(fromR, fromC),
		s.cellCentrePx(toR, toC),
		func() { s.finishMove(u, fromR, fromC, toR, toC) },
	)

	// Send to server immediately — the server doesn't need to wait for visuals.
	s.server.Send(ws.OutMessage{
		Action: ws.UnitMovedAction,
		Data: ds.UnitMovedPayload{
			UnitID: u.ID,
			ToRow:  toR,
			ToCol:  toC,
		},
	})
}

// finishMove is called by the animation's onDone callback.
// It commits board state and updates cell visuals.
func (s *Screen) finishMove(u ds.Unit, fromR, fromC, toR, toC int) {
	s.moveUnit(u, toR, toC)

	if s.selectedUnitID == u.ID || !u.IsOpponent {
		s.activeUnitMoved = true
	}

	s.activeMoveAnim = nil

	s.removePulseWidget(s.boardCellWidgets[fromR][fromC])

	if w := s.boardCellWidgets[fromR][fromC]; w != nil {
		s.restoreSafeZoneCell(fromR, fromC)
		w.SetBackgroundImage(image.NewNineSliceColor(boardCellBgColor))
		w.RemoveChildren()
	}

	if w := s.boardCellWidgets[toR][toC]; w != nil {
		targetBg := unitFriendlyBgColor
		if u.IsOpponent {
			targetBg = unitEnemyBgColor
		}

		w.SetBackgroundImage(image.NewNineSliceColor(targetBg))
		w.RemoveChildren()
		w.AddChild(centeredGraphic(unitImage(u.TemplateID)))
	}

	if !u.IsOpponent {
		if w := s.boardCellWidgets[toR][toC]; w != nil {
			s.pulseWidgets = append(s.pulseWidgets, w)
		}
	}
}

// removePulseWidget removes a specific container from the pulse list.
func (s *Screen) removePulseWidget(w *widget.Container) {
	for i, pw := range s.pulseWidgets {
		if pw == w {
			s.pulseWidgets = append(s.pulseWidgets[:i], s.pulseWidgets[i+1:]...)
			return
		}
	}
}

// restoreSafeZoneCell marks the cell at (r, c) unoccupied if it is a SafeZone,
// allowing future unit placement there.
func (s *Screen) restoreSafeZoneCell(r, c int) {
	for _, sc := range s.safeZoneCells {
		if sc.row == r && sc.col == c {
			sc.occupied = false
			sc.baseColor = boardCellBgColor
			return
		}
	}
}
