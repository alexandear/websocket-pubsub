# WebSockets Go Example: Pub/Sub server

> Prerequisites: go 1.15, make must be installed.

Run server:

```shell
make run-server
```

Establish websocket client connections:

```shell
make run-client
```

## Server

Server possibilities:

- Accept HTTP request `{"command": "SUBSCRIBE"}` to `http://localhost:8080/pubsub` and redirect to websocket
  connection `ws://localhost:8080/ws`.
- Every subscribed client receives broadcast message `{"client_id": "ID", "timestamp": "timestamp"}` every 2 s.
- Accept request `{"command": "UNSUBSCRIBE"}` and terminate websocket connection.
- Accept request `{"command": "NUM_CONNECTIONS"}` and return number of active connections
  `{"num_connections": 4895}`.

## Client

Client do:

- Create websocket connection to server.
- Request current number of connections.
- Write to screen broadcast message from the server.
- Unsubscribe from the server.
