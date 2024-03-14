// Package server init gin server, add middlewares, init database,
// add start server
package server

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	limit "github.com/bu/gin-access-limit"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"github.com/Jourloy/go-metrics-collector/internal/proto"
	"github.com/Jourloy/go-metrics-collector/internal/server/handlers"
	"github.com/Jourloy/go-metrics-collector/internal/server/middlewares"
	"github.com/Jourloy/go-metrics-collector/internal/server/rpc"
	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/Jourloy/go-metrics-collector/internal/server/storage/repository"
)

var (
	Host          = flag.String(`a`, `localhost:8080`, `Host of the server`)
	Key           = flag.String(`key`, ``, `Key for cipher`)
	TrustedSubnet = flag.String(`t`, ``, `CIDR`)
)

type ServerConfig struct {
	Address       string `json:"address"`
	Restore       bool   `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
	TrustedSubnet string `json:"trusted_subnet"`
}

func readConfig() {
	if file, err := os.Open(`./server.config.json`); err == nil {
		defer file.Close()

		b, _ := io.ReadAll(file)
		var config ServerConfig
		err := json.Unmarshal(b, &config)
		if err != nil {
			panic(err)
		}

		Host = &config.Address
		TrustedSubnet = &config.TrustedSubnet
	}
}

// Start initiates the application.
func Start() {
	readConfig()
	flag.Parse()

	// Initiate handlers
	r := gin.New()

	// Set middlewares
	r.Use(gin.Recovery())           // 500 instead of panic
	r.Use(middlewares.Logger())     // Logger
	r.Use(middlewares.GzipDecode()) // Gzip
	r.Use(middlewares.HashDecode()) // Hash

	if *TrustedSubnet != `` {
		r.Use(limit.CIDR(*TrustedSubnet))
	}

	// Check if ADDRESS environment variable is set and assign it to Host
	if hostENV, exist := os.LookupEnv(`ADDRESS`); exist {
		Host = &hostENV
	}

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

	go startGRPC()

	srv := &http.Server{
		Addr:    *Host,
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		panic(err)
	}
	<-ctx.Done()

	log.Println("Server exiting")
}

func startGRPC() {
	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		log.Fatal(err)
	}
	// создаём gRPC-сервер без зарегистрированной службы
	s := grpc.NewServer()
	// регистрируем сервис
	proto.RegisterMetricServiceServer(s, &rpc.MetricServer{})

	fmt.Println("Сервер gRPC начал работу")
	// получаем запрос gRPC
	if err := s.Serve(listen); err != nil {
		log.Fatal(err)
	}
}
