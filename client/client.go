package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/alexandear/websocket-pubsub/command"
	"github.com/alexandear/websocket-pubsub/operation"
)

const (
	httpClientTimeout = 2 * time.Second
)

type Client struct {
	id         int
	httpClient *http.Client
	conn       *websocket.Conn
}

func NewClient(id int) *Client {
	app := &Client{
		id: id,
		httpClient: &http.Client{
			Timeout:       httpClientTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
		},
	}

	return app
}

func (c *Client) Close() {
	if c.conn == nil {
		return
	}

	if err := c.conn.Close(); err != nil {
		log.Printf("close failed: %v", err)
	}
}

func (c *Client) Subscribe(ctx context.Context, server string) error {
	// nolint:bodyclose // does not need to be closed
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, "ws://"+server+"/ws", nil)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}

	c.conn = conn

	return c.sendCommand(command.Subscribe)
}

func (c *Client) Read() {
	if c.conn == nil {
		return
	}

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure,
				websocket.CloseNoStatusReceived) {
				log.Printf("unexpected close error: %v", err)
			}

			return
		}

		broadcast := &operation.RespBroadcast{}
		if err := json.Unmarshal(message, broadcast); err != nil {
			log.Printf("failed to unmarshal RespBroadcast: %v", err)

			continue
		}

		if broadcast.Timestamp != 0 {
			log.Printf("Client ID: %s, server time: %v", broadcast.ClientID, time.Unix(int64(broadcast.Timestamp), 0))

			continue
		}

		numConnections := &operation.RespNumConnections{}
		if err := json.Unmarshal(message, numConnections); err != nil {
			log.Printf("failed to unmarshal RespNumConnections: %v", err)

			continue
		}

		log.Printf("Number of connections: %d", numConnections.NumConnections)

		continue
	}
}

func (c *Client) NumConnections() error {
	return c.sendCommand(command.NumConnections)
}

func (c *Client) Unsubscribe() error {
	return c.sendCommand(command.Unsubscribe)
}

func (c *Client) sendCommand(commandType command.Type) error {
	if c.conn == nil {
		return nil
	}

	log.Printf("sending %s command", commandType)

	b, err := json.Marshal(&operation.ReqCommand{
		Command: commandType,
	})
	if err != nil {
		return fmt.Errorf("marshal ReqCommand failed: %w", err)
	}

	if err := c.conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		return fmt.Errorf("write binary message failed: %w", err)
	}

	return nil
}
