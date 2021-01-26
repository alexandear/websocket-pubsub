package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/alexandear/websocket-pubsub/request"
)

var addr = flag.String("addr", ":8080", "http service address")

type App struct {
	router   *mux.Router
	upgrader websocket.Upgrader
	hub      *Hub

	mu       sync.Mutex
	channels map[string]*Channel
}

func (a *App) Initialize() {
	a.channels = make(map[string]*Channel, 5000)

	a.router.HandleFunc("/pubsub", a.Pubsub).Methods(http.MethodGet)
	a.router.HandleFunc("/ws", a.serveWs).Methods(http.MethodGet)

	go a.hub.run()
}

func (a *App) Pubsub(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var comm request.Command
	if err := decoder.Decode(&comm); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")

		return
	}

	switch comm.CommandType {
	case request.CommandSubscribe:
		a.commandSubscribe(w, r)

		return
	default:
		respondWithError(w, http.StatusBadRequest, "Unknown command")

		return
	}
}

func (a *App) commandSubscribe(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "ws://"+r.Host+"/ws", http.StatusSeeOther)
}

// serveWs handles websocket requests from the peer.
func (a *App) serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{
		id:   uuid.New().String(),
		hub:  a.hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines.
	go client.writePump()
	go client.readPump()
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	r := mux.NewRouter()

	a := &App{
		router: r,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		hub: newHub(),
	}
	a.Initialize()

	log.Fatal(http.ListenAndServe(*addr, a.router))
}
