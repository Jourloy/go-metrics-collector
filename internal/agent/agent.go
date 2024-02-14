// Package agent load env and init collector agent
package agent

import (
	"flag"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/Jourloy/go-metrics-collector/internal/agent/collector"
)

// Start initializes the application.
//
// It loads the `.env.agent` file and logs a warning if the file is not found.
// It then parses command line flags.
// Finally, it creates a collector instance, starts the tickers, and begins collecting data.
func Start() {
	if err := godotenv.Load(`.env.agent`); err != nil {
		zap.L().Warn(`.env.agent not found`)
	}

	flag.Parse()

	agent := collector.CreateCollector()
	agent.StartTickers()
}
