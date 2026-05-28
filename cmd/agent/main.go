package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/serg1732/practicum-yandex-metrics/internal/config"
	"github.com/serg1732/practicum-yandex-metrics/internal/logger"
	"github.com/serg1732/practicum-yandex-metrics/internal/service"
)

var buildVersion = "N/A"
var buildDate = "N/A"
var buildCommit = "N/A"

func main() {
	printBuildInfo()
	log := logger.NewSlogLogger(slog.LevelInfo)
	agentConfig, errConfig := config.GetAgentConfig()

	if errConfig != nil {
		log.Error("Ошибка парсинга flag/env значений", "error", errConfig.Error())
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT,
	)
	defer stop()

	log.Debug("Значения после чтения flags/env", "config", agentConfig)
	go func(ctx context.Context) {
		agent := service.BuildCollector()
		if err := agent.Run(ctx, log, *agentConfig); err != nil {
			log.Error("Ошибка сборщика метрика", "error", err.Error())
			return
		}
	}(ctx)

	<-ctx.Done()
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
