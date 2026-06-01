package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/mock"
)

const (
	mockDelay = 800 * time.Millisecond
)

type MockClient struct {
	inbox  chan InMessage
	status ConnStatus
}

func NewMockClient() *MockClient {
	return &MockClient{
		inbox:  make(chan InMessage, 128),
		status: StatusDisconnected,
	}
}

func (m *MockClient) Inbox() <-chan InMessage { return m.inbox }
func (m *MockClient) Status() ConnStatus      { return m.status }

func (m *MockClient) Connect(playerID string) {
	m.status = StatusConnected
	log.Printf("[mock] connected as %s", playerID)
	m.inbox <- InMessage{Action: ConnectedAction}
}

func (m *MockClient) Disconnect() {
	m.status = StatusDisconnected
}

func (m *MockClient) Send(msg OutMessage) {
	log.Printf("[mock] client sent: %s", msg.Action)
	go m.handleLogic(msg)
}

func (m *MockClient) handleLogic(msg OutMessage) {
	switch msg.Action {
	case NewGameAction:
		m.onNewGame()

	case ReadyToPlay:
		m.onReadyToPlay()

	case UnitPlacedAction:
		m.onUnitPlaced(msg.Data.(ds.UnitPlacedPayload))

	case UnitMovedAction:
		m.onUnitMoved(msg.Data.(ds.UnitMovedPayload))

	case EndTurnAction:
		m.onEndTurn()
	case UseAbility:
		m.onAbilityUsed(msg.Data.(ds.UseAbilityPayload))

	case CancelMatchAction:
		m.inbox <- InMessage{Action: MatchCancelledAction}
	}
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

func (m *MockClient) onNewGame() {
	data, err := mock.LoadData("new_game.json")
	if err != nil {
		log.Fatal(err)
	}

	var payload ds.NewGamePayload
	if err = json.Unmarshal(data, &payload); err != nil {
		log.Fatal(err)
	}

	mock.NewGameState(payload)
	m.inbox <- InMessage{Action: NewGameAction, Data: data}
}

func (m *MockClient) onReadyToPlay() {
	m.send(WaitingForOpponent)
	time.Sleep(mockDelay)

	gs := mock.GetGameState()
	p1HasUnits := len(gs.Players[0].Units) > 0
	p2HasUnits := len(gs.Players[1].Units) > 0

	if !p1HasUnits && !p2HasUnits {
		// All units already on the board — skip placement, go straight to play.
		m.advanceGameLoop()
		return
	}

	m.send(PlaceUnitAction)
}

// onUnitPlaced is called after the real player drops a unit onto the board.
func (m *MockClient) onUnitPlaced(data ds.UnitPlacedPayload) {
	time.Sleep(mockDelay)
	gs := mock.GetGameState()

	unit := mock.PickUnitFromHandByTemplateP1(data.TemplateID)
	unit.Pos = data.Coord
	mock.PlaceUnitAt(unit, data.Coord)
	mock.AddUnitToQueue(unit)

	gs.Players[0].UnitsPlacedThisRound++

	m.runPlacementPhase()
}

// onUnitMoved updates the board state when the real player moves a unit.
// No response needed — the player ends their turn explicitly.
func (m *MockClient) onUnitMoved(data ds.UnitMovedPayload) {
	unit := mock.GetUnitByID(data.UnitID)
	if unit == nil {
		log.Printf("[mock] onUnitMoved: unit %s not found", data.UnitID)
		return
	}
	mock.PlaceUnitAt(unit, data.Coord)
	unit.Pos = data.Coord
}

// onEndTurn is the core game-loop driver.
// Called whenever the real player clicks "End Turn" / "End Round".
func (m *MockClient) onEndTurn() {
	m.send(WaitingForOpponent)
	time.Sleep(mockDelay)

	m.advanceGameLoop()
}

// onAbilityUsed handles a UseAbility request, applies the resulting state,
// and sends the updated state back to the client.
func (m *MockClient) onAbilityUsed(load ds.UseAbilityPayload) {
	states, err := mock.HandleAbility(load)
	if err != nil {
		m.sendErr(err.Error())
		return
	}

	if len(states) > 0 {
		m.sendApplyStates(states...)
	}
}

func (m *MockClient) sendApplyStates(st ...ds.ApplyState) {
	if len(st) == 0 {
		return
	}

	data, err := json.Marshal(st)
	if err != nil {
		log.Fatal(err)
	}

	m.send(ApplyState, data)
}

// ---------------------------------------------------------------------------
// Game loop
// ---------------------------------------------------------------------------

// advanceGameLoop determines what happens next and drives mock AI turns
// until it's the real player's turn again.
func (m *MockClient) advanceGameLoop() {
	gs := mock.GetGameState()

	switch gs.Phase {
	case mock.PlayPhase:
		m.advancePlayPhase()

	case mock.PlacementPhase:
		m.runPlacementPhase()
	}
}

// nextUnitToPlay returns the next unit in the queue that hasn't acted yet,
// or nil if all units in the queue have played this round.
func (m *MockClient) nextUnitToPlay() *ds.Unit {
	gs := mock.GetGameState()
	if gs.ActiveUnit < 0 || gs.ActiveUnit >= len(gs.UnitsQueue) {
		return nil
	}
	return gs.UnitsQueue[gs.ActiveUnit]
}

// playUnit handles a single unit's turn.
// If the unit belongs to the real player, send play_unit and return (wait for player input).
// If it belongs to the mock, simulate the move and continue the loop.
func (m *MockClient) playUnit(unit *ds.Unit) {
	gs := mock.GetGameState()
	gs.ActiveUnit++

	states := []ds.ApplyState{}
	// decrease status duration
	for t, st := range unit.Statuses {
		if st.Duration == mock.StatusPermanentDuration {
			continue
		}

		st.Duration--
		if st.Duration < 1 {
			states = append(states, ds.ApplyState{RemoveStatus: new(t), ToUnitID: unit.ID})
		}
	}

	// shield always decreased by 1 every turn
	if unit.CurrentShield > 0 {
		unit.CurrentShield--
		states = append(states,
			ds.ApplyState{ChangeShield: new(-1), ToUnitID: unit.ID},
			ds.ApplyState{SetShield: new(unit.CurrentShield), ToUnitID: unit.ID},
		)
	}
	m.sendApplyStates(states...)

	if unit.OwnerID != mock.MockedPlayerID {
		// Real player's unit — hand control back to the client.
		m.sendPlayUnit(unit.ID)
		return
	}

	// Mock unit: simulate movement if possible.
	m.simulateMockUnitTurn(unit)

	// Continue the loop after a short pause.
	time.Sleep(mockDelay)
	m.advanceGameLoop()
}

// simulateMockUnitTurn moves the mock unit to a random reachable cell.
func (m *MockClient) simulateMockUnitTurn(unit *ds.Unit) {
	cells := unit.ReachableCells(mock.GetGameState().Board)
	if len(cells) == 0 {
		log.Printf("[mock] unit %s has no reachable cells, skipping", unit.ID)
		return
	}

	pos := mock.RandomReachableCell(*unit)
	mock.PlaceUnitAt(unit, pos)
	unit.Pos = pos

	data, err := json.Marshal(ds.UnitMovedPayload{
		UnitID: unit.ID,
		Coord:  pos,
	})
	if err != nil {
		log.Fatal(err)
	}
	m.inbox <- InMessage{Action: UnitMovedAction, Data: data}
}

// runPlacementPhase handles the end-of-queue placement step:
//  1. If the real player hasn't placed yet → ask them to place.
//  2. If the mock player hasn't placed yet (and has units) → place for them,
//     then ask the real player to place (if they also haven't).
//  3. If both have placed (or have no units left) → start a new round.
func (m *MockClient) runPlacementPhase() {
	gs := mock.GetGameState()

	p1Done := gs.Players[0].UnitsPlacedThisRound >= mock.UnitsPerPlacementPhase
	p2Done := gs.Players[1].UnitsPlacedThisRound >= mock.UnitsPerPlacementPhase

	if p1Done && p2Done {
		m.startNewRound()
		return
	}

	actor := m.getPlacementActor()

	if actor == 0 {
		if !p1Done {
			m.send(PlaceUnitAction)
			return
		}
	} else {
		if !p2Done {
			m.mockPlaceUnit(gs)
			time.Sleep(mockDelay)
			m.advanceGameLoop()
			return
		}
	}

	m.advanceGameLoop()
}

func (m *MockClient) advancePlayPhase() {
	gs := mock.GetGameState()
	nextUnit := m.nextUnitToPlay()

	if nextUnit == nil {
		gs.Phase = mock.PlacementPhase
		m.runPlacementPhase()
		return
	}

	m.playUnit(nextUnit)
}

// mockPlaceUnit picks a random unit from the mock player's hand and places it.
func (m *MockClient) mockPlaceUnit(gs *mock.GameState) {
	unit := mock.PickRandomUnitOfFromHandP2()
	if unit == nil {
		log.Println("[mock] mockPlaceUnit: no units in hand")
		return
	}

	pos := mock.GetRandomUnoccupiedOpponentSafeZoneCell()
	unit.Pos = pos
	mock.PlaceUnitAt(unit, pos)
	mock.AddUnitToQueue(unit)
	gs.Players[1].UnitsPlacedThisRound++

	data, err := json.Marshal(ds.PlaceUnitPayload{Coord: pos, Unit: unit})
	if err != nil {
		log.Fatal(err)
	}
	m.inbox <- InMessage{Action: UnitPlacedAction, Data: data}
}

// startNewRound resets per-round state and begins the next round's play phase.
func (m *MockClient) startNewRound() {
	gs := mock.GetGameState()
	gs.CurrentRound++
	gs.ActiveUnit = 0
	gs.Phase = mock.PlayPhase

	m.send(NewRound)

	if mock.UnitsPerPlacementPhase >= 2 {
		mock.UnitsPerPlacementPhase--
	}
	gs.Players[0].UnitsPlacedThisRound = 0
	gs.Players[1].UnitsPlacedThisRound = 0

	m.advanceGameLoop()
}

func (m *MockClient) send(action Action, dataOpt ...json.RawMessage) {
	msg := InMessage{Action: action}
	if len(dataOpt) > 0 {
		msg.Data = dataOpt[0]
	}

	m.inbox <- msg
}

func (m *MockClient) sendErr(e string) {
	v := ds.ErrorResponse{Message: e}
	data, err := json.Marshal(v)
	if err != nil {
		log.Fatal(err)
	}

	m.send(ErrorAction, data)
}

func (m *MockClient) sendPlayUnit(unitID string) {
	data, err := json.Marshal(ds.PlayUnitPayload{UnitID: unitID})
	if err != nil {
		log.Fatal(err)
	}
	m.inbox <- InMessage{Action: PlayUnitAction, Data: data}
}

func (m *MockClient) getPlacementActor() int {
	gs := mock.GetGameState()
	p1 := gs.Players[0].UnitsPlacedThisRound
	p2 := gs.Players[1].UnitsPlacedThisRound

	if p1 < p2 {
		return 0 // P1
	}
	if p2 < p1 {
		return 1 // P2
	}

	return 0 // tie-breaker: P1 starts
}
