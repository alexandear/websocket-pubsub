package server

import (
	"log"
	"time"
)

const (
	maxClients = 5000
	castSize   = 1000
)

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]struct{}

	broadcastDuration time.Duration

	// Broadcast or unicast messages.
	cast chan *Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub(broadcast time.Duration) *Hub {
	return &Hub{
		broadcastDuration: broadcast,
		cast:              make(chan *Message, castSize),
		register:          make(chan *Client),
		unregister:        make(chan *Client),
		clients:           make(map[*Client]struct{}, maxClients),
	}
}

func (h *Hub) Run() {
	go func() {
		log.Printf("broadcasting with %s", h.broadcastDuration)

		ticker := time.NewTicker(h.broadcastDuration)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now().UTC()

			h.cast <- &Message{
				Communication: CommunicationBroadcast,
				Data:          &BroadcastData{Time: now},
			}
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
		case message := <-h.cast:
			for client := range h.clients {
				switch message.Communication {
				case CommunicationUnicast:
					data, ok := message.Data.(*UnicastData)
					if !ok {
						continue
					}

					if client.id == data.ClientID {
						client.send <- message.Data
					}
				case CommunicationBroadcast:
					select {
					case client.send <- message.Data:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				default:
					log.Printf("unknown message type %d", message.Communication)
				}
			}
		}
	}
}
