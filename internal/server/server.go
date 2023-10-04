package server

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/Jourloy/go-metrics-collector/internal/server/handlers"
	"github.com/Jourloy/go-metrics-collector/internal/server/storage/repository"
	"github.com/gin-gonic/gin"
)

func Start() {
	host := flag.String("host", `localhost:8080`, "Host of the server")

	flag.Parse()

	s := repository.CreateRepository()

	// Initiate handlers
	r := gin.Default()

	r.LoadHTMLGlob(`templates/*`)

	handlers.RegisterAppHandler(r, s)
	handlers.RegisterCollectorHandler(r, s)
	handlers.RegisterValueHandler(r, s)

	if err := r.Run(*host); err != nil {
		if err == http.ErrServerClosed {
			fmt.Println(`Server closed`)
		} else {
			panic(err)
		}
	}
}
