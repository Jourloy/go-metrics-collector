package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type gzipBodyWriter struct {
	gin.ResponseWriter
	Body *bytes.Buffer
}

// Write writes the given byte slice to the response body.
//
// Parameters:
//   - b: the byte slice to be written
func (r gzipBodyWriter) Write(b []byte) (int, error) {
	return r.Body.Write(b)
}

// GzipDecode is a middleware function that compresses and decompresses gzipped request and response bodies.
//
// Parameters:
//   - ctx: the gin context.
//
// Returns:
// - a gin.HandlerFunc
func GzipDecode() gin.HandlerFunc {
	return func(c *gin.Context) {
		writer := gzipBodyWriter{
			ResponseWriter: c.Writer,
			Body:           &bytes.Buffer{},
		}
		c.Writer = writer

		b, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}

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
			writer.ResponseWriter.Header().Set(`Content-Encoding`, `gzip`)

			originalBody := writer.Body.Bytes()
			writer.Body.Reset()

			var gz bytes.Buffer

			w := gzip.NewWriter(&gz)
			w.Write(originalBody)
			w.Close()

			writer.ResponseWriter.Write(gz.Bytes())
		} else {
			writer.ResponseWriter.Write(writer.Body.Bytes())
		}
	}
}
