package value

import (
	"errors"
	"net/http"

	"github.com/Jourloy/go-metrics-collector/cmd/server/storage"
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

func GetValueSevice(s storage.Storage) *ValueSevice {
	return &ValueSevice{
		storage: s,
	}
}

func (v *ValueSevice) ServeHTTP(ctx *gin.Context) {
	metricType, typeFound := ctx.Params.Get(`type`)
	metricName, nameFound := ctx.Params.Get(`name`)

	if !nameFound || !typeFound {
		ctx.String(http.StatusNotFound, errNotFound.Error())
		return
	}

	if metricType == `counter` {
		value := v.storage.GetCounterValue(metricName)
		ctx.String(http.StatusOK, `%d`, value)
	} else if metricType == `gauge` {
		value := v.storage.GetGaugeValue(metricName)
		ctx.String(http.StatusOK, `%f`, value)
	} else {
		ctx.String(http.StatusNotFound, errType.Error())
	}
}
