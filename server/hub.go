package main

import (
	"log"
	"time"
)

const (
	maxClients = 5000

	broadcastFrequency = 2 * time.Second
)

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	clients map[*Client]struct{} // Registered clients.

	// Broadcast current time to the clients.
	broadcast chan []byte

	// Send num connections to client by id.
	clientNumConnections chan string

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:            make(chan []byte),
		clientNumConnections: make(chan string),
		register:             make(chan *Client),
		unregister:           make(chan *Client),
		clients:              make(map[*Client]struct{}, maxClients),
	}
}

func (h *Hub) Run() {
	go func() {
		ticker := time.NewTicker(broadcastFrequency)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now().UTC()

			b, err := now.MarshalBinary()
			if err != nil {
				log.Printf("marshal binary time failed: %v", err)

				continue
			}

			h.broadcast <- b
		}
	}()

	for {
		select {
		case client := <-h.register:
			h.clients[client] = struct{}{}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.clientNumConnections:
			for client := range h.clients {
				if client.id == message {
					select {
					case client.send <- []byte(message):
					default:
					}
				}
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
