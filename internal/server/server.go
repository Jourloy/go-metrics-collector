// Package server init gin server, add middlewares, init database,
// add start server
package server

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	"github.com/Jourloy/go-metrics-collector/internal/server/handlers"
	"github.com/Jourloy/go-metrics-collector/internal/server/middlewares"
	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/Jourloy/go-metrics-collector/internal/server/storage/repository"
)

var (
	Host = flag.String(`a`, `localhost:8080`, `Host of the server`)
	Key  = flag.String(`key`, ``, `Key for cipher`)
)

type ServerConfig struct {
	Address       string `json:"address"`
	Restore       bool   `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
}

func readConfig() {
	if file, err := os.Open(`./server.config.json`); err == nil {
		defer file.Close()

		b, _ := io.ReadAll(file)
		var config ServerConfig
		json.Unmarshal(b, &config)

		Host = &config.Address
	}
}

// Start initiates the application.
func Start() {
	readConfig()

	// Initiate handlers
	r := gin.New()

	// Set middlewares
	r.Use(gin.Recovery())           // 500 instead of panic
	r.Use(middlewares.Logger())     // Logger
	r.Use(middlewares.GzipDecode()) // Gzip
	r.Use(middlewares.HashDecode()) // Hash

	// Check if ADDRESS environment variable is set and assign it to Host
	if hostENV, exist := os.LookupEnv(`ADDRESS`); exist {
		Host = &hostENV
	}

	flag.Parse()

	// Create storage
	//
	// If postgres DSN is set and not valid, ok will be false. In that case,
	// I set s to nil for return 500 error on ping request
	var s storage.Storage
	if storage, ok := repository.CreateRepository(); ok {
		s = storage
	} else {
		s = nil
	}

	// Load HTML templates
	r.LoadHTMLGlob(`templates/*`)

	// Initiate router groups
	pprof.Register(r)
	appGroup := r.Group(`/`)

	// Register application, collector, and value handlers
	handlers.RegisterAppHandler(appGroup, s)

	srv := &http.Server{
		Addr:    *Host,
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		panic(err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	}
	log.Println("Server exiting")
}
