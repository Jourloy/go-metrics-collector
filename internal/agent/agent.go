package agent

import (
	"flag"

	"github.com/Jourloy/go-metrics-collector/internal/agent/collector"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func Start() {
	if err := godotenv.Load(`.env.agent`); err != nil {
		zap.L().Warn(`.env.agent not found`)
	}

	flag.Parse()

	agent := collector.CreateCollector()
	agent.StartTickers()
}
