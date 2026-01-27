package main

import (
	"flag"

	"github.com/serg1732/practicum-yandex-metrics/internal/config"
)

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func parseFlags(config *config.ServerConfig) {
	flag.StringVar(&config.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Parse()
}
