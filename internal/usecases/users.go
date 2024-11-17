package usecases

import (
	"github.com/appxpy/sphere-api/internal/logging"
	"github.com/appxpy/sphere-api/internal/models"
	"github.com/appxpy/sphere-api/internal/storage"
	"github.com/appxpy/sphere-api/internal/util"
	"github.com/gorilla/websocket"
)

type UsersUsecase struct {
	repo *storage.ClientRepository
}

func NewUsersUsecase(repo *storage.ClientRepository) *UsersUsecase {
	return &UsersUsecase{repo: repo}
}

func (u *UsersUsecase) AddClient(client *models.ClientInfo) {
	u.repo.AddClient(client)
	logging.InfoLogger.Printf("Client added: %s", client.ID)
}

func (u *UsersUsecase) RemoveClient(clientID string) {
	u.repo.RemoveClient(clientID)
	logging.InfoLogger.Printf("Client removed: %s", clientID)
}

func (u *UsersUsecase) GetClientInfo(clientID string) (*models.ClientInfo, error) {
	client, exists := u.repo.GetClient(clientID)
	if !exists {
		logging.ErrorLogger.Printf("Error getting client info: %v", util.ErrClientNotFound)
		return nil, util.ErrClientNotFound
	}
	return client, nil
}

func (u *UsersUsecase) GetClientIDByConnection(conn *websocket.Conn) (string, error) {
	id, exists := u.repo.GetClientIDByConnection(conn)
	if !exists {
		logging.ErrorLogger.Printf("Error getting client id by connection: %v", util.ErrClientNotFound)
		return "", util.ErrClientNotFound
	}

	return id, nil
}

func (u *UsersUsecase) GetClients() []*models.ClientInfo {
	return u.repo.GetAllClients()
}
