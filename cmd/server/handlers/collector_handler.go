package handlers

import (
	"fmt"

	"github.com/Jourloy/go-metrics-collector/cmd/server/collector"
	"github.com/Jourloy/go-metrics-collector/cmd/server/storage"
	"github.com/gin-gonic/gin"
)

func RegisterCollectorHandler(r *gin.Engine, s *storage.Storage) {
	// Prepare for .env
	metricEndpoint := `/update`

	collectorHandler := collector.CollectMetric(*s)

	r.POST(metricEndpoint+`/:type`+`/:name`+`/:value`, collectorHandler.ServeHTTP)

	fmt.Println(`Mapped`, metricEndpoint)
}
