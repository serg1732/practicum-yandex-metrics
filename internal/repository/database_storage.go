package repository

import (
	"database/sql"
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/serg1732/practicum-yandex-metrics/internal/config"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func BuildDataBase(log *slog.Logger, config *config.ServerConfig) (*DataBase, error) {
	if config.DSN == "" {
		return nil, errors.New("DSL required")
	}
	db, err := sql.Open("postgres", config.DSN)
	if err != nil {
		log.Error("Error opening database connection", "error", err)
		return nil, err
	}
	if errPing := db.Ping(); errPing != nil {
		log.Error("Error pinging database connection", "error", errPing)
		return nil, errPing
	}

	log.Info("Successfully connected to database")
	return &DataBase{db}, nil
}

func MigrateDataBase(log *slog.Logger, config *config.ServerConfig) error {
	m, err := migrate.New("file://migrations", config.DSN)
	if err != nil {
		log.Error("ошибка миграции", "error", err)
		return err
	}

	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			log.Error("ошибка миграции", "error", err)
			return err
		}
		log.Error("Ошибка при миграции", "error", err)
		return err
	}

	log.Info("Current version:", "version", version)
	log.Info("Dirty:", "dirty", dirty)
	return nil
}

type DataBase struct {
	db *sql.DB
}

func (db *DataBase) Ping() error {
	return db.db.Ping()
}
