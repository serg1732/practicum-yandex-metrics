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

func BuildMemStorage(ctx context.Context, log *slog.Logger, serverConfig *config.ServerConfig) *MemStorageRepository {
	counter := make(map[string]*models.Metrics)
	gauge := make(map[string]*models.Metrics)
	if serverConfig.Restore {
		log.Info("Загрузка данных из файла (выставлен флаг)")
		restoreFromFile(log, serverConfig.FileStoragePath, gauge, counter)
	}
	repo := &MemStorageRepository{
		Config: serverConfig,
		ctx:    ctx,
		MemStorage: models.MemStorage{
			CounterMap: counter,
			GaugeMap:   gauge,
		}}
	repo.runSaver(log)
	return repo
}

type MemStorageRepository struct {
	File       *os.File
	MemStorage models.MemStorage
	Config     *config.ServerConfig
	ctx        context.Context
	mutex      sync.Mutex
}

func (m *MemStorageRepository) runSaver(log *slog.Logger) {
	if m.Config.StoreInternal == 0 {
		log.Info("Задано значение 0 между обновлениями")
		return
	}
	go func() {
		var tick int64 = 1
		for {
			select {
			case <-m.ctx.Done():
				log.Debug("Завершение работы обновления")
				return
			default:
				time.Sleep(time.Second)
				if tick%m.Config.StoreInternal != 0 {
					log.Debug(fmt.Sprintf("Ticks %d", tick))
					tick = (tick + 1) % m.Config.StoreInternal
					continue
				}
				log.Debug(fmt.Sprintf("Обновление файла с хранилищем %d", tick))
				m.mutex.Lock()
				m.Save(log)
				m.mutex.Unlock()
			}
			tick++
		}
	}()
}

func (m *MemStorageRepository) GetAllCounters() (map[string]*models.Metrics, error) {
	return m.MemStorage.CounterMap, nil
}

func (m *MemStorageRepository) GetAllGauges() (map[string]*models.Metrics, error) {
	return m.MemStorage.GaugeMap, nil
}

func (m *MemStorageRepository) GetGauge(name string) (*models.Metrics, error) {
	val, isExist := m.MemStorage.GaugeMap[name]
	if !isExist {
		return nil, nil
	}
	return val, nil
}

func (m *MemStorageRepository) Update(log *slog.Logger, Data *models.Metrics) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if Data.MType == models.Gauge {
		m.MemStorage.GaugeMap[Data.ID] = Data
	} else if Data.MType == models.Counter {
		if counter, isExist := m.MemStorage.CounterMap[Data.ID]; isExist {
			m.MemStorage.CounterMap[Data.ID].Delta = func(a *models.Metrics, b *models.Metrics) *int64 {
				if a == nil || b == nil {
					return nil
				}
				result := *a.Delta + *b.Delta
				return &result
			}(counter, Data)
		} else {
			m.MemStorage.CounterMap[Data.ID] = Data
		}
	}
	if m.Config.StoreInternal == 0 {
		m.Save(log)
	}
	return nil
}

func (m *MemStorageRepository) GetCounter(name string) (*models.Metrics, error) {
	counter, isExist := m.MemStorage.CounterMap[name]
	if !isExist {
		return nil, nil
	}
	return counter, nil
}
func (m *MemStorageRepository) Save(log *slog.Logger) {
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
		log.Error("Ошибка при инициализации записи json в файл:", "error", err.Error())
	}
	file.Close()
}

func restoreFromFile(log *slog.Logger, pathSave string, gauge map[string]*models.Metrics, counter map[string]*models.Metrics) {
	data, err := os.ReadFile(pathSave)
	if err != nil {
		log.Error("Ошибка при чтении файла хранилища", "error", err.Error())
		return
	}
	var sliceData []models.Metrics
	err = json.Unmarshal(data, &sliceData)
	if err != nil {
		log.Error("Ошибка при обработке файла хранилища", "error", err.Error())
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
