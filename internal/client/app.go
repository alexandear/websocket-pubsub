package client

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	gws "github.com/gorilla/websocket"

	"github.com/alexandear/websocket-pubsub/internal/pkg/websocket"
)

const pauseBetweenCommands = 2 * time.Second

type App struct {
	server string

	clients []*Client
}

func NewApp(server string, numClients int) *App {
	app := &App{
		server:  server,
		clients: make([]*Client, 0, numClients),
	}

	for i := 1; i <= numClients; i++ {
		app.clients = append(app.clients, NewClient())
	}

	return app
}

func (a *App) Run(ctx context.Context) {
	wg := &sync.WaitGroup{}

	for i, client := range a.clients {
		i := i
		client := client

		wg.Add(1)

		go func() {
			defer wg.Done()

			conn, _, err := gws.DefaultDialer.DialContext(ctx, "ws://"+a.server+"/ws", nil)
			if err != nil {
				log.Printf("dial failed: %v", err)

				return
			}

			client.SetConn(websocket.NewConn(conn))

			if err := client.Subscribe(); err != nil {
				log.Printf("client %d fails to connect: %v", i, err)
			}
		}()
	}

	wg.Wait()

	for _, client := range a.clients {
		go client.Read()
	}

	if err := a.clients[rand.Intn(len(a.clients))].NumConnections(); err != nil {
		log.Printf("num connections failed: %v", err)
	}

	time.Sleep(pauseBetweenCommands)

	if err := a.clients[rand.Intn(len(a.clients))].Unsubscribe(); err != nil {
		log.Printf("unsubscribe failed: %v", err)
	}

	time.Sleep(pauseBetweenCommands)

	if err := a.clients[rand.Intn(len(a.clients))].NumConnections(); err != nil {
		log.Printf("num connections failed: %v", err)
	}

	time.Sleep(pauseBetweenCommands)
}
