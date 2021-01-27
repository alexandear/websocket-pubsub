package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/alexandear/websocket-pubsub/command"
	"github.com/alexandear/websocket-pubsub/operation"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	// Generated client ID.
	id string

	hub *Hub

	// Websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// Read pumps messages from the websocket connection to the hub.
func (c *Client) Read() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)

	for {
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v", err)
			}

			break
		}

		if messageType != websocket.BinaryMessage {
			log.Printf("unexpected message type: %d", messageType)

			continue
		}

		req := &operation.ReqCommand{}
		if err := json.Unmarshal(message, req); err != nil {
			log.Printf("unmarshal ReqCommand failed: %v", err)

			continue
		}

		switch req.Command {
		case command.Subscribe:
			log.Println("unexpected SUBSCRIBE command")
		case command.Unsubscribe:
			log.Println("received UNSUBSCRIBE command")

			c.hub.unregister <- c
		case command.NumConnections:
			log.Println("received NUM_CONNECTIONS command")

			c.hub.clientNumConnections <- c.id
		}
	}
}

// Write pumps messages from the hub to the websocket connection.
func (c *Client) Write() {
	defer func() {
		_ = c.conn.Close()
	}()

	opened := true
	for opened {
		var message []byte
		message, opened = <-c.send
		_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

		if !opened {
			log.Println("the hub closed the channel")

			_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})

			return
		}

		w, err := c.conn.NextWriter(websocket.BinaryMessage)
		if err != nil {
			log.Printf("next writer failed: %v", err)

			return
		}

		rb, err := c.respBytes(message)
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

func (c *Client) respBytes(message []byte) ([]byte, error) {
	if _, err := uuid.ParseBytes(message); err == nil {
		b, err := json.Marshal(&operation.RespNumConnections{
			NumConnections: len(c.hub.clients),
		})
		if err != nil {
			return nil, fmt.Errorf("marshal RespNumConnections failed: %w", err)
		}

		return b, nil
	}

	var t time.Time
	if err := t.UnmarshalBinary(message); err != nil {
		return nil, fmt.Errorf("unmarshal binary time failed: %w", err)
	}

	b, err := json.Marshal(&operation.RespBroadcast{
		ClientID:  c.id,
		Timestamp: t.Format(time.RFC3339),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal RespBroadcast failed: %w", err)
	}

	return b, nil
}
