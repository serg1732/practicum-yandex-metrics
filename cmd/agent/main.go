package main

import (
	"log/slog"
	"os"

	"github.com/serg1732/practicum-yandex-metrics/internal/config"
	"github.com/serg1732/practicum-yandex-metrics/internal/logger"
	"github.com/serg1732/practicum-yandex-metrics/internal/service"
)

func main() {
	log := logger.NewSlogLogger(slog.LevelInfo)
	agentConfig, errConfig := config.GetAgentConfig()

	if errConfig != nil {
		log.Error("Ошибка парсинга flag/env значений", "error", errConfig.Error())
		os.Exit(1)
	}
	log.Debug("Значения после чтения flags/env", "config", agentConfig)

	agent := service.BuildCollector()
	if err := agent.Run(log, *agentConfig); err != nil {
		log.Error("Ошибка сборщика метрика", "error", err.Error())
		os.Exit(1)
	}
}
