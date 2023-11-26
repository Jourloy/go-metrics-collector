package middlewares

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// logger is a middleware function that logs the details of each incoming request.
//
// It takes a gin.Context as a parameter and returns a gin.HandlerFunc.
// The gin.HandlerFunc is a function that handles the request and response flow.
//
// The function logs the following details:
// - The HTTP method of the request.
// - The status code of the response.
// - The size of the response.
// - The path of the request.
// - The latency of the request.
//
// Parameters:
//   - ctx: the gin context.
//
// Returns:
// - a gin.HandlerFunc
func Logger() gin.HandlerFunc {
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
