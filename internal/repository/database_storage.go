package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/serg1732/practicum-yandex-metrics/internal/config"
	"github.com/serg1732/practicum-yandex-metrics/internal/helpers"
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
	ctx := context.Background()
	return retry(ctx, func() error {
		tx, errTx := db.database.Begin()
		if errTx != nil {
			return errTx
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

		if _, err := db.database.ExecContext(context.TODO(), query, Data.ID, Data.MType, Data.Delta, Data.Value); err != nil {
			log.Error("Ошибка при добавлении в БД", "error", err)
			return err
		}
		log.Debug("Добавлена / обновлена новая метрика", "id", Data.ID)
		return tx.Commit()
	})
}

func (db *DataBase) Updates(log *slog.Logger, Data []*models.Metrics) error {
	ctx := context.Background()
	return retry(ctx, func() error {
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
			_, errStmt := stmt.ExecContext(ctx, metric.ID, metric.MType, metric.Delta, metric.Value)
			if errStmt != nil {
				return errStmt
			}
		}

		log.Debug("Добавлены / обновлены новые метрики")
		return tx.Commit()
	})
}

func (db *DataBase) GetCounter(name string) (*models.Metrics, error) {
	ctx := context.Background()
	var row *sql.Row
	errRetry := retry(ctx, func() error {
		row = db.database.QueryRowContext(context.TODO(),
			"SELECT name, metric_type, delta FROM metrics WHERE name = $1 AND metric_type = $2", name, models.Counter)
		if row.Err() != nil {
			return row.Err()
		}
		return nil
	})
	if errRetry != nil {
		return nil, errRetry
	}
	var metrics models.Metrics
	err := row.Scan(&metrics.ID, &metrics.MType, &metrics.Delta)
	if err != nil {
		return nil, err
	}
	return &metrics, nil
}
func (db *DataBase) GetGauge(name string) (*models.Metrics, error) {
	ctx := context.Background()
	var row *sql.Row
	errRetry := retry(ctx, func() error {
		row = db.database.QueryRowContext(context.TODO(),
			"SELECT name, metric_type, value FROM metrics WHERE name = $1 AND metric_type = $2", name, models.Gauge)
		if row.Err() != nil {
			return row.Err()
		}
		return nil
	})
	if errRetry != nil {
		return nil, errRetry
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
	ctx := context.Background()
	var rows *sql.Rows
	var err error
	errRetry := retry(ctx, func() error {
		rows, err = db.database.QueryContext(context.TODO(),
			"SELECT name, metric_type, delta FROM metrics WHERE metric_type = $1", models.Counter)
		if err != nil {
			return err
		}
		if rows.Err() != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var metrics models.Metrics
			errScan := rows.Scan(&metrics.ID, &metrics.MType, &metrics.Delta)
			if errScan != nil {
				return errScan
			}
			counters[metrics.ID] = &metrics
		}
		return nil
	})
	return counters, errRetry
}
func (db *DataBase) GetAllGauges() (map[string]*models.Metrics, error) {
	gauges := make(map[string]*models.Metrics)
	ctx := context.Background()
	var rows *sql.Rows
	var err error
	errRetry := retry(ctx, func() error {
		rows, err = db.database.QueryContext(context.TODO(),
			"SELECT name, metric_type, value FROM metrics WHERE metric_type = $1", models.Gauge)
		if err != nil {
			return err
		}
		if rows.Err() != nil {
			return rows.Err()
		}
		defer rows.Close()
		for rows.Next() {
			var metrics models.Metrics
			errScan := rows.Scan(&metrics.ID, &metrics.MType, &metrics.Value)
			if errScan != nil {
				return errScan
			}
			gauges[metrics.ID] = &metrics
		}
		return nil
	})
	if errRetry != nil {
		return nil, errRetry
	}
	return gauges, nil
}

func retry(ctx context.Context, fn func() error) error {
	const maxRetries = 3
	delay := 1
	classify := helpers.NewPostgresErrorClassifier()
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		if classify.Classify(err) != helpers.Retriable {
			return err
		}
		lastErr = err

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(delay) * time.Second):
		}
		delay += 2
	}

	return fmt.Errorf("превышено число попыток запросов %s", lastErr)
}
