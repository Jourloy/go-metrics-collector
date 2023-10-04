package agent

import (
	"flag"
	"fmt"

	"github.com/Jourloy/go-metrics-collector/internal/agent/collector"
	"github.com/joho/godotenv"
)

func Start() {
	if err := godotenv.Load(); err != nil {
		fmt.Println(`.env.agent not found`)
	}

	flag.Parse()

	agent := collector.CreateCollector()
	agent.StartTickers()
}
