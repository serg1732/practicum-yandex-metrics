package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5"
	"github.com/serg1732/practicum-yandex-metrics/internal/config"
	"github.com/serg1732/practicum-yandex-metrics/internal/helpers"
	models "github.com/serg1732/practicum-yandex-metrics/internal/model"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

func BuildDataBase(ctx context.Context, log *slog.Logger, config *config.ServerConfig) (*DataBase, error) {
	if config.DSN == "" {
		return nil, errors.New("DSL required")
	}
	db, err := pgxpool.New(ctx, config.DSN)
	if err != nil {
		log.Error("Error opening database connection", "error", err)
		return nil, err
	}
	if errPing := db.Ping(ctx); errPing != nil {
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

type Pool interface {
	Ping(ctx context.Context) error
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	Close()
}

type DataBase struct {
	database Pool
}

func (db *DataBase) Ping(ctx context.Context) error {
	return db.database.Ping(ctx)
}

func (db *DataBase) Update(ctx context.Context, log *slog.Logger, Data *models.Metrics) error {
	return db.retry(ctx, func(tx pgx.Tx) error {
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

		if _, err := tx.Exec(ctx, query, Data.ID, Data.MType, Data.Delta, Data.Value); err != nil {
			log.Error("Ошибка при добавлении в БД", "error", err)
			return err
		}
		log.Debug("Добавлена / обновлена новая метрика", "id", Data.ID)
		return nil
	})
}

func (db *DataBase) Updates(ctx context.Context, log *slog.Logger, Data []*models.Metrics) error {
	return db.retry(ctx, func(tx pgx.Tx) error {
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
		const preparedName string = "insert_metrics"
		_, err := tx.Prepare(ctx, preparedName, query)
		if err != nil {
			return err
		}

		for _, metric := range Data {
			_, errStmt := tx.Exec(ctx, preparedName, metric.ID, metric.MType, metric.Delta, metric.Value)
			if errStmt != nil {
				return errStmt
			}
		}

		log.Debug("Добавлены / обновлены новые метрики")
		return nil
	})
}

func (db *DataBase) GetCounter(ctx context.Context, name string) (*models.Metrics, error) {
	var row pgx.Row
	var metrics models.Metrics
	errRetry := db.retry(ctx, func(tx pgx.Tx) error {
		row = tx.QueryRow(ctx, `SELECT name, metric_type, delta FROM metrics WHERE name = $1 AND metric_type = $2`, name, models.Counter)
		if errScan := row.Scan(&metrics.ID, &metrics.MType, &metrics.Delta); errScan != nil {
			return errScan
		}
		return nil
	})
	if errRetry != nil {
		return nil, errRetry
	}
	return &metrics, nil
}
func (db *DataBase) GetGauge(ctx context.Context, name string) (*models.Metrics, error) {
	var row pgx.Row
	var metrics models.Metrics
	errRetry := db.retry(ctx, func(tx pgx.Tx) error {
		row = tx.QueryRow(ctx, `SELECT name, metric_type, value FROM metrics WHERE name = $1 AND metric_type = $2`, name, models.Gauge)
		if errScan := row.Scan(&metrics.ID, &metrics.MType, &metrics.Value); errScan != nil {
			return errScan
		}
		return nil
	})
	if errRetry != nil {
		return nil, errRetry
	}

	return &metrics, nil
}
func (db *DataBase) GetAllCounters(ctx context.Context) (map[string]*models.Metrics, error) {
	counters := make(map[string]*models.Metrics)
	errRetry := db.retry(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			"SELECT name, metric_type, delta FROM metrics WHERE metric_type = $1", models.Counter,
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		if rows.Err() != nil {
			return err
		}
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
func (db *DataBase) GetAllGauges(ctx context.Context) (map[string]*models.Metrics, error) {
	gauges := make(map[string]*models.Metrics)
	errRetry := db.retry(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`SELECT name, metric_type, value FROM metrics WHERE metric_type = $1`, models.Gauge,
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		if rows.Err() != nil {
			return rows.Err()
		}
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

func (db *DataBase) retry(ctx context.Context, fn func(pgx.Tx) error) error {
	const maxRetries = 3
	delay := 1
	classify := helpers.NewPostgresErrorClassifier()
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		tx, err := db.database.BeginTx(ctx, pgx.TxOptions{})
		defer tx.Rollback(ctx)
		if err != nil && classify.Classify(err) != helpers.Retriable {
			return err
		}
		errFn := fn(tx)
		if errFn == nil {
			if errCommit := tx.Commit(ctx); errCommit == nil {
				return nil
			} else {
				if classify == nil || classify.Classify(errCommit) != helpers.Retriable {
					return errCommit
				}
				lastErr = errCommit
			}
		} else {
			if classify == nil || classify.Classify(errFn) != helpers.Retriable {
				return errFn
			}
			lastErr = errFn
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(delay) * time.Second):
		}
		delay += 2
	}

	return fmt.Errorf("превышено число попыток запросов %s", lastErr)
}
