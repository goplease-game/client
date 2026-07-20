package ws

import (
	"sync"

	game "github.com/goplease-game/game-server"
	"github.com/goplease-game/game-server/ds"
)

// ClientProvider owns the active client and allows switching between
// a real WebSocket client and a mock client for practice mode.
// All screens hold a *ClientProvider instead of a Client directly.
type ClientProvider struct {
	mu      sync.Mutex
	current Client
}

// NewClientProvider returns a provider with no active client.
// SwitchToReal is called by MainScreen on every entry.
func NewClientProvider() *ClientProvider {
	return &ClientProvider{}
}

// Get returns the current client.
func (p *ClientProvider) Get() Client {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.current
}

// SwitchToReal disconnects the current client if needed and installs a
// fresh WSClient. No-op if already on a real client.
func (p *ClientProvider) SwitchToReal() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, already := p.current.(*WSClient); already {
		return
	}
	if p.current != nil {
		p.current.Disconnect()
	}
	p.current = NewWSClient()
}

// SwitchToMock disconnects the current client and installs a MockClient
// backed by the given session and player ID.
func (p *ClientProvider) SwitchToMock(session *game.Session, playerID ds.ID) *MockClient {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.current != nil {
		p.current.Disconnect()
	}
	m := NewMockClient(session, playerID)
	p.current = m
	return m
}
