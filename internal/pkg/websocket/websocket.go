package websocket

import (
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
)

var ErrClosedConn = errors.New("closed connection")

type Conn struct {
	conn *websocket.Conn
}

func NewConn(conn *websocket.Conn) *Conn {
	return &Conn{conn: conn}
}

func (c *Conn) ReadBinaryMessage() ([]byte, error) {
	messageType, message, err := c.conn.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			return nil, fmt.Errorf("read message failed: %w", err)
		}

		return nil, ErrClosedConn
	}

	if messageType != websocket.BinaryMessage {
		return nil, fmt.Errorf("unexpected message type: %d", messageType)
	}

	return message, nil
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) WriteCloseMessage() {
	_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func (c *Conn) WriteBinaryMessage(data []byte) error {
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}
