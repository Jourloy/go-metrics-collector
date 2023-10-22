package handlers

import (
	"github.com/Jourloy/go-metrics-collector/internal/server/app"
	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/gin-gonic/gin"
)

// Registers the app handler in the specified gin.Engine and uses the provided storage.
func RegisterAppHandler(g *gin.RouterGroup, s storage.Storage) {
	appService := app.GetAppSevice(s)

	g.GET(`/live`, appService.Live)
	g.GET(`/`, appService.GetAllMetrics)

	g.POST(`/value`, appService.GetMetricByBody)
	g.GET(`/value/:type/:name/`, appService.GetMetricByParams)

	g.POST(`/update`, appService.UpdateMetricByBody)
	g.POST(`/update/:type/:name/:value`, appService.UpdateMetricByParams)
}
