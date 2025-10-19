package service

//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock

import (
	"context"

	"github.com/tmizuma/japan-tech-careers-api/apps/api-server/internal/domain/model"
	"github.com/tmizuma/japan-tech-careers-api/apps/api-server/internal/infra/httpclient"
	"github.com/tmizuma/japan-tech-careers-api/apps/api-server/internal/shared/logger"
)

// Service is the interface for business logic
type Service interface {
	FetchJobs(ctx context.Context) ([]model.Job, error)
}

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	httpClient httpclient.HttpClient
}

// NewServiceImpl creates a new ServiceImpl
func NewServiceImpl(httpClient httpclient.HttpClient) Service {
	return &ServiceImpl{
		httpClient: httpClient,
	}
}

// FetchJobs fetches jobs using the HTTP client
func (s *ServiceImpl) FetchJobs(ctx context.Context) ([]model.Job, error) {
	logger.Info(ctx, "Fetching jobs from external API")

	jobs, err := s.httpClient.GetJobs(ctx)
	if err != nil {
		logger.Error(ctx, "Failed to fetch jobs from external API")
		return nil, err
	}

	logger.Info(ctx, "Successfully fetched jobs from external API")
	return jobs, nil
}
