package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	upgraderBufferSize = 1024
)

func main() {
	addr := flag.String("addr", ":8080", "http service address")

	flag.Parse()

	a := &App{
		router: mux.NewRouter(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  upgraderBufferSize,
			WriteBufferSize: upgraderBufferSize,
		},
		hub: NewHub(),
	}
	a.Initialize()

	log.Printf("listening on %s", *addr)

	log.Fatal(http.ListenAndServe(*addr, a.router))
}
