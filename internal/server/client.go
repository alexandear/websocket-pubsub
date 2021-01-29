package server

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/alexandear/websocket-pubsub/internal/pkg/command"
	"github.com/alexandear/websocket-pubsub/internal/pkg/operation"
)

const (
	// Maximum message size allowed from peer.
	maxMessageSize = 512

	sendBufferSize = 256
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	// Generated client ID.
	id string

	hub *Hub

	// Websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan Data

	resp *Responder
}

func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	client := &Client{
		id:   uuid.New().String(),
		hub:  hub,
		conn: conn,
		send: make(chan Data, sendBufferSize),
		resp: &Responder{},
	}

	return client
}

// Start allow collection of memory referenced by the caller by doing all work in new goroutines.
func (c *Client) Start() {
	go c.write()
	go c.read()
}

// read pumps messages from the websocket connection to the hub.
func (c *Client) read() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)

	for {
		read, err := c.readMessage()
		if err != nil {
			log.Printf("failed to read: %v", err)
		}

		if !read {
			break
		}
	}
}

func (c *Client) readMessage() (bool, error) {
	messageType, message, err := c.conn.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			return false, fmt.Errorf("read message failed: %w", err)
		}

		return false, nil
	}

	if messageType != websocket.BinaryMessage {
		return true, fmt.Errorf("unexpected message type: %d", messageType)
	}

	req := &operation.ReqCommand{}
	if err := json.Unmarshal(message, req); err != nil {
		return true, fmt.Errorf("unmarshal ReqCommand failed: %w", err)
	}

	switch req.Command {
	case command.Subscribe:
	case command.Unsubscribe:
		c.hub.unregister <- c
	case command.NumConnections:
		c.hub.cast <- &Message{
			Communication: CommunicationUnicast,
			Data: &UnicastData{
				ClientID:       c.id,
				NumConnections: 0,
			},
		}
	default:
		c.hub.unregister <- c
	}

	return true, nil
}

// write pumps messages from the hub to the websocket connection.
func (c *Client) write() {
	defer func() {
		_ = c.conn.Close()
	}()

	opened := true
	for opened {
		var message Data
		message, opened = <-c.send

		if !opened {
			_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})

			return
		}

		if err := c.writeMessage(message); err != nil {
			log.Printf("write message failed: %v", err)
		}
	}
}

func (c *Client) writeMessage(data Data) error {
	resp, err := c.resp.Bytes(data)
	if err != nil {
		return fmt.Errorf("failed to create resp: %w", err)
	}

	if err := c.conn.WriteMessage(websocket.BinaryMessage, resp); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}
