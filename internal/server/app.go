package server

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	gws "github.com/gorilla/websocket"

	"github.com/alexandear/websocket-pubsub/internal/pkg/websocket"
)

const upgraderBufferSize = 1024

type HubI interface {
	Subscribe(client *Client)
	Unsubscribe(client *Client)
	Cast(data MessageData)
	Run()
}

type App struct {
	addr string

	upgrader gws.Upgrader
	hub      HubI
	router   *mux.Router
}

func New(addr string, broadcastFrequency time.Duration) *App {
	a := &App{
		addr: addr,
		upgrader: gws.Upgrader{
			ReadBufferSize:  upgraderBufferSize,
			WriteBufferSize: upgraderBufferSize,
		},
		hub:    NewHub(broadcastFrequency),
		router: mux.NewRouter(),
	}

	a.router.HandleFunc("/ws", a.serveWs).Methods(http.MethodGet)

	return a
}

func (a *App) Run() error {
	go a.hub.Run()

	log.Printf("listening on %s", a.addr)

	return http.ListenAndServe(a.addr, a.router)
}

// serveWs handles websocket requests from the peer.
func (a *App) serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade failed: %v", err)

		return
	}

	wsConn := websocket.NewConn(conn)
	client := NewClient(a.hub, wsConn)
	client.Run()
}
