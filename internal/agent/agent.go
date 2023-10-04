package agent

import (
	"flag"

	"github.com/Jourloy/go-metrics-collector/internal/agent/collector"
)

func Start() {
	flag.Parse()

	agent := collector.CreateCollector()
	agent.StartTickers()
}
