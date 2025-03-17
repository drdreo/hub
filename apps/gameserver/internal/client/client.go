package client

import (
	"github.com/drdreo/hub/gameserver/internal/interfaces"
	"github.com/drdreo/hub/gameserver/internal/session"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 2048
)

// Client represents a connected websocket client
type Client interface {
	ID() string
	Send(message []byte) error
	Room() interfaces.Room
	SetRoom(room interfaces.Room)
	Close()
}

// WebSocketClient implements the Client interface
type WebSocketClient struct {
	id        string
	conn      *websocket.Conn
	send      chan []byte
	room      interfaces.Room
	mu        sync.Mutex
	closed    bool
	OnMessage func(message []byte)
}

// NewClient creates a new WebSocketClient
func NewClient(conn *websocket.Conn) *WebSocketClient {
	return &WebSocketClient{
		id:        uuid.New().String(),
		conn:      conn,
		send:      make(chan []byte, 256),
		closed:    false,
		OnMessage: func(message []byte) {},
	}
}

// ID returns the client's unique ID
func (c *WebSocketClient) ID() string {
	return c.id
}

// Send queues a message to be sent to the client
func (c *WebSocketClient) Send(message []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return websocket.ErrCloseSent
	}

	select {
	case c.send <- message:
		return nil
	default:
		return websocket.ErrCloseSent
	}
}

// Room returns the client's current room
func (c *WebSocketClient) Room() interfaces.Room {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.room
}

// SetRoom updates the client's current room
func (c *WebSocketClient) SetRoom(room interfaces.Room) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.room = room
}

// Close terminates the client connection
func (c *WebSocketClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	c.closed = true

	// Store session if client is in a room
	if c.room != nil {
		// Get the global session store (see integration notes below)
		sessionStore := session.GetSessionStore()

		// Extract relevant player info from room state
		var playerInfo interface{}
		if state, ok := c.room.State().(map[string]interface{}); ok {
			if players, exists := state["players"].(map[string]interface{}); exists {
				playerInfo = players[c.id]
			}
		}

		sessionStore.StoreSession(c.id, session.SessionData{
			ClientID:   c.id,
			RoomID:     c.room.ID(),
			GameType:   c.room.GameType(),
			LastActive: time.Now(),
			// Add game-specific data if needed
			ExtraData: map[string]interface{}{
				"playerInfo": playerInfo,
			},
		})
		c.room.Leave(c)
	}

	c.conn.Close()
	close(c.send)
}

// StartPumps begins reading from and writing to the websocket
func (c *WebSocketClient) StartPumps() {
	go c.writePump()
	go c.readPump()
}

// readPump pumps messages from the websocket to the hub
func (c *WebSocketClient) readPump() {
	defer c.Close()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Call the message handler
		c.OnMessage(message)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			// Start JSON array
			w.Write([]byte("["))
			w.Write(message)

			// Add queued messages to the current websocket message with a delimiter
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte(","))
				w.Write(<-c.send)
			}
			// Close JSON array
			w.Write([]byte("]"))

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
