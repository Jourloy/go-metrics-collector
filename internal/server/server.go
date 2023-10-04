package server

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/Jourloy/go-metrics-collector/internal/server/handlers"
	"github.com/Jourloy/go-metrics-collector/internal/server/storage/repository"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var (
	Host = flag.String("a", `localhost:8080`, "Host of the server")
)

func Start() {
	if err := godotenv.Load(); err != nil {
		fmt.Println(`.env.server not found`)
	}

	flag.Parse()

	s := repository.CreateRepository()

	// Initiate handlers
	r := gin.Default()

	r.LoadHTMLGlob(`templates/*`)

	handlers.RegisterAppHandler(r, s)
	handlers.RegisterCollectorHandler(r, s)
	handlers.RegisterValueHandler(r, s)

	hostENV, exist := os.LookupEnv(`ADDRESS`)
	if !exist {
		if err := r.Run(*Host); err != nil {
			if err == http.ErrServerClosed {
				fmt.Println(`Server closed`)
			} else {
				panic(err)
			}
		}
	} else {
		if err := r.Run(hostENV); err != nil {
			if err == http.ErrServerClosed {
				fmt.Println(`Server closed`)
			} else {
				panic(err)
			}
		}
	}

}
