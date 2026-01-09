package service

import (
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/serg1732/practicum-yandex-metrics/internal/repository"
)

type Collector interface {
	Run(poolInterval int, reportInterval int) error
	UpdateMetrics(metrics map[string]float64) int64
}

func BuildCollector(url string) Collector {
	return CollectorImpl{
		updateCounter:     0,
		lastUpdateMetrics: make(map[string]float64),
		updaterClient:     repository.BuildRestyUpdaterMetric(url),
	}
}

type CollectorImpl struct {
	updateCounter     int64
	lastUpdateMetrics map[string]float64
	updaterClient     repository.IUpdaterClient
}

func (c CollectorImpl) Run(poolInterval int, reportInterval int) error {
	log.Println("Starting metrics collector")
	ticks := 0
	for {
		metrics := getRuntimeMetrics()
		c.updateCounter = c.UpdateMetrics(metrics)
		c.lastUpdateMetrics = metrics
		ticks += poolInterval
		if ticks%reportInterval == 0 {
			err := c.updaterClient.ExternalUpdateMetrics(c.updateCounter, c.lastUpdateMetrics)
			if err != nil {
				return err
			}
			c.updateCounter = 0
		}
		ticks %= reportInterval
		time.Sleep(time.Duration(poolInterval) * time.Second)
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

func (c CollectorImpl) UpdateMetrics(metrics map[string]float64) int64 {
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
