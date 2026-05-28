package config

import (
	"flag"
	"os"
	"runtime"

	"github.com/caarlos0/env/v11"
)

// AgentConfig конфигурация агента (сборщика) метрик.
type AgentConfig struct {
	// RemoteAddr - адрес сервера отправки метрик.
	RemoteAddr string `env:"ADDRESS" json:"address"`
	// Key - проверка hash значений запросов.
	Key string `env:"KEY" json:"key"`
	// CryptoKey ключ асимметричного шифрования
	CryptoKey string `env:"CRYPTO_KEY" json:"crypto_key"`
	// ConfigPath путь до JSON конфига
	ConfigPath string `env:"CONFIG"`
	// ReportInterval - интервал отправки метрик на сервер.
	ReportInterval int `env:"REPORT_INTERVAL" json:"report_interval"`
	// PollInterval - интервал сбора метрик.
	PollInterval int `env:"POLL_INTERVAL" json:"poll_interval"`
	// RateLimit - максимальное количество обработчиков на отправку метрик.
	RateLimit int `env:"RATE_LIMIT" json:"rate_limit"`
}

// GetAgentConfig создает и собирает значение из флагов командной строк и env значений
func GetAgentConfig() (*AgentConfig, error) {
	agentConfig := AgentConfig{
		RemoteAddr:     "localhost:8080",
		ReportInterval: 10,
		PollInterval:   2,
		Key:            "",
		RateLimit:      runtime.NumCPU(),
		CryptoKey:      "",
	}
	if path := configPathFromArgs(os.Args); path != "" {
		err := applyAgentJSONConfig(&agentConfig, path)
		if err != nil {
			return nil, err
		}
	}

	flag.StringVar(&agentConfig.RemoteAddr, "a", agentConfig.RemoteAddr, "address and port to run server")
	flag.IntVar(&agentConfig.ReportInterval, "r", agentConfig.ReportInterval, "interval between reports")
	flag.IntVar(&agentConfig.PollInterval, "p", agentConfig.PollInterval, "interval between polls")
	flag.StringVar(&agentConfig.Key, "k", agentConfig.Key, "key SHA256")
	flag.IntVar(&agentConfig.RateLimit, "l", agentConfig.RateLimit, "rate limit")
	flag.StringVar(&agentConfig.CryptoKey, "crypto-key", agentConfig.CryptoKey, "crypto key public")
	flag.StringVar(&agentConfig.ConfigPath, "c", agentConfig.ConfigPath, "config file path")
	flag.StringVar(&agentConfig.ConfigPath, "config", agentConfig.ConfigPath, "config file path")
	flag.Parse()

	if err := env.Parse(&agentConfig); err != nil {
		return nil, err
	}
	return &agentConfig, nil
}
