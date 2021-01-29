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

	// Broadcast or unicast messages.
	cast chan MessageData

	// Register requests from the clients.
	subscribe chan *Client

	// Unregister requests from clients.
	unsubscribe chan *Client

	broadcastFrequency time.Duration
}

func NewHub(broadcastFrequency time.Duration) *Hub {
	return &Hub{
		cast:               make(chan MessageData, castSize),
		subscribe:          make(chan *Client),
		unsubscribe:        make(chan *Client),
		clients:            make(map[*Client]struct{}, maxClients),
		broadcastFrequency: broadcastFrequency,
	}
}

func (h *Hub) Run() {
	go h.broadcastServerTime()

	for {
		select {
		case client := <-h.subscribe:
			h.clients[client] = struct{}{}
		case client := <-h.unsubscribe:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.response)
			}
		case data := <-h.cast:
			for client := range h.clients {
				if response := h.responseMessage(data, client.id); response != nil {
					client.response <- response
				}
			}
		}
	}
}

func (h *Hub) Subscribe(client *Client) {
	h.subscribe <- client
}

func (h *Hub) Unsubscribe(client *Client) {
	h.unsubscribe <- client
}

func (h *Hub) Cast(data MessageData) {
	h.cast <- data
}

func (h *Hub) broadcastServerTime() {
	log.Printf("broadcasting server time with frequency %s", h.broadcastFrequency)

	ticker := time.NewTicker(h.broadcastFrequency)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now().UTC()

		h.cast <- BroadcastData{
			Time: now,
		}
	}
}

func (h *Hub) responseMessage(data MessageData, clientID string) ResponseMessage {
	switch data := data.(type) {
	case UnicastData:
		if clientID != data.ClientID {
			return nil
		}

		return ResponseUnicast{
			NumConnections: len(h.clients),
		}
	case BroadcastData:
		return ResponseBroadcast{
			ClientID: clientID,
			Time:     data.Time,
		}
	default:
		log.Printf("unknown data type %+v", data)

		return nil
	}
}
