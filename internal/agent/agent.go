package agent

import "github.com/Jourloy/go-metrics-collector/internal/agent/collector"

func Start() {
	agent := collector.CreateCollector()
	agent.StartTickers()
}
