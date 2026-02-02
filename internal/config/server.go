package config

type ServerConfig struct {
	RunAddr         string `env:"ADDRESS"`
	StoreInternal   int64  `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE" default:"false"`
}
