package handlers

import (
	"github.com/Jourloy/go-metrics-collector/cmd/server/storage"
	"github.com/Jourloy/go-metrics-collector/cmd/server/value"
	"github.com/gin-gonic/gin"
)

func RegisterValueHandler(r *gin.Engine, s storage.Storage) {
	valueHandler := value.GetValueSevice(s)

	r.GET(`/value/:type/:name`, valueHandler.ServeHTTP)
}
