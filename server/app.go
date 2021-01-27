package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/alexandear/websocket-pubsub/command"
	"github.com/alexandear/websocket-pubsub/operation"
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
	a.router.HandleFunc("/pubsub", a.Pubsub).Methods(http.MethodGet)
	a.router.HandleFunc("/ws", a.ServeWs).Methods(http.MethodGet)

	go a.hub.Run()
}

func (a *App) Pubsub(w http.ResponseWriter, r *http.Request) {
	req := &operation.ReqCommand{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")

		return
	}

	if req.Command == command.Subscribe {
		a.commandSubscribe(w, r)

		return
	}

	respondWithError(w, http.StatusBadRequest, "Unknown command")
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

func (a *App) commandSubscribe(w http.ResponseWriter, r *http.Request) {
	log.Println("received SUBSCRIBE command")

	http.Redirect(w, r, "ws://"+r.Host+"/ws", http.StatusSeeOther)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	log.Printf("error: code=%d, message=%s", code, message)

	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}
