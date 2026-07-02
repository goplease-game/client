package ws

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/coder/websocket"
	game "github.com/goplease-game/client"
	"github.com/goplease-game/client/config"
)

const (
	pingInterval = time.Second * 30
	dialTimeout  = time.Second * 10
	pingTimeout  = time.Second * 5
)

const playPath = "play/"

// WSClient manages a single WebSocket connection.
// It is safe to call Send from any goroutine.
// Incoming messages are delivered on the Inbox channel.
type WSClient struct { //nolint:revive
	inbox  chan InMessage // buffered; read by the game loop
	status ConnStatus     // read by screens; written only by wsClient goroutines

	mu       sync.Mutex
	conn     *websocket.Conn
	outbox   chan []byte
	stopOnce sync.Once
	stop     chan struct{}

	msgLogger *log.Logger
}

// wsURL returns the WebSocket server address from config.
func wsURL() string {
	return game.ServerWS(playPath)
}

// NewWSClient creates a new WSClient, enabling protocol logging to a file
// when dev mode and log protocol are turned on in config.
func NewWSClient() *WSClient {
	c := &WSClient{
		inbox:     make(chan InMessage, 128), //nolint:mnd
		outbox:    make(chan []byte, 128),    //nolint:mnd
		stop:      make(chan struct{}),
		status:    StatusDisconnected,
		msgLogger: nil,
	}

	dev := config.Get().DevMode
	if dev.Enabled && dev.LogProtocol {
		f, _ := os.OpenFile("protocol_log.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
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

// Inbox returns the channel on which incoming messages are delivered.
func (c *WSClient) Inbox() <-chan InMessage {
	return c.inbox
}

// Status returns the current connection status.
func (c *WSClient) Status() ConnStatus {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status
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

// Disconnect closes the connection gracefully.
func (c *WSClient) Disconnect() { c.close() }

// dial establishes the WebSocket connection, retrying with backoff on
// failure, then starts the read and write loops once connected.
func (c *WSClient) dial(playerID string) {
	var conn *websocket.Conn
	var err error

	// Retry loop with backoff (3 attempts).
	for attempt := range 3 {
		ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
		conn, _, err = websocket.Dial(ctx, wsURL(), nil)
		cancel()
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

// readLoop reads incoming messages from conn and forwards them to the
// inbox until the connection errors or closes.
func (c *WSClient) readLoop(conn *websocket.Conn) {
	defer c.close()
	ctx := context.Background()
	for {
		_, raw, err := conn.Read(ctx)
		if err != nil {
			code := websocket.CloseStatus(err)
			if code != websocket.StatusNormalClosure && code != websocket.StatusGoingAway {
				log.Printf("[ws] read error: %v", err)
			}
			return
		}

		c.logMessage(raw, false)

		var msg InMessage
		err = json.Unmarshal(raw, &msg)
		if err != nil {
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

// writeLoop sends queued outgoing messages and periodic pings on conn
// until the connection errors or a stop is requested.
// Note: Ping is a no-op when running as wasm in the browser — the
// browser's network stack handles WebSocket keepalive transparently,
// so this only provides dead-connection detection on native builds.
func (c *WSClient) writeLoop(conn *websocket.Conn) {
	ctx := context.Background()
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	for {
		select {
		case data := <-c.outbox:
			c.logMessage(data, true)

			err := conn.Write(ctx, websocket.MessageText, data)
			if err != nil {
				log.Printf("[ws] write error: %v", err)
				c.close()
				return
			}
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
			err := conn.Ping(pingCtx)
			cancel()
			if err != nil {
				c.close()
				return
			}
		case <-c.stop:
			_ = conn.Close(websocket.StatusNormalClosure, "")
			return
		}
	}
}

// close disconnects the underlying connection and signals the write loop
// to stop, exactly once.
func (c *WSClient) close() {
	c.stopOnce.Do(func() {
		c.mu.Lock()
		c.status = StatusDisconnected
		conn := c.conn
		c.mu.Unlock()
		if conn != nil {
			_ = conn.CloseNow()
		}
		close(c.stop)
	})
}

// logMessage writes msg to the protocol log if logging is enabled,
// marking it as outgoing or incoming.
func (c *WSClient) logMessage(msg []byte, out bool) {
	if c.msgLogger != nil {
		key := "<-"
		if out {
			key = "->"
		}

		c.msgLogger.Printf("%s %s", key, msg)
	}
}
