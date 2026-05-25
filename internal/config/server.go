package config

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v11"
)

// ServerConfig конфигурация сервера хранилища метрик.
type ServerConfig struct {
	// RunAddr - адрес обработки запросов по работе с метриками в хранилище.
	RunAddr string `env:"ADDRESS" json:"address"`
	// FileStoragePath - путь хранилища метрик в файле.
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"store_file"`
	// DSN - подключение к БД.
	DSN string `env:"DATABASE_DSN" json:"database_dsn"`
	// Key - ключ проверки hash запросов.
	Key string `env:"KEY" json:"key"`
	// AuditFile - путь до сохранения аудита запросов в файле.
	AuditFile string `env:"AUDIT_FILE" json:"audit_file"`
	// AuditURL - адрес для отправки событий.
	AuditURL string `env:"AUDIT_URL" json:"audit_url"`
	// CryptoKey ключ асимметричного шифрования
	CryptoKey string `env:"CRYPTO_KEY" json:"crypto_key"`
	// ConfigPath путь до JSON конфига
	ConfigPath string `env:"CONFIG"`
	// StoreInternal - период сохранения метрик в файловом хранилище.
	StoreInternal int `env:"STORE_INTERVAL" json:"store_internal"`
	// Restore - флаг загрузки метрик из файлового хранилища.
	Restore bool `env:"RESTORE" default:"false" json:"restore"`
}

// GetSeverConfig создает и собирает значение из флагов командной строк и env значений.
func GetSeverConfig() (*ServerConfig, error) {
	serverConfig := ServerConfig{
		RunAddr:         "localhost:8080",
		StoreInternal:   5,
		FileStoragePath: "storage.json",
		Restore:         false,
		DSN:             "",
		Key:             "",
		AuditFile:       "",
		AuditURL:        "",
		CryptoKey:       "",
	}
	path := configPathFromArgs(os.Args)
	if path != "" {
		err := applyServerJSONConfig(&serverConfig, path)
		if err != nil {
			return nil, err
		}
	}
	flag.StringVar(&serverConfig.RunAddr, "a", serverConfig.RunAddr, "address and port to run server")
	flag.IntVar(&serverConfig.StoreInternal, "i", serverConfig.StoreInternal, "time to save file storage server")
	flag.StringVar(&serverConfig.FileStoragePath, "f", serverConfig.FileStoragePath, "address file storage server")
	flag.BoolVar(&serverConfig.Restore, "r", serverConfig.Restore, "restore storage server")
	flag.StringVar(&serverConfig.DSN, "d", serverConfig.DSN, "database connection string")
	flag.StringVar(&serverConfig.Key, "k", serverConfig.Key, "key SHA256")
	flag.StringVar(&serverConfig.AuditFile, "audit-file", serverConfig.AuditFile, "audit file")
	flag.StringVar(&serverConfig.AuditURL, "audit-url", serverConfig.AuditURL, "audit url")
	flag.StringVar(&serverConfig.CryptoKey, "crypto-key", serverConfig.CryptoKey, "crypto key private")
	flag.StringVar(&serverConfig.ConfigPath, "c", serverConfig.ConfigPath, "config file path")
	flag.StringVar(&serverConfig.ConfigPath, "config", serverConfig.ConfigPath, "config file path")

	flag.Parse()

	if err := env.Parse(&serverConfig); err != nil {
		return nil, err
	}
	return &serverConfig, nil
}
