package controller

//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock

import (
	"context"

	"github.com/tamizuma/japan-tech-careers-api/apps/api-server/internal/domain/model"
	"github.com/tamizuma/japan-tech-careers-api/apps/api-server/internal/domain/service"
	"github.com/tamizuma/japan-tech-careers-api/apps/api-server/internal/shared/logger"
)

// Controller is the interface for handling business logic coordination
type Controller interface {
	GetJobs(ctx context.Context) ([]model.Job, error)
}

// ControllerImpl implements the Controller interface
type ControllerImpl struct {
	service service.Service
}

// NewController creates a new ControllerImpl
func NewController(svc service.Service) Controller {
	return &ControllerImpl{
		service: svc,
	}
}

// GetJobs handles the job retrieval logic
func (c *ControllerImpl) GetJobs(ctx context.Context) ([]model.Job, error) {
	logger.Info(ctx, "Controller: GetJobs called")

	jobs, err := c.service.FetchJobs(ctx)
	if err != nil {
		logger.Error(ctx, "Controller: Failed to fetch jobs from service")
		return nil, err
	}

	logger.Info(ctx, "Controller: Successfully fetched jobs from service")
	return jobs, nil
}
