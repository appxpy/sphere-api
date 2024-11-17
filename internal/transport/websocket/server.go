package websocket

import (
	"net/http"

	"github.com/appxpy/sphere-api/internal/storage"
	"github.com/appxpy/sphere-api/internal/usecases"
)

type Server struct {
	handler *Handler
}

func NewServer() *Server {
	repo := storage.NewClientRepository()
	geoUsecase := usecases.NewGeolocationUsecase(repo)
	usersUsecase := usecases.NewUsersUsecase(repo)
	handler := NewHandler(geoUsecase, usersUsecase)
	return &Server{handler: handler}
}

func (s *Server) Start() {
	http.HandleFunc("/ws", s.handler.HandleWS)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
