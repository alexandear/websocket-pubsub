package server

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	flag "github.com/spf13/pflag"

	"github.com/alexandear/websocket-pubsub/internal/server"
)

const (
	upgraderBufferSize = 1024

	defaultBroadcast = 100 * time.Millisecond
)

func Exec() error {
	addr := flag.String("addr", ":8080", "http service address")
	broadcast := flag.Duration("broadcast", defaultBroadcast, "broadcast frequency")

	flag.Parse()

	upgrader := websocket.Upgrader{
		ReadBufferSize:  upgraderBufferSize,
		WriteBufferSize: upgraderBufferSize,
	}

	a := server.New(upgrader, server.NewHub(*broadcast))

	router := mux.NewRouter()
	router.HandleFunc("/ws", a.ServeWs).Methods(http.MethodGet)
	a.Start()

	log.Printf("listening on %s", *addr)

	return http.ListenAndServe(*addr, router)
}
