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
	id string

	hub *Hub

	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)

	for {
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read error:", err)
			}

			log.Println("read:", err)

			break
		}

		if messageType != websocket.BinaryMessage {
			log.Println("received unexpected message type:", messageType)

			continue
		}

		req := &operation.ReqCommand{}
		if err := json.Unmarshal(message, req); err != nil {
			log.Println("unmarshal:", err)

			return
		}

		if req.Command == command.Unsubscribe {
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
				log.Println("write close:", err)

				return
			}
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	defer func() {
		_ = c.conn.Close()
	}()

	for {
		select {
		case b, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				log.Println("the hub closed the channel")

				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Println(err)
				return
			}

			_, _ = w.Write(func() []byte {
				var t time.Time
				if err := t.UnmarshalBinary(b); err != nil {
					log.Println("unmarshal time:", err)

					return nil
				}

				resp := &operation.RespBroadcast{
					ClientID:  c.id,
					Timestamp: t,
				}
				b, err := resp.MarshalJSON()
				if err != nil {
					log.Println("marshal resp:", err)

					return nil
				}

				return b
			}())

			if err := w.Close(); err != nil {
				log.Println("close", err)
				return
			}
		}
	}
}
