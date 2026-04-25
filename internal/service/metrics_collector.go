package service

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/serg1732/practicum-yandex-metrics/internal/config"
	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// Collector представляет интерфейс, отражающий реализацию сборщика(агента) метрик.
type Collector interface {
	// Run функция запуска агента-сборщика метрик.
	// Возвращает ошибку, если не удалось запустить.
	Run(ctx context.Context, log *slog.Logger, agentConfig config.AgentConfig) error
	// UpdateMetrics функция обновления локального хранилища метрик агента.
	UpdateMetrics(log *slog.Logger, metrics map[string]*models.Metrics)
}

// BuildCollector функция создания агента сборщика.
func BuildCollector() Collector {
	return &CollectorImpl{
		updateCounter:     atomic.Int64{},
		lastUpdateMetrics: make(map[string]*models.Metrics),
	}
}

// CollectorImpl агент-сборщик метрик.
type CollectorImpl struct {
	// updateCounter счетчик обновленных метрик.
	updateCounter atomic.Int64
	// lastUpdateMetrics сохраняет / обновляет собираемые метрики.
	lastUpdateMetrics map[string]*models.Metrics
	// mutex мьютех синхронизации обновления метрик.
	mutex sync.RWMutex
	// isNotSupportBatch поддерживается ли обработка батчами.
	isNotSupportBatch bool
}

// Run функция запуска агента-сборщика метрик.
func (c *CollectorImpl) Run(ctx context.Context, log *slog.Logger, agentConfig config.AgentConfig) error {
	log.Info("Запуск сборщика метрик")
	ticks := 0
	c.updateCounter.Store(0)

	chMetricsToSave := make(chan map[string]*models.Metrics)
	go updater(ctx, log, agentConfig, chMetricsToSave, getRuntimeMetrics)
	go updater(ctx, log, agentConfig, chMetricsToSave, getAdditionalMetrics)

	chUpdate := make(chan *models.Metrics, agentConfig.RateLimit)
	for i := 0; i < agentConfig.RateLimit; i++ {
		go worker(ctx, log, chUpdate, repository.BuildRestyUpdaterMetric("http://"+agentConfig.RemoteAddr), agentConfig.Key)
	}
	ticker := time.NewTicker(time.Duration(agentConfig.PollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if !errors.Is(ctx.Err(), context.Canceled) {
				return ctx.Err()
			}
			return nil
		case data := <-chMetricsToSave:
			c.UpdateMetrics(log, data)
		case <-ticker.C:
			ticks += agentConfig.PollInterval
			if ticks%agentConfig.ReportInterval == 0 {
				c.mutex.RLock()
				for _, metric := range c.lastUpdateMetrics {
					chUpdate <- metric
				}
				c.mutex.RUnlock()

				chUpdate <- &models.Metrics{ID: "PollCount", MType: models.Counter, Delta: getPointer(c.updateCounter.Swap(0))}
			}
			ticks %= agentConfig.ReportInterval
		}
	}
}

// getRuntimeMetrics функция съема метрик системы.
func getRuntimeMetrics() (map[string]*models.Metrics, error) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	metrics := map[string]*models.Metrics{
		"Alloc":         {ID: "Alloc", MType: models.Gauge, Value: getPointer(float64(ms.Alloc))},
		"BuckHashSys":   {ID: "BuckHashSys", MType: models.Gauge, Value: getPointer(float64(ms.BuckHashSys))},
		"Frees":         {ID: "Frees", MType: models.Gauge, Value: getPointer(float64(ms.Frees))},
		"GCCPUFraction": {ID: "GCCPUFraction", MType: models.Gauge, Value: getPointer(float64(ms.GCCPUFraction))},
		"GCSys":         {ID: "GCSys", MType: models.Gauge, Value: getPointer(float64(ms.GCSys))},
		"HeapAlloc":     {ID: "HeapAlloc", MType: models.Gauge, Value: getPointer(float64(ms.HeapAlloc))},
		"HeapIdle":      {ID: "HeapIdle", MType: models.Gauge, Value: getPointer(float64(ms.HeapIdle))},
		"HeapInuse":     {ID: "HeapInuse", MType: models.Gauge, Value: getPointer(float64(ms.HeapInuse))},
		"HeapObjects":   {ID: "HeapObjects", MType: models.Gauge, Value: getPointer(float64(ms.HeapObjects))},
		"HeapReleased":  {ID: "HeapReleased", MType: models.Gauge, Value: getPointer(float64(ms.HeapReleased))},
		"HeapSys":       {ID: "HeapSys", MType: models.Gauge, Value: getPointer(float64(ms.HeapSys))},
		"LastGC":        {ID: "LastGC", MType: models.Gauge, Value: getPointer(float64(ms.LastGC))},
		"Lookups":       {ID: "Lookups", MType: models.Gauge, Value: getPointer(float64(ms.Lookups))},
		"MCacheInuse":   {ID: "MCacheInuse", MType: models.Gauge, Value: getPointer(float64(ms.MCacheInuse))},
		"MCacheSys":     {ID: "MCacheSys", MType: models.Gauge, Value: getPointer(float64(ms.MCacheSys))},
		"MSpanInuse":    {ID: "MSpanInuse", MType: models.Gauge, Value: getPointer(float64(ms.MSpanInuse))},
		"MSpanSys":      {ID: "MSpanSys", MType: models.Gauge, Value: getPointer(float64(ms.MSpanSys))},
		"Mallocs":       {ID: "Mallocs", MType: models.Gauge, Value: getPointer(float64(ms.Mallocs))},
		"NextGC":        {ID: "NextGC", MType: models.Gauge, Value: getPointer(float64(ms.NextGC))},
		"NumForcedGC":   {ID: "NumForcedGC", MType: models.Gauge, Value: getPointer(float64(ms.NumForcedGC))},
		"NumGC":         {ID: "NumGC", MType: models.Gauge, Value: getPointer(float64(ms.NumGC))},
		"OtherSys":      {ID: "OtherSys", MType: models.Gauge, Value: getPointer(float64(ms.OtherSys))},
		"PauseTotalNs":  {ID: "PauseTotalNs", MType: models.Gauge, Value: getPointer(float64(ms.PauseTotalNs))},
		"StackInuse":    {ID: "StackInuse", MType: models.Gauge, Value: getPointer(float64(ms.StackInuse))},
		"StackSys":      {ID: "StackSys", MType: models.Gauge, Value: getPointer(float64(ms.StackSys))},
		"Sys":           {ID: "Sys", MType: models.Gauge, Value: getPointer(float64(ms.Sys))},
		"TotalAlloc":    {ID: "TotalAlloc", MType: models.Gauge, Value: getPointer(float64(ms.TotalAlloc))},
		"RandomValue":   {ID: "RandomValue", MType: models.Gauge, Value: getPointer(rand.Float64())},
	}

	return metrics, nil
}

// getAdditionalMetrics получение дополнительных метрик TotalMemory, FreeMemory, CPUutilization1.
func getAdditionalMetrics() (map[string]*models.Metrics, error) {
	v, errMemory := mem.VirtualMemory()
	if errMemory != nil {
		return nil, errMemory
	}
	percent, errCPU := cpu.Percent(time.Second, false)
	if errCPU != nil {
		return nil, errMemory
	}

	metrics := map[string]*models.Metrics{
		"TotalMemory":     {ID: "TotalMemory", MType: models.Gauge, Value: getPointer(float64(v.Total))},
		"FreeMemory":      {ID: "FreeMemory", MType: models.Gauge, Value: getPointer(float64(v.Free))},
		"CPUutilization1": {ID: "CPUutilization1", MType: models.Gauge, Value: getPointer(percent[0])},
	}
	return metrics, nil
}

// UpdateMetrics функция обновления метрик и подсчета их количества.
func (c *CollectorImpl) UpdateMetrics(log *slog.Logger, metrics map[string]*models.Metrics) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for k, v := range metrics {
		if metric, ok := c.lastUpdateMetrics[k]; ok {
			if metric != v {
				c.updateCounter.Add(1)
			}
		} else {
			c.updateCounter.Add(1)
		}
		c.lastUpdateMetrics[k] = v
		log.Debug("успешно обновлены метрики")
	}
}

// updater функция обработки сигналов обновления по таймеру метрик.
func updater(ctx context.Context, log *slog.Logger, config config.AgentConfig,
	metrics chan<- map[string]*models.Metrics, getMetricsFunc func() (map[string]*models.Metrics, error)) {
	ticker := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Debug("завершение работы горутины с обновлением дополнительных метрик")
			return
		case <-ticker.C:
			data, err := getMetricsFunc()
			if err != nil {
				log.Error("ошибка получения метрик", "error", err)
				os.Exit(1)
			}
			metrics <- data
		}
	}
}

// worker обработчик канала по отправке метрик на сервер.
// По сигналу updateChannel отправляет данные на сервер.
func worker(ctx context.Context, log *slog.Logger, updateChannel <-chan *models.Metrics,
	client repository.UpdaterClient, key string,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case metric := <-updateChannel:
			err := client.ExternalUpdateMetric(ctx, log, key, metric)
			if err != nil {
				log.Error("ошибка при обновлении метрики", "error", err)
			}
		}
	}
}

// getPointer функция получения из примитивного типа в указатель.
// Возможные значения float64 или int64.
func getPointer[T float64 | int64](val T) *T {
	return &val
}
