package models

type Response[ResponseType any] struct {
	Type     string        `json:"type"`
	Response *ResponseType `json:"data"`
}

type GetClientInfoRequest struct {
	ClientID string `json:"client_id"`
}

type GetClientsResponse struct {
	Clients []*ClientInfo `json:"clients"`
}

type WhoAmIResponse struct {
	ClientID string `json:"client_id"`
}

type UpdatePositionRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type GetNearestClientResponse struct {
	ID       string  `json:"id"`
	Azimuth  float64 `json:"azimuth"`
	Distance float64 `json:"distance"`
}
