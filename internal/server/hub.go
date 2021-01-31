package server

import (
	"context"
	"log"
	"time"
)

const (
	maxClients = 5000
	castSize   = 1000
)

type ClientI interface {
	ID() string
	CloseResponse()
	Response(message ResponseMessage)
}

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients.
	clients map[ClientI]struct{}

	// Broadcast or unicast messages.
	cast chan CastData

	// Register requests from the clients.
	subscribe chan ClientI

	// Unregister requests from clients.
	unsubscribe chan ClientI

	broadcastFrequency time.Duration
}

func NewHub(broadcastFrequency time.Duration) *Hub {
	return &Hub{
		cast:               make(chan CastData, castSize),
		subscribe:          make(chan ClientI),
		unsubscribe:        make(chan ClientI),
		clients:            make(map[ClientI]struct{}, maxClients),
		broadcastFrequency: broadcastFrequency,
	}
}

func (h *Hub) Run(ctx context.Context) {
	go h.broadcastServerTime()

	for {
		select {
		case client := <-h.subscribe:
			h.clients[client] = struct{}{}
		case client := <-h.unsubscribe:
			if _, ok := h.clients[client]; ok {
				client.CloseResponse()
				delete(h.clients, client)
			}
		case data := <-h.cast:
			for client := range h.clients {
				if response := h.responseMessage(data, client.ID()); response != nil {
					client.Response(response)
				}
			}
		}
	}
}

func (h *Hub) Subscribe(client ClientI) {
	h.subscribe <- client
}

func (h *Hub) Unsubscribe(client ClientI) {
	h.unsubscribe <- client
}

func (h *Hub) Cast(data CastData) {
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

func (h *Hub) responseMessage(data CastData, clientID string) ResponseMessage {
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
