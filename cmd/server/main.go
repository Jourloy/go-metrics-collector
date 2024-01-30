package main

import (
	_ "net/http/pprof"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/Jourloy/go-metrics-collector/internal/server"
)

func init() {
	zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
}

func main() {
	if err := godotenv.Load(`.env.server`); err != nil {
		zap.L().Warn(`.env.server not found`)
	}
	zap.L().Info(`Application initialized`)

	server.Start()
}
