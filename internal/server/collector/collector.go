package collector

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/gin-gonic/gin"
)

type Metric struct {
	ID    string   `json:"id"`              // Name of metric
	MType string   `json:"type"`            // Gauge or Counter
	Delta *int64   `json:"delta,omitempty"` // Value if metric is a counter
	Value *float64 `json:"value,omitempty"` // Value if metric is a gauge
}

type CollectorHandler struct {
	storage storage.Storage
}

var errType error = errors.New(`type not found`)
var errBody error = errors.New(`body not found`)
var errCounter error = errors.New(`counter value not found`)
var errGauge error = errors.New(`gauge value not found`)

// CollectMetric returns a new instance of CollectorHandler which collects metrics using the provided storage.
//
// Parameters:
// - s: the storage implementation used for collecting metrics.
//
// Returns:
// - a pointer to the CollectorHandler instance.
func CollectMetric(s storage.Storage) *CollectorHandler {
	return &CollectorHandler{
		storage: s,
	}
}

func (c *CollectorHandler) ProcessMetrics(ctx *gin.Context) {

	// Check body
	if ctx.Request.Body == nil {
		ctx.String(http.StatusBadRequest, errBody.Error())
		return
	}

	// Read body
	b, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	// Close body
	defer ctx.Request.Body.Close()

	// Unmarshal
	var metric Metric
	if err := json.Unmarshal(b, &metric); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	// Check type
	if metric.MType != `counter` && metric.MType != `gauge` {
		ctx.String(http.StatusBadRequest, errType.Error())
		return
	}

	// Update
	var updated Metric
	if metric.MType == `counter` {
		if metric.Delta == nil {
			ctx.String(http.StatusBadRequest, errCounter.Error())
			return
		}
		u := c.storage.UpdateCounterMetric(metric.ID, *metric.Delta)
		updated = Metric{
			ID:    metric.ID,
			MType: metric.MType,
			Delta: &u,
		}
	} else if metric.MType == `gauge` {
		if metric.Value == nil {
			ctx.String(http.StatusBadRequest, errGauge.Error())
			return
		}
		u := c.storage.UpdateGaugeMetric(metric.ID, *metric.Value)
		updated = Metric{
			ID:    metric.ID,
			MType: metric.MType,
			Value: &u,
		}
	}

	// Response
	ctx.Header(`Content-Type`, `application/json`)
	ctx.JSON(http.StatusOK, updated)
}
