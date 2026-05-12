package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/ognev-dev/goplease-ebitengine-client/mock"
)

type MockClient struct {
	inbox  chan Message
	status ConnStatus
}

func NewMockClient() *MockClient {
	return &MockClient{
		inbox:  make(chan Message, 128),
		status: StatusDisconnected,
	}
}

func (m *MockClient) Inbox() <-chan Message {
	return m.inbox
}

func (m *MockClient) Status() ConnStatus {
	return m.status
}

func (m *MockClient) Connect(playerID string) {
	m.status = StatusConnected
	log.Printf("[mock] connected as %s", playerID)

	m.inbox <- Message{Action: ConnectedAction}
}

func (m *MockClient) Disconnect() {
	m.status = StatusDisconnected
}

func (m *MockClient) Send(v any) {
	b, _ := json.Marshal(v)
	var msg Message
	json.Unmarshal(b, &msg)

	log.Printf("[mock] client sent: %s", msg.Action)

	go m.handleLogic(msg)
}

func (m *MockClient) handleLogic(msg Message) {
	switch msg.Action {
	case NewGameAction:
		m.inbox <- Message{Action: SearchingOppAction}
		time.Sleep(1 * time.Second)

		data, err := mock.LoadData("new_game.json")
		if err != nil {
			log.Fatal(err)
		}

		m.inbox <- Message{
			Action: NewGameAction,
			Data:   data,
		}

	case PlaceUnitAction:
		m.inbox <- Message{
			Action: UnitPlacedAction,
			Data:   msg.Data,
		}

		time.Sleep(2 * time.Second)

		m.inbox <- Message{
			Action: UnitPlacedAction,
			Data:   json.RawMessage(`{"unit_id":"opp_1","row":2,"col":3}`),
		}

	case CancelMatchAction:
		m.inbox <- Message{Action: MatchCancelledAction}
	}
}
