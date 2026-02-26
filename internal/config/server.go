package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

type ServerConfig struct {
	RunAddr         string `env:"ADDRESS"`
	StoreInternal   int64  `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE" default:"false"`
	DSN             string `env:"DATABASE_DSN"`
}

// GetSeverConfig создает и собирает значение из флагов командной строк и env значений.
func GetSeverConfig() (*ServerConfig, error) {
	var serverConfig ServerConfig
	flag.StringVar(&serverConfig.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&serverConfig.StoreInternal, "i", 5, "time to save file storage server")
	flag.StringVar(&serverConfig.FileStoragePath, "f", "storage.json", "address file storage server")
	flag.BoolVar(&serverConfig.Restore, "r", false, "restore storage server")
	flag.StringVar(&serverConfig.DSN, "d", "", "database connection string")
	flag.Parse()

	if err := env.Parse(&serverConfig); err != nil {
		return nil, err
	}
	return &serverConfig, nil
}
