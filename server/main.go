package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	a := &App{
		router: mux.NewRouter(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		hub: NewHub(),
	}
	a.Initialize()

	log.Fatal(http.ListenAndServe(*addr, a.router))
}
