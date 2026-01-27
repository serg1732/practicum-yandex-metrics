package main

import (
	"log"

	"github.com/serg1732/practicum-yandex-metrics/internal/config"
	"github.com/serg1732/practicum-yandex-metrics/internal/service"
)

func main() {
	var agentConfig config.AgentConfig
	parseFlags(&agentConfig)
	agent := service.BuildCollector(agentConfig)
	if err := agent.Run(agentConfig); err != nil {
		log.Fatal(err)
	}
}
