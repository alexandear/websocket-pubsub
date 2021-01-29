package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/alexandear/websocket-pubsub/internal/pkg/command"
	"github.com/alexandear/websocket-pubsub/internal/pkg/operation"
	"github.com/alexandear/websocket-pubsub/internal/pkg/websocket"
)

var ErrNilConn = errors.New("nil ws conn")

type WsConn interface {
	Close() error
	ReadBinaryMessage() ([]byte, error)
	WriteBinaryMessage(data []byte) error
}

type Client struct {
	conn WsConn
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) SetConn(conn WsConn) {
	c.conn = conn
}

func (c *Client) Subscribe() error {
	return c.sendCommand(command.Subscribe)
}

func (c *Client) NumConnections() error {
	return c.sendCommand(command.NumConnections)
}

func (c *Client) Unsubscribe() error {
	return c.sendCommand(command.Unsubscribe)
}

func (c *Client) sendCommand(commandType command.Type) error {
	if c.conn == nil {
		return ErrNilConn
	}

	log.Printf("sending %s command", commandType)

	b, err := json.Marshal(&operation.ReqCommand{
		Command: commandType,
	})
	if err != nil {
		return fmt.Errorf("marshal ReqCommand failed: %w", err)
	}

	if err := c.conn.WriteBinaryMessage(b); err != nil {
		return fmt.Errorf("write binary message failed: %w", err)
	}

	return nil
}

func (c *Client) Read() {
	if c.conn == nil {
		return
	}

	for {
		message, err := c.conn.ReadBinaryMessage()
		if err != nil {
			if !errors.Is(err, websocket.ErrClosedConn) {
				log.Printf("failed to read: %v", err)
			}

			return
		}

		resp, err := determineOperationResp(message)
		if err != nil {
			log.Printf("failed to determine operation: %v", err)

			continue
		}

		switch r := resp.(type) {
		case operation.RespBroadcast:
			log.Printf("Client ID: %s, server time: %v", r.ClientID, time.Unix(int64(r.Timestamp), 0))
		case operation.RespNumConnections:
			log.Printf("Num connections: %d", r.NumConnections)
		}
	}
}

func determineOperationResp(message []byte) (operation.Resp, error) {
	var broadcast operation.RespBroadcast
	if err := json.Unmarshal(message, &broadcast); err != nil {
		return nil, fmt.Errorf("failed to unmarshal RespBroadcast: %w", err)
	}

	if broadcast.Timestamp != 0 {
		return broadcast, nil
	}

	var numConnections operation.RespNumConnections
	if err := json.Unmarshal(message, &numConnections); err != nil {
		return nil, fmt.Errorf("failed to unmarshal RespNumConnections: %w", err)
	}

	if numConnections.NumConnections != 0 {
		return numConnections, nil
	}

	return nil, errors.New("unknown operation resp")
}

func (c *Client) Close() {
	if c.conn == nil {
		return
	}

	if err := c.conn.Close(); err != nil {
		log.Printf("close failed: %v", err)
	}
}
