package websocket

import (
	"math/rand"
	"net/http"

	"github.com/appxpy/sphere-api/internal/logging"
	"github.com/appxpy/sphere-api/internal/models"
	"github.com/appxpy/sphere-api/internal/transport/websocket/api"
	"github.com/appxpy/sphere-api/internal/usecases"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Handler struct {
	geoUsecase   *usecases.GeolocationUsecase
	usersUsecase *usecases.UsersUsecase
	upgrader     websocket.Upgrader

	geolocationAPI *api.GeolocationWebsocketAPI

	state  int
	router *Router
}

func NewHandler(geoUsecase *usecases.GeolocationUsecase, usersUsecase *usecases.UsersUsecase) *Handler {
	handler := &Handler{
		geoUsecase:   geoUsecase,
		usersUsecase: usersUsecase,
		state:        0,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		geolocationAPI: api.NewGeolocationWebsocketAPI(geoUsecase, usersUsecase),
		router:         NewRouter(),
	}

	users := api.NewUsersWebsocketAPI(usersUsecase)

	// Users API
	handler.router.Handle("WhoAmIRequest", users.HandleWhoAmI)
	handler.router.Handle("GetClientsRequest", users.HandleGetClients)
	handler.router.Handle("GetClientInfoRequest", users.HandleGetClientInfo)

	// Geolocation API
	handler.router.Handle("UpdatePositionRequest", handler.geolocationAPI.HandleUpdatePosition)

	return handler
}

func (h *Handler) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logging.ErrorLogger.Printf("Failed to upgrade connection: %v", err)
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	clientID := uuid.New().String()
	sphereID := rand.Intn(511) + 1

	client := &models.ClientInfo{Connection: conn, ID: clientID, SphereID: sphereID}
	h.usersUsecase.AddClient(client)

	for {
		_, msg, errInner := conn.ReadMessage()
		if errInner != nil {
			logging.ErrorLogger.Printf("Error reading message: %v", errInner)
			h.usersUsecase.RemoveClient(clientID)
			notify := h.geoUsecase.UpdateRelatedClients(clientID)
			h.geoUsecase.DeleteClientFromNearestReferences(clientID)
			h.geolocationAPI.NotifyAboutChangedNearestClient(notify)
			break
		}
		if err := h.router.Route(conn, msg); err != nil {
			logging.ErrorLogger.Printf("Error routing message: %v\nMessage: %v", err, string(msg))
		}
	}
}
