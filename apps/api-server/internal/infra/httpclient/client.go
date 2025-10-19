package httpclient

//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/tmizuma/japan-tech-careers-api/apps/api-server/config"
	"github.com/tmizuma/japan-tech-careers-api/apps/api-server/internal/domain/model"
)

// HttpClient is an interface for external API calls
type HttpClient interface {
	GetJobs(ctx context.Context) ([]model.Job, error)
}

// ClientImpl is the implementation of HttpClient
type ClientImpl struct {
	Endpoint   string
	HTTPClient *http.Client
}

// New creates a new HttpClient implementation
func New(cfg *config.Config) HttpClient {
	return &ClientImpl{
		Endpoint: cfg.ApiEndpoint,
		HTTPClient: &http.Client{
			Timeout: time.Duration(cfg.ApiTimeout) * time.Second,
		},
	}
}

// GetJobs fetches jobs from external API (dummy implementation)
func (c *ClientImpl) GetJobs(ctx context.Context) ([]model.Job, error) {
	// This is a dummy implementation
	// In real implementation, you would make HTTP request to c.Endpoint
	jobs := []model.Job{
		{
			ID:          "1",
			Title:       "Senior Go Developer",
			Company:     "Tech Company A",
			Location:    "Tokyo, Japan",
			Description: "Looking for an experienced Go developer",
		},
		{
			ID:          "2",
			Title:       "Backend Engineer",
			Company:     "Startup B",
			Location:    "Osaka, Japan",
			Description: "Join our growing team",
		},
	}

	fmt.Printf("Fetching jobs from: %s\n", c.Endpoint)
	return jobs, nil
}
