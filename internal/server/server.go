package server

import (
	"flag"
	"net/http"
	"os"
	"strconv"
	"time"

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
	r := gin.New()

	// Set logger middleware
	r.Use(logger())

	// Load HTML templates
	r.LoadHTMLGlob("templates/*")

	// Initiate router groups
	appGroup := r.Group(`/`)
	collectorGroup := r.Group(`/update`)
	valueGroup := r.Group(`/value`)

	// Register application, collector, and value handlers
	handlers.RegisterAppHandler(appGroup, s)
	handlers.RegisterCollectorHandler(collectorGroup, s)
	handlers.RegisterValueHandler(valueGroup, s)

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

func logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()

		latency := time.Since(t)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		if method == `GET` {
			responseSize := c.Writer.Size()
			zap.L().Info(
				method,
				zap.String(`status`, strconv.Itoa(status)),
				zap.Int(`size`, responseSize),
				zap.String(`path`, path),
				zap.Duration(`latency`, latency),
			)
		} else {
			zap.L().Info(
				method,
				zap.String(`status`, strconv.Itoa(status)),
				zap.String(`path`, path),
				zap.Duration(`latency`, latency),
			)
		}
	}
}
