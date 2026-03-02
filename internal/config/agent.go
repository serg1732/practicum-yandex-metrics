package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

type AgentConfig struct {
	RemoteAddr     string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
}

// GetAgentConfig создает и собирает значение из флагов командной строк и env значений
func GetAgentConfig() (*AgentConfig, error) {
	var agentConfig AgentConfig
	flag.StringVar(&agentConfig.RemoteAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&agentConfig.ReportInterval, "r", 10, "interval between reports")
	flag.IntVar(&agentConfig.PollInterval, "p", 2, "interval between polls")
	flag.StringVar(&agentConfig.Key, "k", "", "key SHA256")
	flag.Parse()

	if err := env.Parse(&agentConfig); err != nil {
		return nil, err
	}
	return &agentConfig, nil
}
