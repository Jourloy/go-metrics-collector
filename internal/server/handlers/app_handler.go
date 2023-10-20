package handlers

import (
	"net/http"

	"github.com/Jourloy/go-metrics-collector/internal/server/app"
	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/gin-gonic/gin"
)

// Return an HTTP response with the string "Live".
func live(c *gin.Context) {
	c.String(http.StatusOK, "Live")
}

// Registers the app handler in the specified gin.Engine and uses the provided storage.
func RegisterAppHandler(r *gin.Engine, s storage.Storage) {
	appHandler := app.GetAppSevice(s)

	app := r.Group(`/`)
	{
		app.GET(`/`, appHandler.ServeHTTP)
		app.GET(`/live`, live)
	}
}
