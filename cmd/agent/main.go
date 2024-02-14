package main

import (
	"github.com/Jourloy/go-metrics-collector/internal/agent"
	"go.uber.org/zap"
)

var buildVersion string = `N/A`
var buildDate string = `N/A`
var buildCommit string = `N/A`

func init() {
	zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
}

func main() {
	zap.L().Info(`Information about app`, zap.String(`buildVersion`, buildVersion), zap.String(`buildDate`, buildDate), zap.String(`buildCommit`, buildCommit))

	agent.Start()
}
