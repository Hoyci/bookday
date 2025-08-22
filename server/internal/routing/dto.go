package routing

import (
	"time"

	v "github.com/go-ozzo/ozzo-validation/v4"
	models "github.com/hoyci/bookday/internal/infra/database/model"
)

type RouteAssociationResponseDTO struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	DriverID string `json:"driver_id"`
}

type OrderSummaryDTO struct {
	ID string `json:"id"`
}

type RouteStopDTO struct {
	ID        string            `json:"id"`
	Sequence  int               `json:"sequence"`
	Address   string            `json:"address"`
	Status    string            `json:"status"`
	Latitude  float64           `json:"latitude"`
	Longitude float64           `json:"longitude"`
	Orders    []OrderSummaryDTO `json:"orders"`
}

type RouteDetailDTO struct {
	ID        string         `json:"id"`
	Status    string         `json:"status"`
	UpdatedAt time.Time      `json:"updated_at"`
	Stops     []RouteStopDTO `json:"stops"`
}

type UpdateStopStatusDTO struct {
	Status string `json:"status"`
}

func (dto UpdateStopStatusDTO) Validate() error {
	return v.ValidateStruct(&dto,
		v.Field(&dto.Status,
			v.Required.Error("status is required"),
			v.In(
				string(models.StopStatusDelivered),
				string(models.StopStatusFailed),
			).Error("invalid status value"),
		),
	)
}
