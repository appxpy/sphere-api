package models

import (
	"math"

	"github.com/dhconnelly/rtreego"
	"github.com/gorilla/websocket"
)

// EarthRadius - Радиус Земли в метрах
const EarthRadius = 6371000

type ClientInfo struct {
	Connection     *websocket.Conn `json:"-"`
	ID             string          `json:"client_id"`
	SphereID       int             `json:"sphere_id"`
	Position       *Position       `json:"position,omitempty"`
	WindowSettings *WindowSettings `json:"window_settings,omitempty"`
}

func (c *ClientInfo) HasPosition() bool {
	return c.Position != nil
}

// Bounds - Реализуем интерфейс rtreego.Spatial для использования в R-Tree
func (c *ClientInfo) Bounds() rtreego.Rect {
	if c.Position == nil {
		// Если позиция не задана, возвращаем прямоугольник нулевого размера
		p := rtreego.Point{0, 0, 0}
		rect, _ := rtreego.NewRect(p, []float64{0, 0, 0})
		return rect
	}

	p := rtreego.Point{c.Position.X, c.Position.Y, c.Position.Z}

	// Создаем прямоугольник с очень малой площадью
	rect, _ := rtreego.NewRect(p, []float64{.000001, .000001, .000001})
	return rect
}

type Position struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`

	ClosestClientID string  `json:"closest_client_id,omitempty"`
	Distance        float64 `json:"distance,omitempty"`
	Azimuth         float64 `json:"azimuth,omitempty"`
}

// SameLocation - Метод для проверки является ли позиция той же самой (не учитывает X, Y, Z и Heading)
func (p *Position) SameLocation(another *Position) bool {
	return p.Latitude == another.Latitude && p.Longitude == another.Longitude
}

// UpdateXYZ - Метод для обновления X, Y, Z координат на основе широты и долготы
func (p *Position) UpdateXYZ() {
	latRad := p.Latitude * math.Pi / 180
	lonRad := p.Longitude * math.Pi / 180

	p.X = EarthRadius * math.Cos(latRad) * math.Cos(lonRad)
	p.Y = EarthRadius * math.Cos(latRad) * math.Sin(lonRad)
	p.Z = EarthRadius * math.Sin(latRad)
}

type WindowSettings struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}
