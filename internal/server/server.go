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
	"go.uber.org/zap"
)

var (
	Host = flag.String(`a`, `localhost:8080`, `Host of the server`)
)

// Start initiates the application.
func Start() {
	// Initiate handlers
	r := gin.New()

	// Set middlewares
	r.Use(gin.Recovery())
	r.Use(logger())         // Logger
	r.Use(gzipMiddleware()) // Gzip
	// Check if ADDRESS environment variable is set and assign it to Host
	if hostENV, exist := os.LookupEnv(`ADDRESS`); exist {
		Host = &hostENV
	}

	flag.Parse()

	s := repository.CreateRepository()

	// Load HTML templates
	r.LoadHTMLGlob(`templates/*`)

	// Initiate router groups
	appGroup := r.Group(`/`)

	// Register application, collector, and value handlers
	handlers.RegisterAppHandler(appGroup, s)

	// Run the server on the specified host
	if err := r.Run(*Host); err != nil {
		// Handle server closed error
		if err == http.ErrServerClosed {
			zap.L().Info(`Server closed`)
		} else {
			panic(err)
		}
	}
}

// logger is a middleware function that logs the details of each incoming request.
//
// It takes a gin.Context as a parameter and returns a gin.HandlerFunc.
// The gin.HandlerFunc is a function that handles the request and response flow.
//
// The function logs the following details:
// - The HTTP method of the request.
// - The status code of the response.
// - The size of the response (if the method is GET).
// - The path of the request.
// - The latency of the request.
//
// Parameters:
//   - ctx: the gin context.
//
// Returns:
// - a gin.HandlerFunc
func logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		c.Next()

		latency := time.Since(t)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		responseSize := c.Writer.Size()

		zap.L().Info(
			method,
			zap.String(`status`, strconv.Itoa(status)),
			zap.Int(`size`, responseSize),
			zap.String(`path`, path),
			zap.Duration(`latency`, latency),
		)
	}
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write writes the given byte slice to the response body.
//
// Parameters:
//   - b: the byte slice to be written
func (r responseBodyWriter) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

// gzipMiddleware is a middleware function that compresses and decompresses gzipped request and response bodies.
//
// Parameters:
//   - ctx: the gin context.
//
// Returns:
// - a gin.HandlerFunc
func gzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rbw := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = rbw

		b, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}
		defer c.Request.Body.Close()

		// If content encoding is gzip, decompress the response body
		if c.Request.Header.Get(`Content-Encoding`) == `gzip` {
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
		}

		// Set request body
		c.Request.Body = io.NopCloser(strings.NewReader(string(b)))

		// Log request body
		zap.L().Debug(`Request body:`)
		zap.L().Debug(string(b))

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
