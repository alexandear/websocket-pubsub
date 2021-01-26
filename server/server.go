package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/alexandear/websocket-pubsub/request"
)

var addr = flag.String("addr", ":8080", "http service address")

var upgrader = websocket.Upgrader{}

type App struct {
	router *mux.Router

	mu       sync.Mutex
	channels map[string]*Channel
}

func (a *App) Initialize() {
	a.channels = make(map[string]*Channel, 5000)

	a.router.HandleFunc("/pubsub", a.Pubsub).Methods(http.MethodGet)
	a.router.HandleFunc("/ws", a.WebSocket).Methods(http.MethodGet)
}

func (a *App) Close() {
	a.mu.Lock()
	for i, c := range a.channels {
		if err := c.Close(); err != nil {
			log.Println("close channel", i, ":", err)
		}
	}
	a.mu.Unlock()
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
	case request.CommandUnsubscribe:
		a.commandUnsubscribe(w, r)

		return
	default:
		respondWithError(w, http.StatusBadRequest, "Unknown command")

		return
	}
}

func (a *App) commandSubscribe(w http.ResponseWriter, r *http.Request) {
	clientID := r.Header.Get("X-Client-Id")
	println(clientID)
	http.Redirect(w, r, "ws://"+r.Host+"/ws", http.StatusSeeOther)
}

func (a *App) commandUnsubscribe(w http.ResponseWriter, r *http.Request) {
	clientID := r.Header.Get("X-Client-Id")
	println(clientID)
}

func (a *App) WebSocket(w http.ResponseWriter, r *http.Request) {
	clientID := r.Header.Get("X-Client-Id")
	println(clientID)

	{
		a.mu.Lock()

		if _, ok := a.channels[clientID]; ok {
			a.mu.Unlock()
			respondWithError(w, http.StatusBadRequest, "Duplicated X-Client-Id")

			return
		}

		a.mu.Unlock()
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	a.mu.Lock()
	a.channels[clientID] = NewChannel(clientID, c)
	a.mu.Unlock()
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
	}
	a.Initialize()

	defer a.Close()

	log.Fatal(http.ListenAndServe(*addr, a.router))
}
