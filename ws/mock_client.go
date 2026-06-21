package ws

import (
	"encoding/json"
	"fmt"
	"log"

	game "github.com/goplease-game/server"
	"github.com/goplease-game/server/api"
	"github.com/goplease-game/server/ds"
)

// MockClient implements Client over an in-process Session for practice and scenario modes.
type MockClient struct {
	session  *game.Session
	playerID ds.ID
	inbox    chan InMessage
	status   ConnStatus
}

// NewMockClient creates a MockClient that communicates with the given Session as the specified player.
func NewMockClient(session *game.Session, playerID ds.ID) *MockClient {
	m := &MockClient{
		session:  session,
		playerID: playerID,
		inbox:    make(chan InMessage, 128),
		status:   StatusConnected,
	}
	go m.readLoop()
	return m
}

// Inbox returns the channel on which inbound server messages are delivered.
func (m *MockClient) Inbox() <-chan InMessage { return m.inbox }

// Status returns the current connection status.
func (m *MockClient) Status() ConnStatus { return m.status }

// Connect is a no-op for MockClient as the session is already established at construction.
func (m *MockClient) Connect(_ string) {}

// Disconnect closes the player event channel, terminating the read loop.
func (m *MockClient) Disconnect() {
	close(m.session.P1Events)
}

// Send forwards an outbound message to the Session as the player's action.
func (m *MockClient) Send(msg OutMessage) {
	data, err := json.Marshal(msg.Data)
	if err != nil {
		log.Fatalf("[Send] marshal: %v", err)
		return
	}
	m.session.Handle(m.playerID, api.Action(msg.Action), data)
}

// readLoop forwards inbound Session events to the inbox channel.
func (m *MockClient) readLoop() {
	for msg := range m.session.P1Events {
		data, err := json.Marshal(msg.Data)
		if err != nil {
			fmt.Printf("[mock] readLoop marshal error: %s\n", err)
			return
		}
		m.inbox <- InMessage{
			Action: Action(msg.Action),
			Data:   data,
		}
	}
}
