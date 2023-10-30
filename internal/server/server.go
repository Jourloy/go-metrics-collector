package server

import (
	"bytes"
	"compress/gzip"
	"flag"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
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

	// Set middlewares
	r.Use(logger())         // Logger
	r.Use(gzipMiddleware()) // Gzip

	// Load HTML templates
	r.LoadHTMLGlob("templates/*")

	// Initiate router groups
	appGroup := r.Group(`/`)

	// Register application, collector, and value handlers
	handlers.RegisterAppHandler(appGroup, s)

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

		// Log body for debug purposes
		body, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		zap.L().Debug(
			`Request body`,
			zap.ByteString(`body`, body),
		)

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

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func gzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rbw := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = rbw

		// If content encoding is gzip, decompress the response body
		if c.Request.Header.Get(`Content-Encoding`) == `gzip` {
			b, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.Next()
				return
			}
			defer c.Request.Body.Close()

			r, err := gzip.NewReader(bytes.NewReader(b))
			if err != nil {
				c.Next()
				return
			}

			b, err = io.ReadAll(r)
			if err != nil {
				c.Next()
				return
			}

			c.Request.Body = io.NopCloser(strings.NewReader(string(b)))
		}

		// Perform request
		c.Next()

		// If client accepts gzip then compress the response
		if c.Request.Header.Get(`Accept-Encoding`) == `gzip` {
			rbw.ResponseWriter.Header().Set(`Content-Encoding`, `gzip`)

			originalBody := rbw.body.Bytes()
			rbw.body.Reset()

			var gz bytes.Buffer

			w := gzip.NewWriter(&gz)
			w.Write(originalBody)
			w.Close()

			rbw.ResponseWriter.Write(gz.Bytes())

		} else {
			rbw.ResponseWriter.Write(rbw.body.Bytes())
		}
	}
}
