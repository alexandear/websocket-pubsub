package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/alexandear/websocket-pubsub/internal/pkg/command"
	"github.com/alexandear/websocket-pubsub/internal/pkg/operation"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

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
}

func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		id:   uuid.New().String(),
		hub:  hub,
		conn: conn,
		send: make(chan Data, sendBufferSize),
	}
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
			return false, err
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
			Data:          &UnicastData{ClientID: c.id},
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
		_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

		if !opened {
			_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})

			return
		}

		w, err := c.conn.NextWriter(websocket.BinaryMessage)
		if err != nil {
			log.Printf("next writer failed: %v", err)

			return
		}

		rb, err := c.newResp(message)
		if err != nil {
			log.Printf("resp bytes failed: %v", err)

			return
		}

		if _, err = w.Write(rb); err != nil {
			log.Printf("write failed: %v", err)

			continue
		}

		if err := w.Close(); err != nil {
			log.Printf("close failed: %v", err)

			continue
		}
	}
}

func (c *Client) newResp(message Data) ([]byte, error) {
	switch message.Type() {
	case MessageDataNumConn:
		return c.newRespNumConnections()
	case MessageDataTime:
		d, ok := message.(*BroadcastData)
		if !ok {
			return nil, errors.New("wrong broadcast data")
		}

		return c.newRespBroadcast(d.Time)
	default:
		return nil, errors.New("unknown message data type")
	}
}

func (c *Client) newRespNumConnections() ([]byte, error) {
	b, err := json.Marshal(&operation.RespNumConnections{
		NumConnections: len(c.hub.clients),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal RespNumConnections failed: %w", err)
	}

	return b, nil
}

func (c *Client) newRespBroadcast(t time.Time) ([]byte, error) {
	b, err := json.Marshal(&operation.RespBroadcast{
		ClientID:  c.id,
		Timestamp: int(t.Unix()),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal RespBroadcast failed: %w", err)
	}

	return b, nil
}
