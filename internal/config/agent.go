package config

import (
	"flag"
	"runtime"

	"github.com/caarlos0/env/v11"
)

// AgentConfig конфигурация агента (сборщика) метрик.
type AgentConfig struct {
	// RemoteAddr - адрес сервера отправки метрик.
	RemoteAddr string `env:"ADDRESS"`
	// ReportInterval - интервал отправки метрик на сервер.
	ReportInterval int `env:"REPORT_INTERVAL"`
	// PollInterval - интервал сбора метрик.
	PollInterval int `env:"POLL_INTERVAL"`
	// Key - проверка hash значений запросов.
	Key string `env:"KEY"`
	// RateLimit - максимальное количество обработчиков на отправку метрик.
	RateLimit int `env:"RATE_LIMIT"`
}

// GetAgentConfig создает и собирает значение из флагов командной строк и env значений
func GetAgentConfig() (*AgentConfig, error) {
	var agentConfig AgentConfig
	flag.StringVar(&agentConfig.RemoteAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&agentConfig.ReportInterval, "r", 10, "interval between reports")
	flag.IntVar(&agentConfig.PollInterval, "p", 2, "interval between polls")
	flag.StringVar(&agentConfig.Key, "k", "", "key SHA256")
	flag.IntVar(&agentConfig.RateLimit, "l", runtime.NumCPU(), "rate limit")
	flag.Parse()

	if err := env.Parse(&agentConfig); err != nil {
		return nil, err
	}
	return &agentConfig, nil
}
