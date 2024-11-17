package api

import (
	"encoding/json"

	"github.com/appxpy/sphere-api/internal/models"
	"github.com/appxpy/sphere-api/internal/usecases"
	"github.com/appxpy/sphere-api/internal/util"
	"github.com/gorilla/websocket"
)

type UsersWebsocketAPI struct {
	usersUsecase *usecases.UsersUsecase
}

func NewUsersWebsocketAPI(usersUsecase *usecases.UsersUsecase) *UsersWebsocketAPI {
	return &UsersWebsocketAPI{usersUsecase: usersUsecase}
}

func (api *UsersWebsocketAPI) HandleWhoAmI(conn *websocket.Conn, data json.RawMessage) {
	clientID, err := api.usersUsecase.GetClientIDByConnection(conn)
	if err != nil {
		conn.WriteJSON(util.ErrorToInterface(err))
		return
	}

	response, err := json.Marshal(&models.Response[models.WhoAmIResponse]{
		Type:     "WhoAmIResponse",
		Response: &models.WhoAmIResponse{ClientID: clientID},
	})

	if err != nil {
		conn.WriteJSON(util.ErrorToInterface(err))
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, response); err != nil {
		conn.WriteJSON(util.ErrorToInterface(err))
	}
}

func (api *UsersWebsocketAPI) HandleGetClients(conn *websocket.Conn, data json.RawMessage) {
	clients := api.usersUsecase.GetClients()

	response, err := json.Marshal(&models.Response[models.GetClientsResponse]{
		Type:     "GetClientsResponse",
		Response: &models.GetClientsResponse{Clients: clients},
	})

	if err != nil {
		conn.WriteJSON(util.ErrorToInterface(err))
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, response); err != nil {
		conn.WriteJSON(util.ErrorToInterface(err))
	}
}

func (api *UsersWebsocketAPI) HandleGetClientInfo(conn *websocket.Conn, data json.RawMessage) {
	var request models.GetClientInfoRequest
	if err := json.Unmarshal(data, &request); err != nil {
		conn.WriteJSON(util.ErrorToInterface(err))
		return
	}

	clientInfo, err := api.usersUsecase.GetClientInfo(request.ClientID)
	if err != nil {
		conn.WriteJSON(util.ErrorToInterface(err))
		return
	}

	response, err := json.Marshal(&models.Response[models.ClientInfo]{
		Type:     "GetClientInfoResponse",
		Response: clientInfo,
	})

	if err != nil {
		conn.WriteJSON(util.ErrorToInterface(err))
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, response); err != nil {
		conn.WriteJSON(util.ErrorToInterface(err))
	}
}
