package app

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var errType error = errors.New(`type not found`)
var errBody error = errors.New(`body not found`)
var errCounter error = errors.New(`counter value not found`)
var errGauge error = errors.New(`gauge value not found`)
var errNotFound error = errors.New(`404 page not found`)

type AppSevice struct {
	storage storage.Storage
}

type Metric struct {
	ID    string   `json:"id"`              // Name of metric
	MType string   `json:"type"`            // Gauge or Counter
	Delta *int64   `json:"delta,omitempty"` // Value if metric is a counter
	Value *float64 `json:"value,omitempty"` // Value if metric is a gauge
}

// GetAppSevice returns an instance of AppService initialized with the given storage.
//
// Parameters:
//   - s: the storage instance to be used by the AppService.
//
// Return:
//   - *AppService: a pointer to the initialized AppService instance.
func GetAppSevice(s storage.Storage) *AppSevice {
	return &AppSevice{
		storage: s,
	}
}

// GetAllMetrics retrieves all metrics from the storage and returns them in the HTML format.
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

// Live returns the response "Live" with a status code of 200 OK.
func (a *AppSevice) Live(c *gin.Context) {
	c.String(http.StatusOK, "Live")
}

func (a *AppSevice) UpdateMetrics(ctx *gin.Context) {

	// Parse body
	metric, err := a.parseBody(ctx)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	// Update
	var updated Metric
	if metric.MType == `counter` {
		if metric.Delta == nil {
			ctx.String(http.StatusBadRequest, errCounter.Error())
			return
		}
		u := a.storage.UpdateCounterMetric(metric.ID, *metric.Delta)
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
		u := a.storage.UpdateGaugeMetric(metric.ID, *metric.Value)
		updated = Metric{
			ID:    metric.ID,
			MType: metric.MType,
			Value: &u,
		}
	} else {
		ctx.String(http.StatusBadRequest, errType.Error())
		return
	}

	// Response
	ctx.Header(`Content-Type`, `application/json`)
	ctx.JSON(http.StatusOK, updated)
}

// ShowValue handles the request to show a metric value.
func (a *AppSevice) GetMetric(ctx *gin.Context) {
	// Parse body
	template, err := a.parseBody(ctx)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	// Get
	var metric Metric
	if template.MType == `counter` {
		zap.L().Debug(template.ID)
		u, err := a.storage.GetCounterValue(template.ID)
		if !err {
			ctx.String(http.StatusNotFound, errNotFound.Error())
			return
		}
		metric = Metric{
			ID:    template.ID,
			MType: template.MType,
			Delta: &u,
		}
	} else if template.MType == `gauge` {
		u, err := a.storage.GetGaugeValue(template.ID)
		if !err {
			ctx.String(http.StatusNotFound, errNotFound.Error())
			return
		}
		metric = Metric{
			ID:    template.ID,
			MType: template.MType,
			Value: &u,
		}
	} else {
		ctx.String(http.StatusBadRequest, errType.Error())
		return
	}

	// Response
	ctx.Header(`Content-Type`, `application/json`)
	ctx.JSON(http.StatusOK, metric)
}

func (a *AppSevice) parseBody(ctx *gin.Context) (Metric, error) {
	// Check body
	if ctx.Request.Body == nil {
		return Metric{}, errBody
	}

	// Read body
	b, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return Metric{}, err
	}

	// Close body
	defer ctx.Request.Body.Close()

	// Unmarshal
	var body Metric
	if err := json.Unmarshal(b, &body); err != nil {
		return Metric{}, err
	}

	return body, nil
}
