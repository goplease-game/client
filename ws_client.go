package client

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const wsURL = "ws://localhost:8080/goplease/"

type Action string

const (
	ConnectedAction       Action = "connected"
	NewGameAction         Action = "new_game"
	SearchingOppAction    Action = "searching_opp"
	PlaceUnitAction       Action = "place_unit"
	UnitPlacedAction      Action = "unit_placed"
	OppDisconnectedAction Action = "opp_disconnected"
	CancelMatchAction     Action = "cancel_match"
	MatchCancelledAction  Action = "match_canceled"
	ErrorAction           Action = "error"
)

// WSMessage mirrors the server's OutgoingMsg.
type WSMessage struct {
	Action Action          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

// WSClient manages a single WebSocket connection.
// It is safe to call Send from any goroutine.
// Incoming messages are delivered on the Inbox channel.
type WSClient struct {
	Inbox  chan WSMessage // buffered; read by the game loop
	Status ConnStatus     // read by screens; written only by wsClient goroutines

	mu       sync.Mutex
	conn     *websocket.Conn
	outbox   chan []byte
	stopOnce sync.Once
	stop     chan struct{}
}

type ConnStatus int

const (
	StatusDisconnected ConnStatus = iota
	StatusConnecting
	StatusConnected
	StatusError
)

func NewWSClient() *WSClient {
	return &WSClient{
		Inbox:  make(chan WSMessage, 128),
		outbox: make(chan []byte, 128),
		stop:   make(chan struct{}),
		Status: StatusDisconnected,
	}
}

// Connect dials the server in the background.
// Safe to call multiple times — ignored if already connected.
func (c *WSClient) Connect(playerID string) {
	c.mu.Lock()
	if c.Status == StatusConnecting || c.Status == StatusConnected {
		c.mu.Unlock()
		return
	}
	c.Status = StatusConnecting
	c.mu.Unlock()

	go c.dial(playerID)
}

func (c *WSClient) dial(playerID string) {
	url := wsURL + "?player_id=" + playerID

	var conn *websocket.Conn
	var err error

	// Retry loop with backoff (3 attempts).
	for attempt := 1; attempt <= 3; attempt++ {
		conn, _, err = websocket.DefaultDialer.Dial(url, nil)
		if err == nil {
			break
		}
		log.Printf("[ws] dial attempt %d failed: %v", attempt, err)
		time.Sleep(time.Duration(attempt) * time.Second)
	}
	if err != nil {
		log.Printf("[ws] could not connect: %v", err)
		c.mu.Lock()
		c.Status = StatusError
		c.mu.Unlock()
		return
	}

	c.mu.Lock()
	c.conn = conn
	c.Status = StatusConnected
	c.mu.Unlock()

	log.Printf("[ws] connected as player %s", playerID)

	go c.readLoop(conn)
	go c.writeLoop(conn)
}

func (c *WSClient) readLoop(conn *websocket.Conn) {
	defer c.close()
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[ws] read error: %v", err)
			}
			return
		}
		var msg WSMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Printf("[ws] bad JSON: %v", err)
			continue
		}
		select {
		case c.Inbox <- msg:
		default:
			log.Println("[ws] inbox full, dropping message")
		}
	}
}

func (c *WSClient) writeLoop(conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case data := <-c.outbox:
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("[ws] write error: %v", err)
				c.close()
				return
			}
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.close()
				return
			}
		case <-c.stop:
			_ = conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}
	}
}

// Send encodes v as JSON and enqueues it for sending.
func (c *WSClient) Send(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("[ws] marshal error: %v", err)
		return
	}
	select {
	case c.outbox <- b:
	default:
		log.Println("[ws] outbox full, dropping message")
	}
}

func (c *WSClient) close() {
	c.stopOnce.Do(func() {
		c.mu.Lock()
		c.Status = StatusDisconnected
		if c.conn != nil {
			_ = c.conn.Close()
		}
		c.mu.Unlock()
		close(c.stop)
	})
}

// Disconnect closes the connection gracefully.
func (c *WSClient) Disconnect() { c.close() }

func (c *WSClient) NewGame() {
	c.Send(WSMessage{
		Action: NewGameAction,
		Data:   nil,
	})
}
