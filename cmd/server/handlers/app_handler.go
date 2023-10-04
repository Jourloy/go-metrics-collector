package handlers

import (
	"net/http"

	"github.com/Jourloy/go-metrics-collector/cmd/server/app"
	"github.com/Jourloy/go-metrics-collector/cmd/server/storage"
	"github.com/gin-gonic/gin"
)

func live(c *gin.Context) {
	c.String(http.StatusOK, "Live")
}

func RegisterAppHandler(r *gin.Engine, s storage.Storage) {
	appHandler := app.GetAppSevice(s)

	app := r.Group(`/`)
	{
		app.GET(`/`, appHandler.ServeHTTP)
		app.GET(`/live`, live)
	}
}
