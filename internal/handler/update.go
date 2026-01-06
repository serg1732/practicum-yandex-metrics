package handler

import (
	"net/http"
	"strconv"

	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"
)

type UpdateHandler interface {
	UpdateHandler(w http.ResponseWriter, r *http.Request)
}

type UpdateHandlerImpl struct {
	storage repository.MemStorage
}

func BuildUpdateHandler(storage repository.MemStorage) UpdateHandler {
	return &UpdateHandlerImpl{storage: storage}
}

func (h *UpdateHandlerImpl) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	metricType := r.PathValue("metricType")
	metricName := r.PathValue("metricName")
	metricValue := r.PathValue("metricValue")

	if metricType == "gauge" {
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.storage.UpdateGauge(metricName, models.Gauge(val))
	} else if metricType == "counter" {
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.storage.UpdateCounter(metricName, models.Counter(val))
	} else {
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}
	http.Error(w, "OK", http.StatusOK)
}
