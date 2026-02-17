package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/serg1732/practicum-yandex-metrics/internal/config"
	models "github.com/serg1732/practicum-yandex-metrics/internal/model"

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
	return &DataBase{database: db}, nil
}

func MigrateDataBase(log *slog.Logger, config *config.ServerConfig) error {
	if config.DSN == "" {
		log.Error("Конфиг ДБ пустой")
		return errors.New("DSL required")
	}
	m, err := migrate.New("file://migrations", config.DSN)
	if errUp := m.Up(); errUp != nil {
		log.Error("Не удалось <<апнуть>> БД", "error", errUp)
	}
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
	database *sql.DB
}

func (db *DataBase) Ping() error {
	return db.database.Ping()
}

func (db *DataBase) Update(log *slog.Logger, Data *models.Metrics) error {
	tx, _ := db.database.Begin()
	query := `
	INSERT INTO metrics (name, metric_type, delta, value)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (name, metric_type)
	DO UPDATE SET
	  delta = CASE
		WHEN EXCLUDED.delta IS NOT NULL
		  THEN COALESCE(metrics.delta, 0) + EXCLUDED.delta
		ELSE metrics.delta
	  END,
	  value = CASE
		WHEN EXCLUDED.value IS NOT NULL
		  THEN EXCLUDED.value
		ELSE metrics.value
	  END;
	`

	if _, err := db.database.ExecContext(context.TODO(), query, Data.ID, Data.MType, Data.Delta, Data.Value); err != nil {
		tx.Rollback()
		log.Error("Ошибка при добавлении в БД", "error", err)
		return err
	}
	log.Debug("Добавлена / обновлена новая метрика", "id", Data.ID)
	tx.Commit()
	return nil
}

func (db *DataBase) Updates(log *slog.Logger, Data []*models.Metrics) error {
	tx, err := db.database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
	INSERT INTO metrics (name, metric_type, delta, value)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (name, metric_type)
	DO UPDATE SET
	  delta = CASE
		WHEN EXCLUDED.delta IS NOT NULL
		  THEN COALESCE(metrics.delta, 0) + EXCLUDED.delta
		ELSE metrics.delta
	  END,
	  value = CASE
		WHEN EXCLUDED.value IS NOT NULL
		  THEN EXCLUDED.value
		ELSE metrics.value
	  END;
	`

	stmt, err := db.database.Prepare(query)
	if err != nil {
		return err
	}

	for _, metric := range Data {
		res, errStmt := stmt.ExecContext(context.TODO(), metric.ID, metric.MType, metric.Delta, metric.Value)
		count, errAffected := res.RowsAffected()
		if errAffected != nil {
			return errAffected
		}

		log.Debug("Затронуто строк в БД", "count", count)

		if errStmt != nil {
			return errStmt
		}
	}

	log.Debug("Добавлены / обновлены новые метрики")
	return tx.Commit()
}

func (db *DataBase) GetCounter(name string) (*models.Metrics, error) {
	tx, errTx := db.database.Begin()
	if errTx != nil {
		return nil, errTx
	}
	defer tx.Rollback()
	row := db.database.QueryRowContext(context.TODO(),
		"SELECT name, metric_type, delta FROM metrics WHERE name = $1 AND metric_type = $2", name, models.Counter)
	if row.Err() != nil {
		return nil, row.Err()
	}
	var metrics models.Metrics
	err := row.Scan(&metrics.ID, &metrics.MType, &metrics.Delta)
	if err != nil {
		return nil, err
	}
	tx.Commit()
	return &metrics, nil
}
func (db *DataBase) GetGauge(name string) (*models.Metrics, error) {
	row := db.database.QueryRowContext(context.TODO(),
		"SELECT name, metric_type, value FROM metrics WHERE name = $1 AND metric_type = $2", name, models.Gauge)
	if row.Err() != nil {
		return nil, row.Err()
	}
	var metrics models.Metrics
	err := row.Scan(&metrics.ID, &metrics.MType, &metrics.Value)
	if err != nil {
		return nil, err
	}
	return &metrics, nil
}
func (db *DataBase) GetAllCounters() (map[string]*models.Metrics, error) {
	counters := make(map[string]*models.Metrics)
	rows, err := db.database.QueryContext(context.TODO(),
		"SELECT name, metric_type, delta FROM metrics WHERE metric_type = $1", models.Counter)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var metrics models.Metrics
		errScan := rows.Scan(&metrics.ID, &metrics.MType, &metrics.Delta)
		if errScan != nil {
			return nil, errScan
		}
		counters[metrics.ID] = &metrics
	}
	return counters, nil
}
func (db *DataBase) GetAllGauges() (map[string]*models.Metrics, error) {
	gauges := make(map[string]*models.Metrics)
	rows, err := db.database.QueryContext(context.TODO(),
		"SELECT name, metric_type, value FROM metrics WHERE metric_type = $1", models.Gauge)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer rows.Close()
	for rows.Next() {
		var metrics models.Metrics
		errScan := rows.Scan(&metrics.ID, &metrics.MType, &metrics.Value)
		if errScan != nil {
			return nil, errScan
		}
		gauges[metrics.ID] = &metrics
	}
	return gauges, nil
}
