package main

import (
	"log"

	"github.com/serg1732/practicum-yandex-metrics/internal/service"
)

func main() {
	parseFlags()
	agent := service.BuildCollector("http://" + remoteAddr)
	if err := agent.Run(pollInterval, reportInterval); err != nil {
		log.Fatal(err)
	}
}
