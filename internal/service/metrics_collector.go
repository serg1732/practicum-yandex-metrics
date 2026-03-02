package service

import (
	"log/slog"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/serg1732/practicum-yandex-metrics/internal/config"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"
)

type Collector interface {
	Run(log *slog.Logger, agentConfig config.AgentConfig) error
	UpdateMetrics(metrics map[string]float64) int64
}

func BuildCollector(agentConfig config.AgentConfig) Collector {
	return &CollectorImpl{
		updateCounter:     0,
		lastUpdateMetrics: make(map[string]float64),
		updaterClient:     repository.BuildRestyUpdaterMetric("http://" + agentConfig.RemoteAddr),
	}
}

type CollectorImpl struct {
	updateCounter     int64
	lastUpdateMetrics map[string]float64
	updaterClient     repository.UpdaterClient
	mutex             sync.Mutex
	isNotSupportBatch bool
}

func (c *CollectorImpl) Run(log *slog.Logger, agentConfig config.AgentConfig) error {
	log.Info("Запуск сборщика метрик")
	ticks := 0
	limit := agentConfig.ReportInterval + agentConfig.PollInterval
	for {
		metrics := getRuntimeMetrics()
		c.updateCounter = c.UpdateMetrics(metrics)
		c.lastUpdateMetrics = metrics
		ticks += agentConfig.PollInterval
		if ticks%limit == 0 {
			if !c.isNotSupportBatch {
				if err := c.updaterClient.ExternalBatchUpdateJSONMetrics(
					log, c.updateCounter, c.lastUpdateMetrics, agentConfig.Key); err != nil {
					log.Error("API не поддерживает /updates/")
					c.isNotSupportBatch = true
				}
			}
			if c.isNotSupportBatch {
				err := c.updaterClient.ExternalUpdateJSONMetrics(
					log, c.updateCounter, c.lastUpdateMetrics, agentConfig.Key,
				)
				if err != nil {
					log.Error("Ошибка при обновлении метрик ", "error", err)
				}
			}
			c.updateCounter = 0
		}
		ticks %= limit
		time.Sleep(time.Duration(agentConfig.PollInterval) * time.Second)
	}
}

func getRuntimeMetrics() map[string]float64 {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	metrics := map[string]float64{
		"Alloc":         float64(ms.Alloc),
		"BuckHashSys":   float64(ms.BuckHashSys),
		"Frees":         float64(ms.Frees),
		"GCCPUFraction": ms.GCCPUFraction,
		"GCSys":         float64(ms.GCSys),
		"HeapAlloc":     float64(ms.HeapAlloc),
		"HeapIdle":      float64(ms.HeapIdle),
		"HeapInuse":     float64(ms.HeapInuse),
		"HeapObjects":   float64(ms.HeapObjects),
		"HeapReleased":  float64(ms.HeapReleased),
		"HeapSys":       float64(ms.HeapSys),
		"LastGC":        float64(ms.LastGC),
		"Lookups":       float64(ms.Lookups),
		"MCacheInuse":   float64(ms.MCacheInuse),
		"MCacheSys":     float64(ms.MCacheSys),
		"MSpanInuse":    float64(ms.MSpanInuse),
		"MSpanSys":      float64(ms.MSpanSys),
		"Mallocs":       float64(ms.Mallocs),
		"NextGC":        float64(ms.NextGC),
		"NumForcedGC":   float64(ms.NumForcedGC),
		"NumGC":         float64(ms.NumGC),
		"OtherSys":      float64(ms.OtherSys),
		"PauseTotalNs":  float64(ms.PauseTotalNs),
		"StackInuse":    float64(ms.StackInuse),
		"StackSys":      float64(ms.StackSys),
		"Sys":           float64(ms.Sys),
		"TotalAlloc":    float64(ms.TotalAlloc),
		"RandomValue":   rand.Float64(),
	}
	return metrics
}

func (c *CollectorImpl) UpdateMetrics(metrics map[string]float64) int64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	updateCounter := c.updateCounter
	for k, v := range metrics {
		if metric, ok := c.lastUpdateMetrics[k]; ok {
			if metric != v {
				updateCounter++
			}
		} else {
			updateCounter++
		}
		c.lastUpdateMetrics[k] = v
	}
	return updateCounter
}
