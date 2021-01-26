package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type Packet struct {
	mt      int
	message []byte
}

type Channel struct {
	clientID string
	conn     *websocket.Conn
	send     chan Packet // Outgoing packets queue.
}

func NewChannel(clientID string, conn *websocket.Conn) *Channel {
	c := &Channel{
		clientID: clientID,
		conn:     conn,
		send:     make(chan Packet, 100),
	}

	go c.reader()
	go c.writer()

	return c
}

func (c *Channel) handle(pkt Packet) {
	c.send <- pkt
}

func (c *Channel) reader() {
	defer c.conn.Close()

	for {
		mt, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("recv: %s", message)

		c.handle(Packet{
			mt:      mt,
			message: message,
		})
	}
}

func (c *Channel) writer() {
	for pkt := range c.send {
		if err := c.conn.WriteMessage(pkt.mt, pkt.message); err != nil {
			log.Println("write:", err)
		}
	}
}
