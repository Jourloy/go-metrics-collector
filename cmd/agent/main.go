package main

import (
	"github.com/Jourloy/go-metrics-collector/internal/agent"
	"go.uber.org/zap"
)

func init() {
	zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
}

func main() {
	agent.Start()
}
