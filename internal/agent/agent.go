package agent

import (
	"flag"
	"fmt"

	"github.com/Jourloy/go-metrics-collector/internal/agent/collector"
	"github.com/joho/godotenv"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(`.env.agent`); err != nil {
		fmt.Println(`.env.agent not found`)
	}
}

func Start() {
	flag.Parse()

	agent := collector.CreateCollector()
	agent.StartTickers()
}
