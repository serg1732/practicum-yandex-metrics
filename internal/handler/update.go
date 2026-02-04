package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
)

type UpdateStorage interface {
	Update(log *slog.Logger, name string, Data *models.Metrics)
}

type UpdateHandlerImpl struct {
	storage UpdateStorage
}

func BuildUpdateHandler(storage UpdateStorage) UpdateHandlerImpl {
	return UpdateHandlerImpl{storage: storage}
}

func (h *UpdateHandlerImpl) UpdatePathValuesHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Error("Некорректный метод в запросе", "method", r.Method)
			http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
			return
		}
		metricType := r.PathValue("metricType")
		metricName := r.PathValue("metricName")
		metricValue := r.PathValue("metricValue")

		if metricType == models.Gauge {
			val, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				log.Error("Ошибка при конвертации значения метрики", "error", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			h.storage.Update(log, metricName, &models.Metrics{ID: metricName, MType: models.Gauge, Value: &val})
			w.WriteHeader(http.StatusOK)
		} else if metricType == models.Counter {
			val, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				log.Error("Ошибка при конвертации значения метрики", "error", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			h.storage.Update(log, metricName, &models.Metrics{ID: metricName, MType: models.Counter, Delta: &val})
			w.WriteHeader(http.StatusOK)
		} else {
			log.Error("Неизвестный тип метрики из запроса", "type", metricType)
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}
	}
}

func (h *UpdateHandlerImpl) UpdateJSONHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric models.Metrics
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&metric); err != nil {
			log.Error("Ошибка при конвертации тела запрос в JSON")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if metric.MType == models.Gauge {
			h.storage.Update(log, metric.ID, &metric)
			log.Debug("Успешное обновление метрики", "name", metric.ID, "type", metric.MType)
			w.WriteHeader(http.StatusOK)
		} else if metric.MType == models.Counter {
			h.storage.Update(log, metric.ID, &metric)
			log.Debug("Успешное обновление метрики", "name", metric.ID, "type", metric.MType)
			w.WriteHeader(http.StatusOK)
		} else {
			log.Error("Неизвестный тип метрики из запроса", "type", metric.MType)
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}
	}
}
