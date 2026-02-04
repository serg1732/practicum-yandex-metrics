package repository

import (
	"context"
	"testing"

	"github.com/serg1732/practicum-yandex-metrics/internal/config"
	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestUpdateGauge(t *testing.T) {
	memStorage := BuildMemStorage(context.Background(),
		&config.ServerConfig{
			RunAddr:         "localhost:8080",
			StoreInternal:   0,
			FileStoragePath: "storage-gauge.json",
			Restore:         false,
		})

	testData := []struct {
		name          string
		gauge         *models.Metrics
		expectedExist bool
	}{
		{
			name:          "test1",
			gauge:         &models.Metrics{ID: "test1", MType: models.Gauge, Value: nil},
			expectedExist: true,
		},
		{
			name:          "test2",
			gauge:         &models.Metrics{ID: "test2", MType: models.Gauge, Value: new(float64)},
			expectedExist: true,
		},
		{
			name: "test3",
			gauge: &models.Metrics{ID: "test3", MType: models.Gauge, Value: func(f float64) *float64 {
				return &f
			}(3.14)},
			expectedExist: true,
		},
		{
			name: "test4",
			gauge: &models.Metrics{ID: "test4", MType: models.Gauge, Value: func(f float64) *float64 {
				return &f
			}(3.14)},
			expectedExist: true,
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {
			memStorage.Update(data.name, data.gauge)
			val, isExist := memStorage.GetGauge(data.name)

			assert.Equal(t, data.expectedExist, isExist)
			assert.Equal(t, data.gauge, val)
		})
	}
}

func TestUpdateCounter(t *testing.T) {
	memStorage := BuildMemStorage(context.Background(),
		&config.ServerConfig{
			RunAddr:         "localhost:8080",
			StoreInternal:   10,
			FileStoragePath: "storage-counter.json",
			Restore:         false,
		})

	testData := []struct {
		name          string
		counter       *models.Metrics
		expectedExist bool
	}{
		{
			name:          "test1",
			counter:       &models.Metrics{ID: "test1", MType: models.Counter, Delta: nil},
			expectedExist: true,
		},
		{
			name:          "test2",
			counter:       &models.Metrics{ID: "test2", MType: models.Counter, Delta: new(int64)},
			expectedExist: true,
		},
		{
			name: "test3",
			counter: &models.Metrics{ID: "test3", MType: models.Counter, Delta: func(f int64) *int64 {
				return &f
			}(314)},
			expectedExist: true,
		},
		{
			name: "test4",
			counter: &models.Metrics{ID: "test4", MType: models.Counter, Delta: func(f int64) *int64 {
				return &f
			}(-314)},
			expectedExist: true,
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {
			memStorage.Update(data.name, data.counter)
			val, isExist := memStorage.GetCounter(data.name)

			assert.Equal(t, data.expectedExist, isExist)
			assert.Equal(t, data.counter, val)
		})
	}
}
