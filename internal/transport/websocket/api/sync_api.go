package api

import (
	"encoding/json"

	"github.com/appxpy/sphere-api/internal/logging"
	"github.com/appxpy/sphere-api/internal/models"
	"github.com/appxpy/sphere-api/internal/usecases"
	"github.com/appxpy/sphere-api/internal/util"
	"github.com/gorilla/websocket"
)

type SyncWebsocketAPI struct {
	usersUsecase *usecases.UsersUsecase
	geoUsecase   *usecases.GeolocationUsecase
}

func NewSyncWebsocketAPI(usersUsecase *usecases.UsersUsecase, geoUsecase *usecases.GeolocationUsecase) *SyncWebsocketAPI {
	return &SyncWebsocketAPI{
		usersUsecase: usersUsecase,
		geoUsecase:   geoUsecase,
	}
}

func (api *SyncWebsocketAPI) HandleSyncStateMessage(conn *websocket.Conn, data json.RawMessage) {
	// Parse SyncStateMessage
	var message models.SyncStateMessage
	if err := json.Unmarshal(data, &message); err != nil {
		conn.WriteJSON(util.ErrorToInterface(err))
		return
	}

	// Get sender's client_id
	senderID, err := api.usersUsecase.GetClientIDByConnection(conn)
	if err != nil {
		conn.WriteJSON(util.ErrorToInterface(err))
		return
	}

	// Prepare SyncStateMessage to be sent
	response := &models.Response[models.SyncStateMessage]{
		Type:     "SyncStateResponse",
		Response: &message,
	}

	// Find clients who have the sender as their nearest client
	clientsReferencingSender := api.geoUsecase.GetClientsWhoReferenceClientAsNearest(senderID)
	for _, clientID := range clientsReferencingSender {
		client, err := api.usersUsecase.GetClientInfo(clientID)
		if err == nil && client.Connection != nil {
			if err := client.Connection.WriteJSON(response); err != nil {
				logging.ErrorLogger.Printf("Error sending SyncStateResponse to client %s: %v", client.ID, err)
			}
		}
	}
}
