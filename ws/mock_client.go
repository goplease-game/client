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

	case EndUnitPlacement:
		// place unit for second player
		gs := mock.GetGameState()
		unit := mock.PickRandomUnit()
		// no more units, move so make a move
		if unit == nil {
			// ...
			m.inbox <- InMessage{Action: UnitPlacedAction}
		}

		row, col := mock.GetRandomSafeZoneCell(len(gs.Board), len(gs.Board[0]))

		m.inbox <- InMessage{Action: UnitPlacedAction}

	case PlaceUnitAction:
		m.inbox <- InMessage{
			Action: UnitPlacedAction,
			//Data:   msg.Data,
		}

		time.Sleep(2 * time.Second)

		m.inbox <- InMessage{
			Action: UnitPlacedAction,
			Data:   json.RawMessage(`{"unit_id":"opp_1","row":2,"col":3}`),
		}

	case CancelMatchAction:
		m.inbox <- InMessage{Action: MatchCancelledAction}
	}
}
