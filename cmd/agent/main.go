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

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	printBuildInfo()
	log := logger.NewSlogLogger(slog.LevelInfo)
	agentConfig, errConfig := config.GetAgentConfig()

	if errConfig != nil {
		log.Error("Ошибка парсинга flag/env значений", "error", errConfig.Error())
		return
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
			return
		}
	}(ctx)

	<-ctx.Done()
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", valueOrNA(buildVersion))
	fmt.Printf("Build date: %s\n", valueOrNA(buildDate))
	fmt.Printf("Build commit: %s\n", valueOrNA(buildCommit))
}

func valueOrNA(value string) string {
	if value == "" {
		return "N/A"
	}

	return value
}
