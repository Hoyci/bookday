package admin

import "time"

type DriverStatusDTO struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Email        string        `json:"email"`
	CurrentRoute *RouteInfoDTO `json:"current_route,omitempty"`
}

type RouteInfoDTO struct {
	ID     string        `json:"id"`
	Status string        `json:"status"`
	Stops  []StopInfoDTO `json:"stops"`
}

type StopInfoDTO struct {
	Sequence  int       `json:"sequence"`
	Address   string    `json:"address"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}
