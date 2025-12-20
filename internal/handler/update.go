package handler

import (
	"net/http"
	"strconv"

	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"
)

var MemStorage = repository.BuildMemStorage()

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	metricType := r.PathValue("metricType")
	metricName := r.PathValue("metricName")
	metricValue := r.PathValue("metricValue")

	if metricType == "gauge" {
		if err := processGauge(metricName, metricValue); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else if metricType == "counter" {
		if err := processCounter(metricName, metricValue); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}
}

func processCounter(key string, value string) error {
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}
	MemStorage.UpdateGauge(key, models.Gauge(val))
	return nil
}
func processGauge(key string, value string) error {
	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}
	MemStorage.UpdateCounter(key, models.Counter(val))
	return nil
}
