package handlers

import (
	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/Jourloy/go-metrics-collector/internal/server/value"
	"github.com/gin-gonic/gin"
)

// Register a value handler for the specified Gin Engine.
func RegisterValueHandler(g *gin.RouterGroup, s storage.Storage) {
	valueService := value.GetValueSevice(s)

	g.GET(`/:type/:name`, valueService.ShowValue)
}
