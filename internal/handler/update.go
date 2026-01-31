package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"
)

type UpdateHandler interface {
	UpdatePathValuesHandler(w http.ResponseWriter, r *http.Request)
	UpdateJSONHandler(w http.ResponseWriter, r *http.Request)
}

type UpdateHandlerImpl struct {
	storage repository.MemStorage
}

func BuildUpdateHandler(storage repository.MemStorage) UpdateHandler {
	return &UpdateHandlerImpl{storage: storage}
}

func (h *UpdateHandlerImpl) UpdatePathValuesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	metricType := r.PathValue("metricType")
	metricName := r.PathValue("metricName")
	metricValue := r.PathValue("metricValue")

	if metricType == models.Gauge {
		//log.Printf("Received update for gauge: %s - %s", metricName, metricValue)
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.storage.UpdateGauge(metricName, &val)
		w.WriteHeader(http.StatusOK)
	} else if metricType == models.Counter {
		//log.Printf("Received update for counter: %s - %s", metricName, metricValue)
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.storage.UpdateCounter(metricName, &val)
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}
}

func (h *UpdateHandlerImpl) UpdateJSONHandler(w http.ResponseWriter, r *http.Request) {
	var metric models.Metrics
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if metric.MType == models.Gauge {
		h.storage.UpdateGauge(metric.ID, metric.Value)
		w.WriteHeader(http.StatusOK)
	} else if metric.MType == models.Counter {
		h.storage.UpdateCounter(metric.ID, metric.Delta)
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}
}
