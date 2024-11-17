package websocket

import (
	"encoding/json"

	"github.com/appxpy/sphere-api/internal/util"
	"github.com/gorilla/websocket"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Router struct {
	routes map[string]func(*websocket.Conn, json.RawMessage)
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]func(*websocket.Conn, json.RawMessage)),
	}
}

func (r *Router) Handle(messageType string, handler func(*websocket.Conn, json.RawMessage)) {
	r.routes[messageType] = handler
}

func (r *Router) Route(conn *websocket.Conn, msg []byte) error {
	var message Message
	if err := json.Unmarshal(msg, &message); err != nil {
		return err
	}

	handler, found := r.routes[message.Type]
	if !found {
		return util.ErrInvalidMessage
	}

	handler(conn, message.Data)
	return nil
}
