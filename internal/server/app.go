package server

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const ()

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

	client := NewClient(a.hub, conn)
	a.hub.register <- client

	go client.Start()
}
