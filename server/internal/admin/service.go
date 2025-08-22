// internal/admin/service.go

package admin

import (
	"context"
	"errors"

	"github.com/charmbracelet/log"
	"github.com/hoyci/bookday/internal/auth"
	models "github.com/hoyci/bookday/internal/infra/database/model"
	"github.com/hoyci/bookday/internal/routing"
	"github.com/hoyci/bookday/pkg/fault"
)

type service struct {
	authRepo    auth.Repository
	routingRepo routing.Repository
	log         *log.Logger
}

func NewService(authRepo auth.Repository, routingRepo routing.Repository, logger *log.Logger) Service {
	return &service{
		authRepo:    authRepo,
		routingRepo: routingRepo,
		log:         logger,
	}
}

func (s *service) GetDriversStatus(ctx context.Context) ([]*DriverStatusDTO, error) {
	s.log.Info("fetching status for all drivers")

	drivers, err := s.authRepo.FindUsersByRole(ctx, models.RoleDriver)
	if err != nil {
		s.log.Error("failed to fetch drivers from repository", "error", err)
		return nil, err
	}

	var statuses []*DriverStatusDTO
	for _, driver := range drivers {
		statusDTO := &DriverStatusDTO{
			ID:    driver.ID(),
			Name:  driver.Name(),
			Email: driver.Email(),
		}

		activeRoute, err := s.routingRepo.FindActiveRouteByDriverID(ctx, driver.ID())
		if err != nil {
			var f *fault.Error
			if errors.As(err, &f) && f.Kind == fault.KindNotFound {
			} else {
				s.log.Error("failed to check active route for driver", "driver_id", driver.ID(), "error", err)
			}
		}

		if activeRoute != nil {
			routeInfo := &RouteInfoDTO{
				ID:     activeRoute.ID(),
				Status: string(activeRoute.Status()),
			}

			for _, stop := range activeRoute.Stops() {
				routeInfo.Stops = append(routeInfo.Stops, StopInfoDTO{
					Sequence:  stop.Sequence(),
					Address:   stop.Address(),
					Latitude:  stop.Latitude(),
					Longitude: stop.Longitude(),
					Status:    string(stop.Status()),
					UpdatedAt: stop.UpdatedAt(),
				})
			}
			statusDTO.CurrentRoute = routeInfo
		}

		statuses = append(statuses, statusDTO)
	}

	return statuses, nil
}
