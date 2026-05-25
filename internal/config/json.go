package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func applyAgentJSONConfig(cfg *AgentConfig, path string) error {
	if path == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ошибка чтения конфига агента: %w", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("ошибка при парсинге конфига: %w", err)
	}

	if v, ok := raw["address"]; ok {
		if err := json.Unmarshal(v, &cfg.RemoteAddr); err != nil {
			return fmt.Errorf("ошибка при парсинге address: %w", err)
		}
	}

	if v, ok := raw["report_interval"]; ok {
		seconds, err := parseDurationSeconds(v)
		if err != nil {
			return fmt.Errorf("ошибка при парсинге report_interval: %w", err)
		}
		cfg.ReportInterval = seconds
	}

	if v, ok := raw["poll_interval"]; ok {
		seconds, err := parseDurationSeconds(v)
		if err != nil {
			return fmt.Errorf("ошибка при парсинге poll_interval: %w", err)
		}
		cfg.PollInterval = seconds
	}

	if v, ok := raw["rate_limit"]; ok {
		seconds, err := parseDurationSeconds(v)
		if err != nil {
			return fmt.Errorf("ошибка при парсинге poll_interval: %w", err)
		}
		cfg.RateLimit = seconds
	}

	if v, ok := raw["crypto_key"]; ok {
		if err := json.Unmarshal(v, &cfg.CryptoKey); err != nil {
			return fmt.Errorf("ошибка при парсинге crypto_key: %w", err)
		}
	}

	if v, ok := raw["key"]; ok {
		if err := json.Unmarshal(v, &cfg.Key); err != nil {
			return fmt.Errorf("ошибка при парсинге crypto_key: %w", err)
		}
	}

	return nil
}

func applyServerJSONConfig(cfg *ServerConfig, path string) error {
	if path == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ошибка при чтении конфига сервера: %w", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("ошибка при парсинге конфига: %w", err)
	}

	if v, ok := raw["address"]; ok {
		if err := json.Unmarshal(v, &cfg.RunAddr); err != nil {
			return fmt.Errorf("ошибка при парсинге address: %w", err)
		}
	}

	if v, ok := raw["store_interval"]; ok {
		seconds, err := parseDurationSeconds(v)
		if err != nil {
			return fmt.Errorf("ошибка при парсинге report_interval: %w", err)
		}
		cfg.StoreInternal = seconds
	}

	if v, ok := raw["crypto_key"]; ok {
		if err := json.Unmarshal(v, &cfg.CryptoKey); err != nil {
			return fmt.Errorf("ошибка при парсинге crypto_key: %w", err)
		}
	}

	if v, ok := raw["key"]; ok {
		if err := json.Unmarshal(v, &cfg.Key); err != nil {
			return fmt.Errorf("ошибка при парсинге crypto_key: %w", err)
		}
	}

	if v, ok := raw["crypto_key"]; ok {
		if err := json.Unmarshal(v, &cfg.CryptoKey); err != nil {
			return fmt.Errorf("ошибка при парсинге crypto_key: %w", err)
		}
	}

	return nil
}

func parseDurationSeconds(raw json.RawMessage) (int, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		d, err := time.ParseDuration(s)
		if err != nil {
			return 0, err
		}

		return int(d.Seconds()), nil
	}

	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return n, nil
	}

	return 0, fmt.Errorf("должен быть формат \"1s\" или число")
}

func configPathFromArgs(args []string) string {
	for i := 1; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "-c" || arg == "-config":
			if i+1 < len(args) {
				return args[i+1]
			}

		case strings.HasPrefix(arg, "-c="):
			return strings.TrimPrefix(arg, "-c=")

		case strings.HasPrefix(arg, "-config="):
			return strings.TrimPrefix(arg, "-config=")
		}
	}

	return os.Getenv("CONFIG")
}
