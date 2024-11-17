package util

import (
	"errors"

	"github.com/appxpy/sphere-api/internal/models"
)

var (
	OK                    = map[string]interface{}{"status": "ok"}
	ErrClientNotFound     = errors.New("client not found")
	ErrInvalidMessage     = errors.New("invalid message format")
	ErrNoClientsAvailable = errors.New("no clients available")
	ErrNoPositionProvided = errors.New("no position provided for client")
)

func ErrorToInterface(err error) *models.Response[struct {
	Error string `json:"error"`
}] {
	return &models.Response[struct {
		Error string `json:"error"`
	}]{
		Type: "error",
		Response: &struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		},
	}
}
