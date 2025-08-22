package admin

import "context"

type Service interface {
	GetDriversStatus(ctx context.Context) ([]*DriverStatusDTO, error)
}
