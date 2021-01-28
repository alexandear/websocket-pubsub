package main

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	sendBufferSize = 256
)

type App struct {
	router   *mux.Router
	upgrader websocket.Upgrader
	hub      *Hub
}

func (a *App) Initialize() {
	a.router.HandleFunc("/ws", a.ServeWs).Methods(http.MethodGet)

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
		send: make(chan []byte, sendBufferSize),
	}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines.
	go client.Write()
	go client.Read()
}
