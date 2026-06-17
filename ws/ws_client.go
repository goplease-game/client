package ws

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ognev-dev/goplease-ebitengine-client/config"
)

// WSClient manages a single WebSocket connection.
// It is safe to call Send from any goroutine.
// Incoming messages are delivered on the Inbox channel.
type WSClient struct {
	inbox  chan InMessage // buffered; read by the game loop
	status ConnStatus     // read by screens; written only by wsClient goroutines

	mu       sync.Mutex
	conn     *websocket.Conn
	outbox   chan []byte
	stopOnce sync.Once
	stop     chan struct{}

	msgLogger *log.Logger
}

func wsURL() string {
	return config.Get().ServerAddr
}

func NewWSClient() *WSClient {
	c := &WSClient{
		inbox:     make(chan InMessage, 128),
		outbox:    make(chan []byte, 128),
		stop:      make(chan struct{}),
		status:    StatusDisconnected,
		msgLogger: nil,
	}

	dev := config.Get().DevMode
	if dev.Enabled && dev.LogProtocol {
		f, _ := os.OpenFile("protocol_log.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		c.msgLogger = log.New(f, "", log.Ltime|log.Lmicroseconds)
	}

	return c
}

// Connect dials the server in the background.
// Safe to call multiple times — ignored if already connected.
func (c *WSClient) Connect(playerID string) {
	c.mu.Lock()
	if c.status == StatusConnecting || c.status == StatusConnected {
		c.mu.Unlock()
		return
	}
	c.status = StatusConnecting
	c.mu.Unlock()

	go c.dial(playerID)
}

func (c *WSClient) Inbox() <-chan InMessage {
	return c.inbox
}

func (c *WSClient) Status() ConnStatus {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status
}

func (c *WSClient) dial(playerID string) {
	var conn *websocket.Conn
	var err error

	// Retry loop with backoff (3 attempts).
	for attempt := range 3 {
		conn, _, err = websocket.DefaultDialer.Dial(wsURL(), nil)
		if err == nil {
			break
		}
		log.Printf("[ws] dial attempt %d failed: %v", attempt, err)
		time.Sleep(time.Duration(attempt) * time.Second)
	}
	if err != nil {
		log.Printf("[ws] could not connect: %v", err)
		c.mu.Lock()
		c.status = StatusError
		c.mu.Unlock()
		return
	}

	c.mu.Lock()
	c.conn = conn
	c.status = StatusConnected
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

		c.logMessage(raw, false)

		var msg InMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Printf("[ws] bad JSON: %v", err)
			continue
		}
		select {
		case c.inbox <- msg:
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
			c.logMessage(data, true)

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
func (c *WSClient) Send(v OutMessage) {
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

func (c *WSClient) logMessage(msg []byte, out bool) {
	if c.msgLogger != nil {
		key := "<-"
		if out {
			key = "->"
		}

		c.msgLogger.Printf("%s %s", key, msg)
	}
}

func (c *WSClient) close() {
	c.stopOnce.Do(func() {
		c.mu.Lock()
		c.status = StatusDisconnected
		if c.conn != nil {
			_ = c.conn.Close()
		}
		c.mu.Unlock()
		close(c.stop)
	})
}

// Disconnect closes the connection gracefully.
func (c *WSClient) Disconnect() { c.close() }
