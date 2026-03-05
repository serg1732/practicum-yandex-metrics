package service

import (
	"context"
	"log/slog"
	"math/rand"
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

type Collector interface {
	Run(log *slog.Logger, agentConfig config.AgentConfig) error
	UpdateMetrics(ctx context.Context, log *slog.Logger, metrics <-chan map[string]*models.Metrics)
}

func BuildCollector() Collector {
	return &CollectorImpl{
		updateCounter:     atomic.Int64{},
		lastUpdateMetrics: make(map[string]*models.Metrics),
	}
}

type CollectorImpl struct {
	updateCounter     atomic.Int64
	lastUpdateMetrics map[string]*models.Metrics
	mutex             sync.RWMutex
	isNotSupportBatch bool
}

func (c *CollectorImpl) Run(log *slog.Logger, agentConfig config.AgentConfig) error {
	log.Info("Запуск сборщика метрик")
	ticks := 0
	c.updateCounter.Store(0)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	limit := agentConfig.ReportInterval + agentConfig.PollInterval
	chUpdate := make(chan *models.Metrics, 40)
	chStartUpdate := make(chan struct{})
	chMetricsToSave := make(chan map[string]*models.Metrics, 5)

	go updaterOtherMetric(ctx, log, chStartUpdate, chMetricsToSave)
	go c.UpdateMetrics(ctx, log, chMetricsToSave)

	for i := 0; i < agentConfig.RateLimit; i++ {
		go worker(ctx, log, chUpdate, repository.BuildRestyUpdaterMetric("http://"+agentConfig.RemoteAddr), agentConfig.Key)
	}
	for {
		chMetricsToSave <- getRuntimeMetrics()
		chStartUpdate <- struct{}{}
		ticks += agentConfig.PollInterval
		if ticks%limit == 0 {
			c.mutex.RLock()
			for _, metric := range c.lastUpdateMetrics {
				chUpdate <- metric
			}
			c.mutex.RUnlock()

			chUpdate <- &models.Metrics{ID: "PollCount", MType: models.Counter, Delta: getPointer(c.updateCounter.Swap(0))}
		}
		ticks %= limit
		time.Sleep(time.Duration(agentConfig.PollInterval) * time.Second)
	}
}

func getRuntimeMetrics() map[string]*models.Metrics {
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

	return metrics
}

func (c *CollectorImpl) UpdateMetrics(ctx context.Context, log *slog.Logger, metrics <-chan map[string]*models.Metrics) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-metrics:
			c.mutex.Lock()
			for k, v := range data {
				if metric, ok := c.lastUpdateMetrics[k]; ok {
					if metric != v {
						c.updateCounter.Add(1)
					}
				} else {
					c.updateCounter.Add(1)
				}
				c.lastUpdateMetrics[k] = v
			}
			c.mutex.Unlock()
		}
	}
}

func updaterOtherMetric(ctx context.Context, log *slog.Logger, startUpdate chan struct{},
	result chan<- map[string]*models.Metrics) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-startUpdate:
			v, errMemory := mem.VirtualMemory()
			if errMemory != nil {
				log.Error("ошибка при получении метрик памяти", "error", errMemory)
			}
			percent, errCPU := cpu.Percent(time.Second, false)
			if errCPU != nil {
				log.Error("ошибка при получении метрик CPU", "error", errCPU)
			}

			metrics := map[string]*models.Metrics{
				"TotalMemory":     {ID: "TotalMemory", MType: models.Gauge, Value: getPointer(float64(v.Total))},
				"FreeMemory":      {ID: "FreeMemory", MType: models.Gauge, Value: getPointer(float64(v.Free))},
				"CPUutilization1": {ID: "CPUutilization1", MType: models.Gauge, Value: getPointer(percent[0])},
			}
			result <- metrics
		}
	}
}

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

func getPointer[T float64 | int64](val T) *T {
	return &val
}
