package main

import (
	"encoding/json"
	"log"
	"time"

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
//
// The application runs Read in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
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

			b, err := json.Marshal(&operation.RespNumConnections{
				NumConnections: len(c.hub.clients),
			})
			if err != nil {
				log.Printf("marshal RespNumConnections failed: %v", err)

				continue
			}

			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
				log.Printf("write binary message failed: %v", err)
			}
		}
	}
}

// Write pumps messages from the hub to the websocket connection.
//
// A goroutine running Write is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) Write() {
	defer func() {
		_ = c.conn.Close()
	}()

	opened := true
	for opened {
		var b []byte
		b, opened = <-c.send
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

		var t time.Time
		if uerr := t.UnmarshalBinary(b); uerr != nil {
			log.Printf("unmarshal binary time failed: %v", uerr)

			continue
		}

		rb, err := json.Marshal(&operation.RespBroadcast{
			ClientID:  c.id,
			Timestamp: t.Format(time.RFC3339),
		})
		if err != nil {
			log.Printf("marshal RespBroadcast failed: %v", err)

			continue
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
