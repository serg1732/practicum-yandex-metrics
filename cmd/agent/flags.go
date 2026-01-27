package main

import (
	"flag"

	"github.com/serg1732/practicum-yandex-metrics/internal/config"
)

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func parseFlags(agentConfig *config.AgentConfig) {
	flag.StringVar(&agentConfig.RemoteAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&agentConfig.ReportInterval, "r", 10, "interval between reports")
	flag.IntVar(&agentConfig.PollInterval, "p", 2, "interval between polls")
	flag.Parse()
}
