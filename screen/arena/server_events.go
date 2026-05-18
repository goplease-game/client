package arena

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
)

func (s *Screen) handleServerMessage(msg ws.InMessage) {
	fmt.Printf("received: %v\n", msg)

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
	}
}

// handlePlaceUnit is called when the server enters the unit-placement phase.
func (s *Screen) handlePlaceUnit() {
	s.ready = true
	s.unitPlacedThisTurn = false
	s.hideAbilityPanel()
	s.setupUnitPanel()
	s.setStatus("Place a unit on the board")
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

	if payload.Row < 0 || payload.Row >= len(s.boardCellWidgets) ||
		payload.Col < 0 || payload.Col >= len(s.boardCellWidgets[payload.Row]) {
		return
	}

	cell := s.boardCellWidgets[payload.Row][payload.Col]
	cell.SetBackgroundImage(image.NewNineSliceColor(unitEnemyBgColor))
	cell.AddChild(widget.NewGraphic(
		widget.GraphicOpts.Image(unitImage(payload.Unit.TemplateID)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	))

	opponentUnitID := fmt.Sprintf("opp_%d_%d_%d", payload.Unit.TemplateID, payload.Row, payload.Col)
	s.board[payload.Row][payload.Col].Unit = &ds.Unit{
		ID:         opponentUnitID,
		TemplateID: payload.Unit.TemplateID,
		Row:        payload.Row,
		Col:        payload.Col,
		IsOpponent: true,
	}

	s.addUnitToQueue(opponentUnitID)
}
