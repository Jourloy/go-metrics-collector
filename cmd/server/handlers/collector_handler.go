package handlers

import (
	"fmt"
	"net/http"

	"github.com/Jourloy/go-metrics-collector/cmd/server/collector"
	"github.com/Jourloy/go-metrics-collector/cmd/server/storage/repository"
)

func RegisterCollectorHandler(mux *http.ServeMux) {
	// Prepare for .env
	metricEndpoint := `/update/`

	// Initiate database
	s := repository.CreateRepository()
	fmt.Println(`Repository handler registered on`, metricEndpoint)

	mux.Handle(metricEndpoint, collector.CollectMetric(s))
	fmt.Println(`Collector handler registered on`, metricEndpoint)
}
