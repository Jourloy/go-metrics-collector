package handlers

import (
	"github.com/Jourloy/go-metrics-collector/internal/server/collector"
	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/gin-gonic/gin"
)

func RegisterCollectorHandler(r *gin.Engine, s storage.Storage) {
	metricEndpoint := `/update`

	collectorHandler := collector.CollectMetric(s)

	r.POST(metricEndpoint+`/:type/:name/:value`, collectorHandler.ServeHTTP)
}
