package server

import (
	"flag"
	"net/http"
	"os"

	"github.com/Jourloy/go-metrics-collector/internal/server/handlers"
	"github.com/Jourloy/go-metrics-collector/internal/server/storage/repository"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var (
	Host = flag.String("a", `localhost:8080`, "Host of the server")
)

// Initialize the application.
func init() {
	if err := godotenv.Load(`.env.server`); err != nil {
		zap.L().Warn(`.env.server not found`)
	}
}

// Start initiates the application.
func Start() {
	// Check if ADDRESS environment variable is set and assign it to Host
	if hostENV, exist := os.LookupEnv("ADDRESS"); exist {
		Host = &hostENV
	}

	flag.Parse()

	s := repository.CreateRepository()

	// Initiate handlers
	r := gin.Default()

	// Load HTML templates
	r.LoadHTMLGlob("templates/*")

	// Register application, collector, and value handlers
	handlers.RegisterAppHandler(r, s)
	handlers.RegisterCollectorHandler(r, s)
	handlers.RegisterValueHandler(r, s)

	// Run the server on the specified host
	if err := r.Run(*Host); err != nil {
		// Handle server closed error
		if err == http.ErrServerClosed {
			zap.L().Info("Server closed")
		} else {
			panic(err)
		}
	}
}
