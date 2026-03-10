package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	log.Debug("Значения после чтения flags/env", "config", agentConfig)
	go func(ctx context.Context) {
		agent := service.BuildCollector()
		if err := agent.Run(ctx, log, *agentConfig); err != nil {
			log.Error("Ошибка сборщика метрика", "error", err.Error())
			os.Exit(1)
		}
	}(ctx)

	<-ctx.Done()
}
