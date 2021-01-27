package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func main() {
	addr := flag.String("addr", ":8080", "http service address")

	flag.Parse()

	a := &App{
		router: mux.NewRouter(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		hub: NewHub(),
	}
	a.Initialize()

	log.Printf("listening on %s", *addr)

	log.Fatal(http.ListenAndServe(*addr, a.router))
}
