# WebSockets Go Example: Pub/Sub server

## Getting started

> Prerequisites: go 1.15, make must be installed.

Run server:

```shell
go run . server --broadcast 100ms
```

Establish websocket client connections:

```shell
go run . client --clients 5000
```

## Server

Server features:

- Accept HTTP request `{"command": "SUBSCRIBE"}` to `http://localhost:8080/ws` and upgrade to websocket connection.
- Every subscribed client receives broadcast message `{"client_id": "ID", "timestamp": UNIX_SECONDS}` every 100 ms.
- Accept request `{"command": "UNSUBSCRIBE"}` and terminate websocket connection.
- Accept request `{"command": "NUM_CONNECTIONS"}` and return number of active connections
  `{"num_connections": 4895}`.

## Client

Client do:

- Create 5000 websocket connections to server.
- Stdout broadcast messages from the server.
- Request current number of connections.
- Stdout current number of connections to server.
- Unsubscribe one connection from the server.
- Stdout current number of connections to server.

## Development

Build:

```shell
make build
```

Run unit tests:

```shell
make test
```

Run linters:

```shell
make lint
```

Regenerate mocks:

```shell
make generate
```

Run server with `-race` option:

```shell
make server
```

Run client with `-race` option:

```shell
make client
```
