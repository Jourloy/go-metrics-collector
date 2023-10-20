package main

import (
	"github.com/Jourloy/go-metrics-collector/internal/server"
	"go.uber.org/zap"
)

func init() {
	zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
}

func main() {
	server.Start()
}
