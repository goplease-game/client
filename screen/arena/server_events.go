package arena

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/sfx"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
	"golang.org/x/image/colornames"
)

// handleServerMessage dispatches an incoming server message to the appropriate handler.
func (s *Screen) handleServerMessage(msg ws.InMessage) {
	fmt.Printf("[arena] received: %v\n", msg.Action)
	if msg.Data != nil {
		fmt.Printf("JSON: %s\n", string(msg.Data))
	}

	switch msg.Action {
	case ws.ErrorAction:
		s.handleServerError(msg.Data)
	case ws.YouWin, ws.YouLose, ws.OpponentSurrendered:
		s.handleGameOver(msg.Action)
	case ws.PlaceUnitAction:
		s.handlePlaceUnit()
	case ws.EndTurnAction:
		s.handleEndTurn()
	case ws.PlayUnitAction:
		s.handlePlayUnit(msg.Data)
	case ws.WaitingForOpponent:
		s.handleWaitingForOpponent()
	case ws.UnitPlacedAction:
		s.handleOpponentUnitPlaced(msg.Data)
	case ws.NewRound:
		s.handleNewRound(msg.Data)
	case ws.UnitMovedAction:
		s.handleUnitMoved(msg.Data)
	case ws.ApplyState:
		s.handleApplyState(msg.Data)
	case ws.UseAbility:
		var payload ds.UseAbilityPayload
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			log.Fatal("handleUseAbility unmarshal:", err)
		}
		s.handleUseAbility(payload)
	default:
		fmt.Printf("[arena] unhandled action: %v\n", msg.Action)
	}
}

// handlePlaceUnit is called when the server enters the unit-placement phase.
// It marks the screen as ready and prompts the player to deploy a unit.
func (s *Screen) handlePlaceUnit() {
	s.ready = true
	s.unitPlacedThisTurn = false
	s.activeUnitID = ""
	s.highlightActiveUnit()
	s.hideAbilityPanel()
	s.setupUnitPanel()

	if len(s.player.Units) == 1 {
		s.setStatus(fmt.Sprintf("%s is ready to be deployed", s.player.Units[0].Name))
	} else {
		s.setStatus("Deploy a unit to the board")
	}

	s.hideNextAction()
	s.startTurnTimer()
}

func (s *Screen) handleServerError(data json.RawMessage) {
	var msg ds.ErrorResponse
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Fatal("handleServerError unmarshal:", err)
	}

	s.setStatus("ERROR: " + msg.Message)
}

func (s *Screen) handleGameOver(reason ws.Action) {
	switch reason {
	case ws.YouWin, ws.OpponentSurrendered:
		var explain string
		if reason == ws.OpponentSurrendered {
			explain = "Your opponent surrendered"
		}
		s.showGameOverOverlay(true, explain)
	case ws.YouLose:
		s.showGameOverOverlay(false, "")
	}
}

// handleEndTurn is called when the server signals the player may end their turn.
func (s *Screen) handleEndTurn() {
	s.setNextActionLabel("END\nTURN")
	s.enableNextActionBtn()
	s.endTurnBtnPulseActive = true
	s.activeUnitID = ""
	s.highlightActiveUnit()
}

// handlePlayUnit is called when it is a specific unit's turn to act.
// It shows the unit's ability panel, highlights it on the board, and enables the Next button.
func (s *Screen) handlePlayUnit(data json.RawMessage) {
	var payload ds.PlayUnitPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Fatal("handlePlayUnit unmarshal:", err)
	}

	unit := s.unitByID(payload.UnitID)
	if unit == nil {
		log.Fatal("handlePlayUnit: unit not found:", payload.UnitID)
	}

	if s.unitPanelIn {
		s.footerRef.RemoveChild(s.unitPanelRef)
		s.unitPanelIn = false
	}

	unit.CurrentAP = unit.BaseAP

	s.activeUnitID = payload.UnitID
	s.activeUnitMoved = false
	s.deselectUnit()
	s.highlightActiveUnit()
	s.showAbilityPanel(unit)
	s.enableNextActionBtn()
	s.updateActiveUnitStatusLabel()

	s.showNextActionBtn()
	s.updateNextActionLabel()

	s.startTurnTimer()
	s.ready = true
}

// handleWaitingForOpponent is called when the local player is waiting for the opponent.
func (s *Screen) handleWaitingForOpponent() {
	s.activeUnitID = ""
	s.hideAbilityPanel()
	s.hideUnitPanel()
	s.showNextActionHourglass()
	s.setStatus("Waiting for opponent...")
}

// handleOpponentUnitPlaced is called when the opponent places a unit on the board.
// It renders the unit card on the destination cell and adds it to the turn queue.
func (s *Screen) handleOpponentUnitPlaced(data json.RawMessage) {
	var payload ds.PlaceUnitPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Fatal("handleOpponentUnitPlaced unmarshal:", err)
	}

	cellWidget := s.boardCellWidgets[payload.Coord]
	if cellWidget == nil {
		return
	}

	cellWidget.SetColor(unitEnemyBgColor)
	s.buildBoardCard(cellWidget, payload.Unit, false)

	u := *payload.Unit
	u.Pos = payload.Coord
	u.IsOpponent = true

	s.board.Cells[payload.Coord].Unit = &u
	s.addUnitToQueue(&u)
	sfx.Play(unitPlacedSound)
}

func (s *Screen) handleNewRound(data json.RawMessage) {
	s.roundNumber++
	s.showNewRoundBanner(s.roundNumber)

	if s.activeUnitID == "" {
		return
	}

	if u := s.unitByID(s.activeUnitID); u != nil {
		s.showAbilityPanel(u)
	}
}

// handleUnitMoved is called when any unit (friendly or opponent) moves on the board.
// It starts the movement animation; finishMove is called when the animation completes.
func (s *Screen) handleUnitMoved(data json.RawMessage) {
	var payload ds.UnitMovedPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Fatal("handleUnitMoved unmarshal:", err)
	}

	u := s.unitByID(payload.UnitID)
	if u == nil {
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

	s.addMoveAnim(s.moveUnitAnim(u, to))
}

// handleApplyState processes a batch of atomic state mutations from the server.
// Each ApplyState is applied sequentially to the target unit.
func (s *Screen) handleApplyState(data json.RawMessage) {
	var payload []ds.ApplyState
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Fatal("handleApplyState unmarshal:", err)
	}

	var skipTurn bool

	// Apply state data immediately (HP, AP, statuses etc).
	for _, st := range payload {
		if st.SkipTurn && st.ToUnitID == s.activeUnitID {
			unit := s.unitByID(st.ToUnitID)
			if unit != nil && !unit.IsOpponent {
				skipTurn = true
			}
		}

		if st.SetPhantomAP != nil {
			s.player.PhantomAP = *st.SetPhantomAP
		}

		target := s.unitByID(st.ToUnitID)
		if target == nil {
			continue
		}
		s.applyStateImmediate(target, st)
	}

	// Visual feedback waits for fx to finish.
	if s.pendingVisuals != nil {
		s.pendingVisuals.applyStates = payload
		s.pendingVisuals.serverDone = true
		s.tryFlushPendingVisuals(s.pendingVisuals)
		return
	}

	// No pending fx — show visuals immediately.
	for _, st := range payload {
		if target := s.unitByID(st.ToUnitID); target != nil {
			s.applyStateVisuals(target, st)
		}
	}

	if skipTurn {
		s.handleWaitingForOpponent()
		s.server.Send(ws.OutMessage{Action: ws.EndTurnAction})
	}
}

func (s *Screen) handleUseAbility(load ds.UseAbilityPayload) {
	unit := s.unitByID(load.UnitID)
	if unit == nil {
		log.Fatalf("handleUseAbility: unit %s not found", load.UnitID)
	}

	pending := &pendingVisuals{}
	s.pendingVisuals = pending
	s.playAbilityFx(load.AbilityID, unit, load.Target, func() {
		pending.fxDone = true
		s.tryFlushPendingVisuals(pending)
	})

	ab := ability.ByID(load.AbilityID)
	unit.SetCooldown(ab.ID, ab.Cooldown)
}

// applyStateImmediate applies data mutations to the unit (no visuals).
// Called as soon as ApplyState arrives from the server.
func (s *Screen) applyStateImmediate(target *ds.Unit, st ds.ApplyState) {
	// --- Movement ---
	if st.MoveTo != nil {
		s.moveUnitForced(target, *st.MoveTo)
	}

	// --- Absolute values ---
	if st.SetHP != nil {
		target.CurrentHP = *st.SetHP
	}
	if st.SetBaseHP != nil {
		target.BaseHP = *st.SetBaseHP
	}
	if st.SetAP != nil {
		target.CurrentAP = *st.SetAP
		if target.ID == s.activeUnitID && !target.IsOpponent {
			s.showAbilityPanel(target)
			s.updateNextActionLabel()
		}
	}
	if st.SetMP != nil {
		target.CurrentMP = *st.SetMP
	}
	if st.SetShield != nil {
		target.CurrentShield = *st.SetShield
	}
	if st.SetAtk != nil {
		target.CurrentAtk = *st.SetAtk
	}
	if st.SetCooldown != nil {
		for abID, cd := range *st.SetCooldown {
			target.SetCooldown(abID, cd)
		}
	}

	// --- Status effects ---
	if st.AddStatus != nil {
		s.addUnitStatus(target, *st.AddStatus, st.AddStatusMeta)
	}
	if st.RemoveStatus != nil {
		s.removeUnitStatus(target, *st.RemoveStatus)
	}

	if st.SetStatusDuration != nil {
		s.updateUnitStatusDuration(target, st.SetStatusDuration)
	}

	// --- Death ---
	if st.IsDead {
		s.killUnit(target)
	}
}

// applyStateVisuals shows floating text and refreshes board card.
// Called only after fx animation has finished.
func (s *Screen) applyStateVisuals(target *ds.Unit, st ds.ApplyState) {
	if st.IsDead {
		return // death visuals handled in killUnit
	}

	if st.ShowText != nil {
		s.showFloatingText(target.Pos, *st.ShowText, colornames.Gold)
	}

	// --- Floating text for delta changes ---
	if st.ChangeHP != nil {
		s.showFloatingStat(target.Pos, *st.ChangeHP, "HP")
	}
	if st.ChangeShield != nil {
		s.showFloatingStat(target.Pos, *st.ChangeShield, "Shield")
	}
	if st.ChangeAtk != nil {
		s.showFloatingStat(target.Pos, *st.ChangeAtk, "ATK")
	}
	if st.ChangeAP != nil {
		s.showFloatingStat(target.Pos, *st.ChangeAP, "AP")
	}
	if st.UseAbility != nil {
		pl := *st.UseAbility
		ab := ability.ByID(pl.AbilityID)
		s.showFloatingText(target.Pos, ab.Name, colornames.Gold)
		s.handleUseAbility(pl)
	}

	s.showUnitOnBoard(target)
	s.rebuildQueuePanel()
	if target.ID == s.activeUnitID {
		s.showAbilityPanel(target)
	}
}
