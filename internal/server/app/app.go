package app

import (
	"net/http"

	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/gin-gonic/gin"
)

type AppSevice struct {
	storage storage.Storage
}

// Initializes and returns a new instance of AppSevice.
//
// It takes a `storage.Storage` parameter `s` and returns a pointer to `AppSevice`.
func GetAppSevice(s storage.Storage) *AppSevice {
	return &AppSevice{
		storage: s,
	}
}

// Handles the HTTP request.
//
// It retrieves the values from the storage and merges them into a single map.
// The merged map is then passed to the HTML template for rendering.
func (a *AppSevice) GetAllMetrics(ctx *gin.Context) {
	gauge, counter := a.storage.GetValues()
	merged := make(map[string]any, len(gauge)+len(counter))

	for name, value := range counter {
		merged[name] = value
	}
	for name, value := range gauge {
		merged[name] = value
	}

	ctx.HTML(http.StatusOK, `index.tmpl`, gin.H{
		`merged`: merged,
	})
}

// Return an HTTP response with the string "Live".
func (a *AppSevice) Live(c *gin.Context) {
	c.String(http.StatusOK, "Live")
}
