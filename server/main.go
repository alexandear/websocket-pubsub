package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	upgraderBufferSize = 1024

	defaultBroadcast = 100 * time.Millisecond
)

func main() {
	addr := flag.String("addr", ":8080", "http service address")
	broadcast := flag.Duration("broadcast", defaultBroadcast, "broadcast frequency")

	flag.Parse()

	a := &App{
		router: mux.NewRouter(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  upgraderBufferSize,
			WriteBufferSize: upgraderBufferSize,
		},
		hub: NewHub(*broadcast),
	}
	a.Initialize()

	log.Printf("listening on %s", *addr)

	log.Fatal(http.ListenAndServe(*addr, a.router))
}
