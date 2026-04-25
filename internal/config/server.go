package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

// ServerConfig конфигурация сервера хранилища метрик.
type ServerConfig struct {
	// RunAddr - адрес обработки запросов по работе с метриками в хранилище.
	RunAddr string `env:"ADDRESS"`
	// StoreInternal - период сохранения метрик в файловом хранилище.
	StoreInternal int64 `env:"STORE_INTERVAL"`
	// FileStoragePath - путь хранилища метрик в файле.
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	// Restore - флаг загрузки метрик из файлового хранилища.
	Restore bool `env:"RESTORE" default:"false"`
	// DSN - подключение к БД.
	DSN string `env:"DATABASE_DSN"`
	// Key - ключ проверки hash запросов.
	Key string `env:"KEY"`
	// AuditFile - путь до сохранения аудита запросов в файле.
	AuditFile string `env:"AUDIT_FILE"`
	// AuditUrl - адрес для отправки событий.
	AuditUrl string `env:"AUDIT_URL"`
}

// GetSeverConfig создает и собирает значение из флагов командной строк и env значений.
func GetSeverConfig() (*ServerConfig, error) {
	var serverConfig ServerConfig
	flag.StringVar(&serverConfig.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&serverConfig.StoreInternal, "i", 5, "time to save file storage server")
	flag.StringVar(&serverConfig.FileStoragePath, "f", "storage.json", "address file storage server")
	flag.BoolVar(&serverConfig.Restore, "r", false, "restore storage server")
	flag.StringVar(&serverConfig.DSN, "d", "", "database connection string")
	flag.StringVar(&serverConfig.Key, "k", "", "key SHA256")
	flag.StringVar(&serverConfig.AuditFile, "audit-file", "", "audit file")
	flag.StringVar(&serverConfig.AuditUrl, "audit-url", "", "audit url")
	flag.Parse()

	if err := env.Parse(&serverConfig); err != nil {
		return nil, err
	}
	return &serverConfig, nil
}
