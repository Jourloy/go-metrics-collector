package main

import "github.com/Jourloy/go-metrics-collector/cmd/agent/collector"

func main() {
	agent := collector.CreateCollector()
	agent.StartTickers()
}
