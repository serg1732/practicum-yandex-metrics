package repository

import models "github.com/serg1732/practicum-yandex-metrics/internal/model"

type MemStorage interface {
	GetCounter(name string) (*int64, bool)
	GetGauge(name string) (*float64, bool)
	UpdateCounter(name string, Data *int64)
	UpdateGauge(name string, Data *float64)
}

func BuildMemStorage() MemStorage {
	return MemStorageRepository{
		MemStorage: models.MemStorage{
			CounterMap: make(map[string]*int64),
			GaugeMap:   make(map[string]*float64),
		}}
}

type MemStorageRepository struct {
	MemStorage models.MemStorage
}

func (m MemStorageRepository) GetGauge(name string) (*float64, bool) {
	val, isExist := m.MemStorage.GaugeMap[name]
	return val, isExist
}

func (m MemStorageRepository) UpdateGauge(name string, Data *float64) {
	m.MemStorage.GaugeMap[name] = Data
}

func (m MemStorageRepository) GetCounter(name string) (*int64, bool) {
	counter, isExist := m.MemStorage.CounterMap[name]
	return counter, isExist
}

func (m MemStorageRepository) UpdateCounter(name string, data *int64) {
	if counter, isExist := m.MemStorage.CounterMap[name]; isExist {
		m.MemStorage.CounterMap[name] = func(a *int64, b *int64) *int64 {
			if a == nil || b == nil {
				return nil
			}
			result := *a + *b
			return &result
		}(counter, data)
	} else {
		m.MemStorage.CounterMap[name] = data
	}
}
