package main

import (
	"fmt"
	"net/http"

	"github.com/Jourloy/go-metrics-collector/cmd/server/handlers"
	"github.com/Jourloy/go-metrics-collector/cmd/server/storage/repository"
	"github.com/gin-gonic/gin"
)

func main() {
	// Prepare for .env
	port := `8080`

	s := repository.CreateRepository()

	// Initiate handlers
	r := gin.Default()

	r.LoadHTMLGlob(`templates/*`)

	handlers.RegisterAppHandler(r, s)
	handlers.RegisterCollectorHandler(r, s)
	handlers.RegisterValueHandler(r, s)

	if err := r.Run(fmt.Sprintf(`:%s`, port)); err != nil {
		if err == http.ErrServerClosed {
			fmt.Println(`Server closed`)
		} else {
			panic(err)
		}
	}
}
