package repository

import models "github.com/serg1732/practicum-yandex-metrics/internal/model"

type MemStorage interface {
	UpdateGauge(name string, Data models.Gauge)
	UpdateCounter(name string, Data models.Counter)
}

func BuildMemStorage() MemStorageRepository {
	return MemStorageRepository{
		MemStorage: models.MemStorage{
			CounterMap: make(map[string]models.Counter),
			GaugeMap:   make(map[string]models.Gauge),
		}}
}

type MemStorageRepository struct {
	MemStorage models.MemStorage
}

func (m MemStorageRepository) UpdateGauge(name string, Data models.Gauge) {
	m.MemStorage.GaugeMap[name] = Data
}

func (m MemStorageRepository) UpdateCounter(name string, Data models.Counter) {
	if counter, ok := m.MemStorage.CounterMap[name]; ok {
		m.MemStorage.CounterMap[name] = counter + Data
	}
}
