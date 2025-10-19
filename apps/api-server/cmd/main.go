package main

import (
	"context"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/tamizuma/japan-tech-careers-api/apps/api-server/config"
	"github.com/tamizuma/japan-tech-careers-api/apps/api-server/internal/application"
	"github.com/tamizuma/japan-tech-careers-api/apps/api-server/internal/shared/logger"
)

var chiLambda *chiadapter.ChiLambda

func init() {
	ctx := context.Background()

	// Load configuration
	cfg := config.NewConfig()
	logger.Info(ctx, "Configuration loaded")

	// Initialize application with DI
	app, err := application.New(cfg)
	if err != nil {
		logger.Error(ctx, "Failed to initialize application")
		panic(err)
	}

	logger.Info(ctx, "Application initialized successfully")

	// Create Lambda adapter
	chiLambda = chiadapter.New(app.Router)
}

func main() {
	ctx := context.Background()

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		// Running in Lambda
		logger.Info(ctx, "Starting in Lambda mode")
		lambda.Start(chiLambda.ProxyWithContext)
	} else {
		// Running locally
		logger.Info(ctx, "Starting in local mode on port 8080")

		// Reload config and app for local development
		cfg := config.NewConfig()
		app, err := application.New(cfg)
		if err != nil {
			logger.Error(ctx, "Failed to initialize application")
			panic(err)
		}

		http.ListenAndServe(":8080", app.Router)
	}
}
