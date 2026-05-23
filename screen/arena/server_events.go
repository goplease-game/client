package arena

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
)

func (s *Screen) handleServerMessage(msg ws.InMessage) {
	fmt.Printf("received: %v\n", msg.Action)

	switch msg.Action {
	case ws.PlaceUnitAction:
		s.handlePlaceUnit()
	case ws.EndRoundAction:
		s.handleEndRound()
	case ws.EndTurnAction:
		s.handleEndTurn()
	case ws.PlayUnitAction:
		s.handlePlayUnit(msg.Data)
	case ws.WaitingForOpponent:
		s.handleWaitingForOpponent()
	case ws.UnitPlacedAction:
		s.handleOpponentUnitPlaced(msg.Data)
	case ws.UnitMovedAction:
		s.handleUnitMoved(msg.Data)
	}
}

// handlePlaceUnit is called when the server enters the unit-placement phase.
func (s *Screen) handlePlaceUnit() {
	s.ready = true
	s.unitPlacedThisTurn = false
	s.highlightActiveUnit("")
	s.hideAbilityPanel()
	s.setupUnitPanel()

	if len(s.player.Units) == 1 {
		s.setStatus(fmt.Sprintf("%s is ready to be deployed", s.player.Units[0].Name))
	} else {
		s.setStatus("Deploy a unit to the board")
	}
}

// handleEndRound is called when the current round ends and the player may end their turn.
func (s *Screen) handleEndRound() {
	s.setNextActionLabel("END\nROUND")
	s.enableNextActionBtn()
	s.endTurnBtnPulseActive = true
	s.setStatus("You can end your turn")
	s.highlightActiveUnit("")
}

// handleEndTurn is called when the player's turn can be ended.
func (s *Screen) handleEndTurn() {
	s.setNextActionLabel("END\nTURN")
	s.enableNextActionBtn()
	s.endTurnBtnPulseActive = true
	s.highlightActiveUnit("")
}

// handlePlayUnit is called when it is a specific unit's turn to act.
func (s *Screen) handlePlayUnit(data json.RawMessage) {
	var payload ds.PlayUnitPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Fatal("handlePlayUnit unmarshal:", err)
	}

	unit, ok := s.unitByID(payload.UnitID)
	if !ok {
		log.Fatal("handlePlayUnit: unit not found:", payload.UnitID)
	}

	if s.unitPanelIn {
		s.footerRef.RemoveChild(s.unitPanelRef)
		s.unitPanelIn = false
	}

	s.activeUnitMoved = false
	s.deselectUnit()

	s.showAbilityPanel(unit)
	s.highlightActiveUnit(payload.UnitID)
	s.setNextActionLabel("SKIP\nTURN")
	s.enableNextActionBtn()
	s.setStatus(fmt.Sprintf("Play unit: %s", unit.Name))
}

// handleWaitingForOpponent is called when we are waiting for the other player.
func (s *Screen) handleWaitingForOpponent() {
	s.hideAbilityPanel()
	s.setStatus("Waiting for opponent...")
}

// handleOpponentUnitPlaced is called when the opponent places a unit on the board.
func (s *Screen) handleOpponentUnitPlaced(data json.RawMessage) {
	var payload ds.PlaceUnitPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Fatal("handleOpponentUnitPlaced unmarshal:", err)
	}

	coord := payload.Coord

	cellWidget := s.boardCellWidgets[coord]
	if cellWidget == nil {
		return
	}

	cellWidget.SetColor(unitEnemyBgColor)
	buildBoardCard(cellWidget, *payload.Unit, false)

	opponentUnitID := payload.Unit.ID

	u := *payload.Unit
	u.Pos = coord
	u.IsOpponent = true

	s.board.Cells[coord].Unit = &u

	s.addUnitToQueue(opponentUnitID)
}

func (s *Screen) handleUnitMoved(data json.RawMessage) {
	var payload ds.UnitMovedPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Fatal("handleUnitMoved unmarshal:", err)
	}

	u, ok := s.unitByID(payload.UnitID)
	if !ok {
		log.Printf("[warn] handleUnitMoved: unit %s not found", payload.UnitID)
		return
	}

	from := u.Pos
	to := payload.Coord

	if s.selectedUnitID == u.ID {
		s.deselectUnit()
	}

	if w := s.boardCellWidgets[from]; w != nil {
		w.RemoveChildren()
	}

	s.activeMoveAnim = newMoveAnim(
		unitImage(u.TemplateID),
		s.cellCentrePx(from),
		s.cellCentrePx(to),
		func() { s.finishMove(u, from, to) },
	)
}
