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

type WsConn interface {
	SetReadLimit(limit int64)
	Close() error
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	id string

	hub  HubI
	conn WsConn

	// Buffered channel of outbound messages.
	response chan ResponseMessage
}

func NewClient(hub HubI, conn WsConn) *Client {
	client := &Client{
		id:       uuid.New().String(),
		hub:      hub,
		conn:     conn,
		response: make(chan ResponseMessage, sendBufferSize),
	}

	client.hub.Subscribe(client)

	return client
}

// Run allow collection of memory referenced by the caller by doing all work in new goroutines.
func (c *Client) Run() {
	go c.write()
	go c.read()
}

// read pumps messages from the websocket connection to the hub.
func (c *Client) read() {
	defer func() {
		c.hub.Unsubscribe(c)
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
		c.hub.Unsubscribe(c)
	case command.NumConnections:
		c.hub.Cast(UnicastData{ClientID: c.id})
	default:
		c.hub.Unsubscribe(c)
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
		var message interface{}
		message, opened = <-c.response

		if !opened {
			_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})

			return
		}

		if err := c.writeMessage(message); err != nil {
			log.Printf("write message failed: %v", err)
		}
	}
}

func (c *Client) writeMessage(message interface{}) error {
	var resp json.RawMessage

	switch m := message.(type) {
	case ResponseUnicast:
		r, err := json.Marshal(&operation.RespNumConnections{
			NumConnections: m.NumConnections,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal num connections response: %w", err)
		}

		resp = r
	case ResponseBroadcast:
		r, err := json.Marshal(&operation.RespBroadcast{
			ClientID:  m.ClientID,
			Timestamp: int(m.Time.Unix()),
		})
		if err != nil {
			return fmt.Errorf("failed to marshal broadcast response: %w", err)
		}

		resp = r
	default:
		return fmt.Errorf("unknown response message type: %+v", m)
	}

	if err := c.conn.WriteMessage(websocket.BinaryMessage, resp); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}
