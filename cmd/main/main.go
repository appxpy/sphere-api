package main

import (
	"github.com/appxpy/sphere-api/internal/transport/websocket"
)

func main() {
	server := websocket.NewServer()
	server.Start()
}
