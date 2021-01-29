package server

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	sendBufferSize = 256
)

type App struct {
	upgrader websocket.Upgrader
	hub      *Hub
}

func New(upgrader websocket.Upgrader, hub *Hub) *App {
	return &App{
		upgrader: upgrader,
		hub:      hub,
	}
}

func (a *App) Start() {
	go a.hub.Run()
}

// ServeWs handles websocket requests from the peer.
func (a *App) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade failed: %v", err)

		return
	}

	client := &Client{
		id:   uuid.New().String(),
		hub:  a.hub,
		conn: conn,
		send: make(chan Data, sendBufferSize),
	}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines.
	go client.Write()
	go client.Read()
}
