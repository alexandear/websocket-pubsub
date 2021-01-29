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
	cast chan *Message

	// Register requests from the clients.
	subscribe chan *Client

	// Unregister requests from clients.
	unsubscribe chan *Client

	broadcastFrequency time.Duration
}

func NewHub(broadcastFrequency time.Duration) *Hub {
	return &Hub{
		cast:               make(chan *Message, castSize),
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
		case message := <-h.cast:
			for client := range h.clients {
				if response := h.responseMessage(message, client); response != nil {
					client.response <- response
				}
			}
		}
	}
}

func (h *Hub) broadcastServerTime() {
	log.Printf("broadcasting server time with frequency %s", h.broadcastFrequency)

	ticker := time.NewTicker(h.broadcastFrequency)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now().UTC()

		h.cast <- &Message{
			CastType: Broadcast,
			Data: &BroadcastData{
				Time: now,
			},
		}
	}
}

func (h *Hub) responseMessage(message *Message, client *Client) ResponseMessage {
	switch message.CastType {
	case Unicast:
		data, ok := message.Data.(*UnicastData)
		if !ok {
			return nil
		}

		if client.id == data.ClientID {
			return ResponseUnicast{
				NumConnections: len(client.hub.clients),
			}
		}

		return nil
	case Broadcast:
		data, ok := message.Data.(*BroadcastData)
		if !ok {
			return nil
		}

		return ResponseBroadcast{
			ClientID: client.id,
			Time:     data.Time,
		}
	default:
		log.Printf("unknown cast type %d", message.CastType)

		return nil
	}
}
