package app

import (
	"net/http"

	"github.com/Jourloy/go-metrics-collector/cmd/server/storage"
	"github.com/gin-gonic/gin"
)

type AppSevice struct {
	storage storage.Storage
}

func GetAppSevice(s storage.Storage) *AppSevice {
	return &AppSevice{
		storage: s,
	}
}

func (a *AppSevice) ServeHTTP(ctx *gin.Context) {

	merged := make(map[string]any)

	gauge, counter := a.storage.GetValues()

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
