package application

import (
	"github.com/tmizuma/japan-tech-careers-api/apps/api-server/config"
	"github.com/tmizuma/japan-tech-careers-api/apps/api-server/internal/domain/service"
	"github.com/tmizuma/japan-tech-careers-api/apps/api-server/internal/infra/controller"
	"github.com/tmizuma/japan-tech-careers-api/apps/api-server/internal/infra/httpclient"
	"github.com/tmizuma/japan-tech-careers-api/apps/api-server/internal/infra/router"
)

// Application holds all dependencies
type Application struct {
	Router     *router.Router
	Config     *config.Config
	Controller controller.Controller
	Service    service.Service
}

// New creates a new Application with all dependencies injected
func New(cfg *config.Config) (*Application, error) {
	// Build dependency chain: config -> httpclient -> service -> controller -> router
	httpClient := httpclient.New(cfg)
	svc := service.NewServiceImpl(httpClient)
	ctrl := controller.NewController(svc)
	r := router.NewRouter(ctrl)

	return &Application{
		Router:     r,
		Config:     cfg,
		Controller: ctrl,
		Service:    svc,
	}, nil
}
