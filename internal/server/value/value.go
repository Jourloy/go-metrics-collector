package value

import (
	"errors"
	"net/http"

	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/gin-gonic/gin"
)

type ValueSevice struct {
	storage storage.Storage
}

type ParsedURL struct {
	Name string
}

var errNotFound error = errors.New(`404 page not found`)
var errType error = errors.New(`type not found`)

// Create a new instance of the ValueService struct.
func GetValueSevice(s storage.Storage) *ValueSevice {
	return &ValueSevice{
		storage: s,
	}
}

// Handle the HTTP request for the ValueService.
func (v *ValueSevice) ShowValue(ctx *gin.Context) {
	// Retrieve the metric type and name from the request parameters.
	metricType, typeFound := ctx.Params.Get(`type`)
	metricName, nameFound := ctx.Params.Get(`name`)

	// If either the metric type or name is not found, return a 404 error.
	if !nameFound || !typeFound {
		ctx.String(http.StatusNotFound, errNotFound.Error())
		return
	}

	// Retrieve the value based on the metric type and name.
	if metricType == `counter` {
		value, ok := v.storage.GetCounterValue(metricName)
		if !ok {
			ctx.String(http.StatusNotFound, errNotFound.Error())
			return
		}
		ctx.String(http.StatusOK, `%d`, value)
	} else if metricType == `gauge` {
		value, ok := v.storage.GetGaugeValue(metricName)
		if !ok {
			ctx.String(http.StatusNotFound, errNotFound.Error())
			return
		}
		ctx.String(http.StatusOK, `%g`, value)
	} else {
		// If the metric type is neither "counter" nor "gauge", return a 404 error.
		ctx.String(http.StatusNotFound, errType.Error())
	}
}
