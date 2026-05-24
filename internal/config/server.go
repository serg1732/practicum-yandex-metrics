package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

// ServerConfig конфигурация сервера хранилища метрик.
type ServerConfig struct {
	// RunAddr - адрес обработки запросов по работе с метриками в хранилище.
	RunAddr string `env:"ADDRESS"`
	// FileStoragePath - путь хранилища метрик в файле.
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	// DSN - подключение к БД.
	DSN string `env:"DATABASE_DSN"`
	// Key - ключ проверки hash запросов.
	Key string `env:"KEY"`
	// AuditFile - путь до сохранения аудита запросов в файле.
	AuditFile string `env:"AUDIT_FILE"`
	// AuditURL - адрес для отправки событий.
	AuditURL string `env:"AUDIT_URL"`
	// CryptoKey ключ асимметричного шифрования
	CryptoKey string `env:"CRYPTO_KEY"`
	// StoreInternal - период сохранения метрик в файловом хранилище.
	StoreInternal int64 `env:"STORE_INTERVAL"`
	// Restore - флаг загрузки метрик из файлового хранилища.
	Restore bool `env:"RESTORE" default:"false"`
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
	flag.StringVar(&serverConfig.AuditURL, "audit-url", "", "audit url")
	flag.StringVar(&serverConfig.CryptoKey, "crypto-key", "", "crypto key private")
	flag.Parse()

	if err := env.Parse(&serverConfig); err != nil {
		return nil, err
	}
	return &serverConfig, nil
}
