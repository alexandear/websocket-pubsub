package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
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
	redirect, err := c.subscribeRedirect(ctx, server, "/pubsub")
	if err != nil {
		return fmt.Errorf("subsribe failed: %w", err)
	}

	// nolint:bodyclose // does not need to be closed
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, redirect, nil)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}

	c.conn = conn

	return nil
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

func (c *Client) subscribeRedirect(ctx context.Context, host, path string) (string, error) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   path,
	}

	log.Printf("subscribing %d to %s", c.id, u.String())

	bs, err := json.Marshal(&operation.ReqCommand{
		Command: command.Subscribe,
	})
	if err != nil {
		return "", fmt.Errorf("marshal ReqCommand failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), bytes.NewReader(bs))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout:       httpClientTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to do request: %w", err)
	}

	if cerr := resp.Body.Close(); cerr != nil {
		return "", fmt.Errorf("failed to close: %w", cerr)
	}

	loc, err := resp.Location()
	if errors.Is(err, http.ErrNoLocation) {
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("failed to get location: %w", err)
	}

	return loc.String(), nil
}

func (c *Client) sendCommand(commandType command.Type) error {
	if c.conn == nil {
		return nil
	}

	if commandType == command.Subscribe {
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
