package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func live(c *gin.Context) {
	c.String(http.StatusOK, "Live")
}

func RegisterLiveHandler(r *gin.Engine) {
	r.GET("/live", live)

	fmt.Println(`Mapped /live`)
}
