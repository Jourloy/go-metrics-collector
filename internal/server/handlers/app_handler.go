package handlers

import (
	"github.com/Jourloy/go-metrics-collector/internal/server/app"
	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/gin-gonic/gin"
)

// Registers the app handler in the specified gin.Engine and uses the provided storage.
func RegisterAppHandler(g *gin.RouterGroup, s storage.Storage) {
	appService := app.GetAppSevice(s)

	g.GET(`/ping`, appService.Pong)
	g.GET(`/`, appService.GetAllMetrics)

	// Below code looks ugly, but it is needed to make the handler work.
	//
	// If I use only `/update` and `/update/:type/:name/:value` main logic works fine but
	// Yandex's autotests not, because Yandex's autotests want 400 code if name or value are
	// invalid (empty = invalid), but if name or value are empty, gin will return 404.
	//
	// So I add all possible routes for local process and return 400 if needed.

	g.POST(`/value`, appService.GetMetricByBody)
	g.GET(`/value/:type`, appService.GetMetricByParams)
	g.GET(`/value/:type/:name`, appService.GetMetricByParams)

	g.POST(`/update`, appService.UpdateMetricByBody)
	g.POST(`/update/:type`, appService.UpdateMetricByParams)
	g.POST(`/update/:type/:name`, appService.UpdateMetricByParams)
	g.POST(`/update/:type/:name/:value`, appService.UpdateMetricByParams)

	g.POST(`/updates`, appService.UpdateManyMetrics)
}
