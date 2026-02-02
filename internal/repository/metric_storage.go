package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/serg1732/practicum-yandex-metrics/internal/config"
	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
)

//go:generate mockgen -source=metric_storage.go -destination=mocks/mock_metric_storage.go -package=mocks
type MemStorage interface {
	GetCounter(name string) (*models.Metrics, bool)
	GetGauge(name string) (*models.Metrics, bool)
	Update(name string, Data *models.Metrics)
	GetAllCounters() map[string]*models.Metrics
	GetAllGauges() map[string]*models.Metrics
}

func BuildMemStorage(serverConfig *config.ServerConfig, ctx context.Context) MemStorage {
	counter := make(map[string]*models.Metrics)
	gauge := make(map[string]*models.Metrics)
	if serverConfig.Restore {
		slog.Info("Загрузка данных из файла (выставлен флаг)")
		restoreFromFile(serverConfig.FileStoragePath, gauge, counter)
	}
	repo := &MemStorageRepository{
		Config: serverConfig,
		ctx:    ctx,
		MemStorage: models.MemStorage{
			CounterMap: counter,
			GaugeMap:   gauge,
		}}
	repo.runSaver()
	return repo
}

func restoreFromFile(pathSave string, gauge map[string]*models.Metrics, counter map[string]*models.Metrics) {
	data, err := os.ReadFile(pathSave)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	var sliceData []models.Metrics
	err = json.Unmarshal(data, &sliceData)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	for _, metric := range sliceData {
		if metric.MType == models.Counter {
			counter[metric.ID] = &metric
		} else if metric.MType == models.Gauge {
			gauge[metric.ID] = &metric
		}
	}
}

type MemStorageRepository struct {
	File       *os.File
	MemStorage models.MemStorage
	Config     *config.ServerConfig
	ctx        context.Context
	mutex      sync.Mutex
}

func (m *MemStorageRepository) runSaver() {
	if m.Config.StoreInternal == 0 {
		slog.Info("Задано значение 0 между обновлениями")
		return
	}
	go func() {
		var tick int64 = 1
		for {
			select {
			case <-m.ctx.Done():
				slog.Debug("Завершение работы обновления")
				return
			default:
				time.Sleep(time.Second)
				if tick%m.Config.StoreInternal != 0 {
					slog.Debug(fmt.Sprintf("Ticks %d", tick))
					tick = (tick + 1) % m.Config.StoreInternal
					continue
				}
				slog.Debug(fmt.Sprintf("Обновление файла с хранилищем %d", tick))
				m.mutex.Lock()
				m.Save()
				m.mutex.Unlock()
			}
			tick++
		}
	}()
}

func (m *MemStorageRepository) GetAllCounters() map[string]*models.Metrics {
	return m.MemStorage.CounterMap
}

func (m *MemStorageRepository) GetAllGauges() map[string]*models.Metrics {
	return m.MemStorage.GaugeMap
}

func (m *MemStorageRepository) GetGauge(name string) (*models.Metrics, bool) {
	val, isExist := m.MemStorage.GaugeMap[name]
	return val, isExist
}

func (m *MemStorageRepository) Update(name string, Data *models.Metrics) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if Data.MType == models.Gauge {
		m.MemStorage.GaugeMap[name] = Data
	} else if Data.MType == models.Counter {
		if counter, isExist := m.MemStorage.CounterMap[name]; isExist {
			m.MemStorage.CounterMap[name].Delta = func(a *models.Metrics, b *models.Metrics) *int64 {
				if a == nil || b == nil {
					return nil
				}
				result := *a.Delta + *b.Delta
				return &result
			}(counter, Data)
		} else {
			m.MemStorage.CounterMap[name] = Data
		}
	}
	if m.Config.StoreInternal == 0 {
		m.Save()
	}
}

func (m *MemStorageRepository) GetCounter(name string) (*models.Metrics, bool) {
	counter, isExist := m.MemStorage.CounterMap[name]
	return counter, isExist
}
func (m *MemStorageRepository) Save() {
	file, _ := os.Create(m.Config.FileStoragePath)
	sliceSave := make([]models.Metrics, 0)

	for _, v := range m.MemStorage.CounterMap {
		sliceSave = append(sliceSave, *v)
	}
	for _, v := range m.MemStorage.GaugeMap {
		sliceSave = append(sliceSave, *v)
	}
	err := json.NewEncoder(file).Encode(sliceSave)
	if err != nil {
		slog.Info(err.Error())
	}
	file.Close()
}
