package server

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	gws "github.com/gorilla/websocket"

	"github.com/alexandear/websocket-pubsub/internal/pkg/websocket"
)

const upgraderBufferSize = 1024

type App struct {
	addr string

	upgrader gws.Upgrader
	hub      HubI
	router   *mux.Router
}

func New(addr string, hub HubI) *App {
	a := &App{
		addr: addr,
		upgrader: gws.Upgrader{
			ReadBufferSize:  upgraderBufferSize,
			WriteBufferSize: upgraderBufferSize,
		},
		hub:    hub,
		router: mux.NewRouter(),
	}

	a.router.HandleFunc("/ws", a.serveWs).Methods(http.MethodGet)

	return a
}

func (a *App) Run(ctx context.Context) error {
	go a.hub.Run(ctx)

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
	client.Run(r.Context())
}
