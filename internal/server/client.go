package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/alexandear/websocket-pubsub/internal/pkg/command"
	"github.com/alexandear/websocket-pubsub/internal/pkg/operation"
	"github.com/alexandear/websocket-pubsub/internal/pkg/websocket"
)

const (
	sendBufferSize = 256
)

type WsConn interface {
	Close() error
	ReadBinaryMessage() ([]byte, error)
	WriteBinaryMessage(data []byte) error
	WriteCloseMessage()
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

	for {
		message, err := c.conn.ReadBinaryMessage()
		if err != nil {
			if !errors.Is(err, websocket.ErrClosedConn) {
				log.Printf("failed to read: %v", err)
			}

			return
		}

		if err := c.processCommand(message); err != nil {
			log.Printf("faield to process command: %v", err)
		}
	}
}

func (c *Client) processCommand(data []byte) error {
	req := &operation.ReqCommand{}
	if err := json.Unmarshal(data, req); err != nil {
		return fmt.Errorf("unmarshal to ReqCommand failed: %w", err)
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

	return nil
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
			c.conn.WriteCloseMessage()

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

	if err := c.conn.WriteBinaryMessage(resp); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}
