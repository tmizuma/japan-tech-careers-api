package router

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tmizuma/japan-tech-careers-api/apps/api-server/internal/infra/controller"
	"github.com/tmizuma/japan-tech-careers-api/apps/api-server/internal/shared/logger"
)

// Router wraps the chi router with dependencies
type Router struct {
	*chi.Mux
	controller controller.Controller
}

// NewRouter creates a new router with all handlers
func NewRouter(ctrl controller.Controller) *Router {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	router := &Router{
		Mux:        r,
		controller: ctrl,
	}

	// Routes
	r.Get("/", router.handleRoot)
	r.Get("/jobs", router.handleGetJobs)

	return router
}

// handleRoot is a health check endpoint
func (r *Router) handleRoot(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger.Info(ctx, "Health check endpoint called")

	response := map[string]string{
		"message": "Japan Tech Careers API is running",
		"status":  "healthy",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleGetJobs fetches jobs from the controller
func (r *Router) handleGetJobs(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger.Info(ctx, "GET /jobs endpoint called")

	jobs, err := r.controller.GetJobs(ctx)
	if err != nil {
		logger.Error(ctx, "Failed to fetch jobs")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to fetch jobs",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"jobs":  jobs,
		"count": len(jobs),
	})
}
