package main

import (
	_ "net/http/pprof"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/Jourloy/go-metrics-collector/internal/server"
)

var buildVersion string = `N/A`
var buildDate string = `N/A`
var buildCommit string = `N/A`

func init() {
	zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
}

func main() {
	if err := godotenv.Load(`.env.server`); err != nil {
		zap.L().Warn(`.env.server not found`)
	}
	zap.L().Info(`Application initialized`)

	zap.L().Info(`Information about app`, zap.String(`buildVersion`, buildVersion), zap.String(`buildDate`, buildDate), zap.String(`buildCommit`, buildCommit))

	server.Start()
}
