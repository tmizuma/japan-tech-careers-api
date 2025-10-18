package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
	chiRouter "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var chiLambda *chiadapter.ChiLambda

func init() {
	r := chiRouter.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Routes
	r.Get("/", handleRoot)

	chiLambda = chiadapter.New(r)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"message": "hello",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		// Running in Lambda
		lambda.Start(chiLambda.ProxyWithContext)
	} else {
		// Running locally
		r := chiRouter.NewRouter()
		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer)
		r.Get("/", handleRoot)

		http.ListenAndServe(":8080", r)
	}
}
