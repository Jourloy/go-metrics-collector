package handlers

import (
	"github.com/Jourloy/go-metrics-collector/internal/server/collector"
	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/gin-gonic/gin"
)

// Register a collector handler in a gin.Engine instance.
func RegisterCollectorHandler(g *gin.RouterGroup, s storage.Storage) {
	collectorService := collector.CollectMetric(s)

	g.POST(`/`, collectorService.ProcessMetrics)
}
