package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"

	"github.com/caarlos0/env/v6"
	"github.com/serg1732/practicum-yandex-metrics/internal/config"
)

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func parseFlags(config *config.ServerConfig) {
	flag.StringVar(&config.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&config.StoreInternal, "i", 5, "time to save file storage server")
	flag.StringVar(&config.FileStoragePath, "f", "storage.json", "address file storage server")
	flag.BoolVar(&config.Restore, "r", false, "restore storage server")
	flag.Parse()
	slog.Debug("Прочитан конфиг flags: " + fmt.Sprintf("%v", config))
}

// parseEnvs обрабатывает env значения
func parseEnvs(serverConfig *config.ServerConfig) {
	err := env.Parse(serverConfig)
	slog.Debug("Прочитан конфиг ENV: " + fmt.Sprintf("%v", serverConfig))
	if err != nil {
		log.Fatal(err)
	}
}
