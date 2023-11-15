package app

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var errType error = errors.New(`type is invalid or not found`)
var errValue error = errors.New(`value is invalid or not found`)
var errName error = errors.New(`name is invalid or not found`)
var errCounter error = errors.New(`counter value not found`)
var errGauge error = errors.New(`gauge value not found`)
var errNotFound error = errors.New(`404 page not found`)
var errBody error = errors.New(`body not found`)

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

// Live returns the response "Live" with a status code of 200 OK.
//
// Parameters:
//   - ctx: the gin context.
func (a *AppSevice) Pong(c *gin.Context) {
	// Check storage
	if !a.checkStorage(c) {
		return
	}

	c.String(http.StatusOK, `Pong`)
}

// UpdateMetricByParams updates a metric by its parameters.
//
// Parameters:
//   - ctx: the gin context.
func (a *AppSevice) UpdateMetricByParams(ctx *gin.Context) {
	if !a.checkStorage(ctx) {
		return
	}

	name := ctx.Param(`name`)
	mType := ctx.Param(`type`)
	value := ctx.Param(`value`)

	// Check URL params
	if mType == `` {
		zap.L().Error(errType.Error())
		ctx.String(http.StatusBadRequest, errType.Error())
		return
	}

	if name == `` {
		zap.L().Error(errName.Error())
		ctx.String(http.StatusNotFound, errName.Error())
		return
	}

	if value == `` {
		zap.L().Error(errValue.Error())
		ctx.String(http.StatusBadRequest, errValue.Error())
		return
	}

	// Check metric type
	if !a.checkMetricType(mType, ctx) {
		return
	}

	// Update metric
	metric, err := a.updateMetric(name, mType, nil, nil, &value)
	if err != nil {
		zap.L().Error(err.Error())
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, metric)
}

// UpdateMetricByBody updates a metric by parsing the request body.
//
// Parameters:
//   - ctx: the gin context.
func (a *AppSevice) UpdateMetricByBody(ctx *gin.Context) {
	// Check storage
	if !a.checkStorage(ctx) {
		return
	}

	// Check body
	metric, err := a.parseBody(ctx)
	if err != nil {
		zap.L().Error(err.Error())
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	// Check metric type
	if !a.checkMetricType(metric.MType, ctx) {
		return
	}

	// Update metric
	updated, err := a.updateMetric(metric.ID, metric.MType, metric.Value, metric.Delta, nil)
	if err != nil {
		zap.L().Error(err.Error())
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, updated)
}

type Metrics []Metric

// UpdateManyMetrics updates multiple metrics from butch request.
func (a *AppSevice) UpdateManyMetrics(ctx *gin.Context) {
	// Check storage
	if !a.checkStorage(ctx) {
		return
	}

	// Check body
	if ctx.Request.Body == nil {
		zap.L().Error(errBody.Error())
		ctx.String(http.StatusBadRequest, errBody.Error())
		return
	}

	// Read body
	b, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		zap.L().Error(errBody.Error())
		ctx.String(http.StatusBadRequest, errBody.Error())
		return
	}

	// Unmarshal
	var body Metrics
	if err := json.Unmarshal(b, &body); err != nil {
		zap.L().Error(errBody.Error())
		ctx.String(http.StatusBadRequest, errBody.Error())
		return
	}

	for _, metric := range body {
		// Check metric type
		if !a.checkMetricType(metric.MType, ctx) {
			continue
		}

		_, err := a.updateMetric(metric.ID, metric.MType, metric.Value, metric.Delta, nil)
		if err != nil {
			zap.L().Error(`Failed to update metric`, zap.Error(err))
			continue
		}
	}

	ctx.JSON(http.StatusOK, body)
}

// updateMetric updates a metric based on the provided parameters.
//
// Parameters:
// - name: the name of the metric.
// - mType: the type of the metric. Only `counter` and `gauge` are supported.
// - value: the value of the metric (optional).
// - delta: the delta value of the metric (optional).
// - strValue: the string value of the metric (optional).
//
// Returns:
// - Metric: the updated metric.
// - error: an error if the metric update fails.
func (a *AppSevice) updateMetric(name string, mType string, value *float64, delta *int64, strValue *string) (Metric, error) {
	// Update counter metric
	if mType == `counter` {
		var v int64

		// Find `Delta` value
		if delta != nil {
			v = *delta
		} else if strValue != nil {
			parsedValue, err := strconv.ParseInt(*strValue, 10, 64)
			if err != nil {
				return Metric{}, errCounter
			}
			v = parsedValue
		} else {
			return Metric{}, errCounter
		}

		// Update metric
		u := a.storage.UpdateCounterMetric(name, v)
		updated := Metric{
			ID:    name,
			MType: mType,
			Delta: &u,
		}

		return updated, nil
	}

	// Update gauge metric
	var v float64

	// Find `Value` value
	if value != nil {
		v = *value
	} else if strValue != nil {
		parsedValue, err := strconv.ParseFloat(*strValue, 64)
		if err != nil {
			return Metric{}, errGauge
		}
		v = parsedValue
	} else {
		return Metric{}, errGauge
	}

	// Update metric
	u := a.storage.UpdateGaugeMetric(name, v)
	updated := Metric{
		ID:    name,
		MType: mType,
		Value: &u,
	}

	return updated, nil
}

// ShowValue handles the request to show a metric value.
//
// Parameters:
//   - ctx: the gin context.
func (a *AppSevice) GetMetricByParams(ctx *gin.Context) {
	if !a.checkStorage(ctx) {
		return
	}

	name := ctx.Param(`name`)
	mType := ctx.Param(`type`)

	// Check URL params
	if name == `` || mType == `` {
		zap.L().Error(errNotFound.Error())
		ctx.String(http.StatusBadRequest, errNotFound.Error())
		return
	}

	// Check metric type
	if !a.checkMetricType(mType, ctx) {
		return
	}

	// Get counter metric
	if mType == `counter` {
		u, err := a.storage.GetCounterValue(name)
		if !err {
			zap.L().Error(errNotFound.Error())
			ctx.String(http.StatusNotFound, errNotFound.Error())
			return
		}

		ctx.String(http.StatusOK, `%d`, u)
		return
	}

	// Get gauge metric
	if mType == `gauge` {
		u, err := a.storage.GetGaugeValue(name)
		if !err {
			zap.L().Error(errNotFound.Error())
			ctx.String(http.StatusNotFound, errNotFound.Error())
			return
		}

		ctx.String(http.StatusOK, `%g`, u)
		return
	}
}

// GetMetricByBody retrieves a metric based on the request body.
//
// Parameters:
//   - ctx: the gin context.
func (a *AppSevice) GetMetricByBody(ctx *gin.Context) {
	// Check storage
	if !a.checkStorage(ctx) {
		return
	}

	// Check body
	template, err := a.parseBody(ctx)
	if err != nil {
		zap.L().Error(err.Error())
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	// Check metric type
	if !a.checkMetricType(template.MType, ctx) {
		return
	}

	var metric Metric

	// Get counter metric
	if template.MType == `counter` {
		u, ok := a.storage.GetCounterValue(template.ID)
		if !ok {
			zap.L().Error(errNotFound.Error())
			ctx.String(http.StatusNotFound, errNotFound.Error())
			return
		}
		metric = Metric{
			ID:    template.ID,
			MType: template.MType,
			Delta: &u,
		}
	}

	// Get gauge metric
	if template.MType == `gauge` {
		u, ok := a.storage.GetGaugeValue(template.ID)
		if !ok {
			zap.L().Error(errNotFound.Error())
			ctx.String(http.StatusNotFound, errNotFound.Error())
			return
		}
		metric = Metric{
			ID:    template.ID,
			MType: template.MType,
			Value: &u,
		}
	}

	ctx.Header(`Content-Type`, `application/json`)
	ctx.JSON(http.StatusOK, metric)
}

// GetAllMetrics retrieves all metrics from the storage and returns them in the HTML format.
//
// Parameters:
//   - ctx: the gin context.
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

func (a *AppSevice) checkMetricType(mType string, ctx *gin.Context) bool {
	if mType != `counter` && mType != `gauge` {
		zap.L().Error(errType.Error())
		ctx.String(http.StatusBadRequest, errType.Error())
		return false
	}

	return true
}

// checkStorage checks if the storage is initialized.
//
// Parameter(s):
//   - c: a gin.Context object
//
// Returns:
//   - true if the storage is initialized, false otherwise.
func (a *AppSevice) checkStorage(c *gin.Context) bool {
	if a.storage == nil {
		zap.L().Error(`storage not initialized`)
		c.String(http.StatusInternalServerError, `storage not initialized`)
		return false
	}
	return true
}

// parseBody parses the request body and returns a Metric object and an error.
//
// Parameters:
//   - ctx: the gin context.
//
// Returns:
//   - Metric: the parsed Metric object.
//   - error: an error if the parsing fails.
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

	// Unmarshal
	var body Metric
	if err := json.Unmarshal(b, &body); err != nil {
		return Metric{}, err
	}

	return body, nil
}
