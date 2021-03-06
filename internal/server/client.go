package server

import (
	"context"
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

//go:generate mockgen -source=$GOFILE -package mock -destination mock/interfaces.go

type HubI interface {
	Subscribe(client ClientI)
	Unsubscribe(client ClientI)
	Cast(data CastData)
	Run(ctx context.Context)
}

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

	return client
}

func (c *Client) ID() string {
	return c.id
}

// Run allow collection of memory referenced by the caller by doing all work in new goroutines.
func (c *Client) Run(ctx context.Context) {
	go c.write()
	go c.read()

	for range ctx.Done() {
		return
	}
}

func (c *Client) CloseResponse() {
	close(c.response)
}

func (c *Client) Response(message ResponseMessage) {
	c.response <- message
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
				log.Printf("failed to read from client: %v", err)
			}

			return
		}

		if err := c.processCommand(message); err != nil {
			log.Printf("failed to process command: %v", err)
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
		c.hub.Subscribe(c)
	case command.Unsubscribe:
		c.hub.Unsubscribe(c)
	case command.NumConnections:
		c.hub.Cast(UnicastData{ClientID: c.id})
	default:
		c.hub.Unsubscribe(c)

		return nil
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
		var message ResponseMessage
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

func (c *Client) writeMessage(message ResponseMessage) error {
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
