package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/mock"
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

func (m *MockClient) Inbox() <-chan InMessage {
	return m.inbox
}

func (m *MockClient) Status() ConnStatus {
	return m.status
}

func (m *MockClient) Connect(playerID string) {
	m.status = StatusConnected
	log.Printf("[mock] connected as %s", playerID)

	m.inbox <- InMessage{Action: ConnectedAction}
}

func (m *MockClient) Disconnect() {
	m.status = StatusDisconnected
}

func (m *MockClient) Send(v OutMessage) {
	b, _ := json.Marshal(v)
	var msg OutMessage
	json.Unmarshal(b, &msg)

	log.Printf("[mock] client sent: %s", msg.Action)

	go m.handleLogic(msg)
}

func (m *MockClient) handleLogic(msg OutMessage) {
	switch msg.Action {
	case NewGameAction:
		data, err := mock.LoadData("new_game.json")
		if err != nil {
			log.Fatal(err)
		}

		var newGameData ds.NewGamePayload
		err = json.Unmarshal(data, &newGameData)
		if err != nil {
			log.Fatal(err)
		}

		mock.NewGameState(newGameData)

		m.inbox <- InMessage{
			Action: NewGameAction,
			Data:   data,
		}

	case ReadyToPlay:
		m.inbox <- InMessage{
			Action: WaitingForOpponent,
		}

		time.Sleep(1 * time.Second)

		m.inbox <- InMessage{
			Action: PlaceUnitAction,
		}

	case EndTurnAction:
		m.inbox <- InMessage{Action: WaitingForOpponent}
		time.Sleep(1 * time.Second)

		// place unit for second player
		gs := mock.GetGameState()
		unit := mock.PickRandomUnit()
		// no more units, move so make a move | use skill
		if unit == nil {
			// ...
			// TODO send "play_unit"
			m.inbox <- InMessage{Action: "unit_moved"}
		}

		row, col := mock.GetRandomSafeZoneCell(len(gs.Board), len(gs.Board[0]))

		upl := ds.PlaceUnitPayload{
			Row:  row,
			Col:  col,
			Unit: unit,
		}

		data, err := json.Marshal(upl)
		if err != nil {
			log.Fatal(err)
		}

		m.inbox <- InMessage{
			Action: UnitPlacedAction,
			Data:   data,
		}

		// imagine other player also clicked "end_turn"

	case UnitPlacedAction:
		//var req ds.UnitPlacedPayload
		// in real-life scenario server should validate placement and notify other player about new unit

		// after unit placed there is nothing to do, let user end turn
		m.inbox <- InMessage{
			Action: EndTurnAction,
		}

	case CancelMatchAction:
		m.inbox <- InMessage{Action: MatchCancelledAction}
	}
}
