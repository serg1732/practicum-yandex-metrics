package repository

import (
	"context"
	"log/slog"
	"testing"

	"github.com/serg1732/practicum-yandex-metrics/internal/config"
	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestUpdateGauge(t *testing.T) {
	memStorage := BuildMemStorage(context.Background(), slog.Default(),
		&config.ServerConfig{
			RunAddr:         "localhost:8080",
			StoreInternal:   0,
			FileStoragePath: "storage-gauge.json",
			Restore:         false,
		})

	testData := []struct {
		name  string
		gauge *models.Metrics
	}{
		{
			name:  "test1",
			gauge: &models.Metrics{ID: "test1", MType: models.Gauge, Value: nil},
		},
		{
			name:  "test2",
			gauge: &models.Metrics{ID: "test2", MType: models.Gauge, Value: new(float64)},
		},
		{
			name: "test3",
			gauge: &models.Metrics{ID: "test3", MType: models.Gauge, Value: func(f float64) *float64 {
				return &f
			}(3.14)},
		},
		{
			name: "test4",
			gauge: &models.Metrics{ID: "test4", MType: models.Gauge, Value: func(f float64) *float64 {
				return &f
			}(3.14)},
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {
			errUpdate := memStorage.Update(slog.Default(), data.gauge)
			val, err := memStorage.GetGauge(data.name)

			assert.Nil(t, err)
			assert.Nil(t, errUpdate)
			if assert.NotNil(t, data.gauge) {
				assert.Equal(t, data.gauge, val)
			}
		})
	}
}

func TestUpdateCounter(t *testing.T) {
	memStorage := BuildMemStorage(context.Background(), slog.Default(),
		&config.ServerConfig{
			RunAddr:         "localhost:8080",
			StoreInternal:   10,
			FileStoragePath: "storage-counter.json",
			Restore:         false,
		})

	testData := []struct {
		name    string
		counter *models.Metrics
	}{
		{
			name:    "test1",
			counter: &models.Metrics{ID: "test1", MType: models.Counter, Delta: nil},
		},
		{
			name:    "test2",
			counter: &models.Metrics{ID: "test2", MType: models.Counter, Delta: new(int64)},
		},
		{
			name: "test3",
			counter: &models.Metrics{ID: "test3", MType: models.Counter, Delta: func(f int64) *int64 {
				return &f
			}(314)},
		},
		{
			name: "test4",
			counter: &models.Metrics{ID: "test4", MType: models.Counter, Delta: func(f int64) *int64 {
				return &f
			}(-314)},
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {
			errUpdate := memStorage.Update(slog.Default(), data.counter)
			val, err := memStorage.GetCounter(data.name)
			assert.Nil(t, err)
			assert.Nil(t, errUpdate)
			if assert.NotNil(t, data.counter) {
				assert.Equal(t, data.counter, val)
			}
		})
	}
}
