package app

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/gin-gonic/gin"
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

func (a *AppSevice) UpdateMetricByParams(ctx *gin.Context) {
	name := ctx.Param(`name`)
	mType := ctx.Param(`type`)
	value := ctx.Param(`value`)

	// Check URL params
	if mType == `` {
		ctx.String(http.StatusBadRequest, errType.Error())
		return
	}

	if name == `` {
		ctx.String(http.StatusNotFound, errName.Error())
		return
	}

	if value == `` {
		ctx.String(http.StatusBadRequest, errValue.Error())
		return
	}

	// Update metric
	metric, err := a.updateMetric(name, mType, nil, nil, &value)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.Header(`Content-Type`, `application/json`)
	ctx.JSON(http.StatusOK, metric)

}

func (a *AppSevice) UpdateMetricByBody(ctx *gin.Context) {
	// Check body
	metric, err := a.parseBody(ctx)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	// Update metric
	updated, err := a.updateMetric(metric.ID, metric.MType, metric.Value, metric.Delta, nil)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.Header(`Content-Type`, `application/json`)
	ctx.JSON(http.StatusOK, updated)
}

func (a *AppSevice) updateMetric(name string, mType string, value *float64, delta *int64, strValue *string) (Metric, error) {
	var updated Metric
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

		// Update counter
		u := a.storage.UpdateCounterMetric(name, v)
		updated = Metric{
			ID:    name,
			MType: mType,
			Delta: &u,
		}
	} else if mType == `gauge` {
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

		// Update gauge
		u := a.storage.UpdateGaugeMetric(name, v)
		updated = Metric{
			ID:    name,
			MType: mType,
			Value: &u,
		}
	} else {
		return Metric{}, errType
	}

	return updated, nil
}

// ShowValue handles the request to show a metric value.
func (a *AppSevice) GetMetricByParams(ctx *gin.Context) {
	name := ctx.Param(`name`)
	mType := ctx.Param(`type`)

	// Check URL params
	if name == `` || mType == `` {
		ctx.String(http.StatusBadRequest, errNotFound.Error())
		return
	}

	// Get metric
	metric, err := a.getMetric(name, mType)
	if err != nil {
		ctx.String(http.StatusNotFound, errNotFound.Error())
		return
	}

	ctx.Header(`Content-Type`, `plain/text`)
	if mType == `counter` {
		ctx.String(http.StatusOK, `%d`, metric.Delta)
	} else {
		ctx.String(http.StatusOK, `%g`, metric.Value)
	}

}

func (a *AppSevice) GetMetricByBody(ctx *gin.Context) {
	// Check body
	template, err := a.parseBody(ctx)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	// Get metric
	metric, err := a.getMetric(template.ID, template.MType)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.Header(`Content-Type`, `application/json`)
	ctx.JSON(http.StatusOK, metric)
}

func (a *AppSevice) getMetric(name string, mType string) (Metric, error) {
	var metric Metric
	if mType == `counter` {
		u, err := a.storage.GetCounterValue(name)
		if !err {
			return Metric{}, errNotFound
		}
		metric = Metric{
			ID:    name,
			MType: mType,
			Delta: &u,
		}
	} else if mType == `gauge` {
		u, err := a.storage.GetGaugeValue(name)
		if !err {
			return Metric{}, errNotFound
		}
		metric = Metric{
			ID:    name,
			MType: mType,
			Value: &u,
		}
	}

	return metric, nil
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
