package main

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"
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
		app.clients = append(app.clients, NewClient(i))
	}

	return app
}

func (a *App) Run(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
		wg := &sync.WaitGroup{}

		for i, client := range a.clients {
			i := i
			client := client

			wg.Add(1)

			go func() {
				defer wg.Done()

				if err := client.Subscribe(ctx, a.server); err != nil {
					log.Printf("client %d fails to connect", i)
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
}
