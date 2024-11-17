package api

import (
	"encoding/json"

	"github.com/appxpy/sphere-api/internal/logging"
	"github.com/appxpy/sphere-api/internal/models"
	"github.com/appxpy/sphere-api/internal/usecases"
	"github.com/appxpy/sphere-api/internal/util"
	"github.com/gorilla/websocket"
)

type GeolocationWebsocketAPI struct {
	usersUsecase *usecases.UsersUsecase
	geoUsecase   *usecases.GeolocationUsecase
}

func NewGeolocationWebsocketAPI(geoUsecase *usecases.GeolocationUsecase, usersUsecase *usecases.UsersUsecase) *GeolocationWebsocketAPI {
	return &GeolocationWebsocketAPI{geoUsecase: geoUsecase, usersUsecase: usersUsecase}
}

func (api *GeolocationWebsocketAPI) HandleUpdatePosition(conn *websocket.Conn, data json.RawMessage) {
	var request models.UpdatePositionRequest
	if err := json.Unmarshal(data, &request); err != nil {
		conn.WriteJSON(util.ErrorToInterface(err))
		return
	}

	clientID, err := api.usersUsecase.GetClientIDByConnection(conn) // Implement this function to retrieve the client ID
	if err != nil {
		conn.WriteJSON(util.ErrorToInterface(err))
	}

	position := &models.Position{
		Latitude:  request.Latitude,
		Longitude: request.Longitude,
	}

	notify := api.geoUsecase.UpdatePosition(clientID, position.Latitude, position.Longitude)
	logging.InfoLogger.Printf("Client %s updated position to %f, %f, notifying %v", clientID, position.Latitude, position.Longitude, notify)
	api.NotifyAboutChangedNearestClient(notify)
}

func (api *GeolocationWebsocketAPI) NotifyAboutChangedNearestClient(notify []string) {
	// Notify clients that their target position changed
	for _, recieverID := range notify {
		reciever, err := api.usersUsecase.GetClientInfo(recieverID)
		if err != nil {
			continue
		}

		response := &models.GetNearestClientResponse{
			ID:       reciever.Position.ClosestClientID,
			Azimuth:  reciever.Position.Azimuth,
			Distance: reciever.Position.Distance,
		}

		logging.InfoLogger.Printf("Sending new target position to client %s", recieverID)
		reciever.Connection.WriteJSON(&models.Response[models.GetNearestClientResponse]{
			Type:     "GetNearestClientResponse",
			Response: response,
		})
	}
}
