package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/serg1732/practicum-yandex-metrics/internal/config"
)

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func parseFlags(config *config.ServerConfig) {
	flag.StringVar(&config.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Parse()
}

// parseEnvs обрабатывает env значения
func parseEnvs(serverConfig *config.ServerConfig) {
	err := env.Parse(serverConfig)
	if err != nil {
		log.Fatal(err)
	}
}
