package middlewares

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	Key = flag.String(`k`, ``, `Key for hash`)
)

func parseEnv() {
	if env, exist := os.LookupEnv(`KEY`); exist {
		Key = &env
	}
}

type hashResponseWriter struct {
	gin.ResponseWriter
	Body *bytes.Buffer
}

// Write writes the given byte slice to the response body.
//
// Parameters:
//   - b: the byte slice to be written
func (r hashResponseWriter) Write(b []byte) (int, error) {
	return r.Body.Write(b)
}

// HashDecode compare the hash of the request body with the hash in the header and return 400 if they don't match.
//
// **Do not use this middleware in production.**
// User can send new data with old hash.
func HashDecode() gin.HandlerFunc {
	parseEnv()

	return func(c *gin.Context) {
		writer := hashResponseWriter{
			ResponseWriter: c.Writer,
			Body:           &bytes.Buffer{},
		}
		c.Writer = writer

		// Check key
		if *Key == `` {
			zap.L().Debug(`Key is empty`)
			c.Next()
			return
		}

		// Check hash
		h := c.Request.Header.Get(`HashSHA256`)
		if h == `` {
			zap.L().Debug(`Header is empty`)
			c.Next()
			return
		}

		// Read body
		b, err := io.ReadAll(c.Request.Body)
		if err != nil {
			zap.L().Debug(`Body is empty`)
			c.Next()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(b))

		key := sha256.Sum256([]byte(*Key))

		// Create cipher block
		aesblock, err := aes.NewCipher(key[:])
		if err != nil {
			zap.L().Error(`Cannot create AES block`, zap.Error(err))
			c.String(400, `bad request`)
		}

		// Create GCM
		aesgcm, err := cipher.NewGCM(aesblock)
		if err != nil {
			zap.L().Error(`Cannot create AES GCM`, zap.Error(err))
			c.String(400, `bad request`)
		}

		nonce := key[len(key)-aesgcm.NonceSize():]

		// Encode boyd and check hash
		compareH := aesgcm.Seal(nil, nonce, b, nil)
		if string(compareH) != h {
			c.String(400, `bad request`)
		}

		zap.L().Debug(`Hash checked`)

		c.Next()

		// Encode body and update header
		newH := aesgcm.Seal(nil, nonce, writer.Body.Bytes(), nil)
		c.Header(`HashSHA256`, hex.EncodeToString(newH[:]))
	}
}
