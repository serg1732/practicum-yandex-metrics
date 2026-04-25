package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
)

// UpdateStorage представляет интерфейс, отражающий реализацию хранилища.
type UpdateStorage interface {
	Update(ctx context.Context, log *slog.Logger, Data *models.Metrics) error
	Updates(ctx context.Context, log *slog.Logger, Data []*models.Metrics) error
}

// Auditor представляет интерфейс, отражающий реализацию аудитора обработки метрик.
type Auditor interface {
	// BroadCast уведомление подписчиков об успешной обработке метрики.
	BroadCast(data *models.AuditEvent)
}

// UpdateHandlerImpl обработчик запросов на запись в хранилище.
type UpdateHandlerImpl struct {
	// auditor аудитор запросов.
	auditor Auditor
	// storage хранилище метрик.
	storage UpdateStorage
}

// BuildUpdateHandler функция создания обработчика запросов на запись в хранилище.
func BuildUpdateHandler(storage UpdateStorage, auditor Auditor) UpdateHandlerImpl {
	return UpdateHandlerImpl{storage: storage, auditor: auditor}
}

// UpdatePathValuesHandler обработчик запрос с обовлением метрики из Path значений.
func (h *UpdateHandlerImpl) UpdatePathValuesHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Debug("Некорректный метод в запросе", "method", r.Method)
			http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
			return
		}
		metricType := r.PathValue("metricType")
		metricName := r.PathValue("metricName")
		metricValue := r.PathValue("metricValue")

		if metricType == models.Gauge {
			val, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				log.Debug("Ошибка при конвертации значения метрики", "error", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			errUpdate := h.storage.Update(r.Context(), log, &models.Metrics{ID: metricName, MType: models.Gauge, Value: &val})
			if errUpdate != nil {
				log.Error("Ошибка при обновлении метрики", "error", errUpdate)
			}
			w.WriteHeader(http.StatusOK)
		} else if metricType == models.Counter {
			val, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				log.Debug("Ошибка при конвертации значения метрики", "error", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			errUpdate := h.storage.Update(r.Context(), log, &models.Metrics{ID: metricName, MType: models.Counter, Delta: &val})
			if errUpdate != nil {
				log.Error("Ошибка при обновлении метрики", "error", errUpdate)
			}
			w.WriteHeader(http.StatusOK)
		} else {
			log.Debug("Неизвестный тип метрики из запроса", "type", metricType)
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}
		h.auditor.BroadCast(&models.AuditEvent{
			TS:        time.Now().Unix(),
			Metrics:   []string{metricName},
			IPAddress: strings.Split(r.RemoteAddr, ":")[0],
		})
	}
}

// UpdateJSONHandler обработчик обновления значения метрики.
func (h *UpdateHandlerImpl) UpdateJSONHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric models.Metrics
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&metric); err != nil {
			log.Error("Ошибка при конвертации тела запрос в JSON")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if metric.MType == models.Gauge {
			err := h.storage.Update(r.Context(), log, &metric)
			if err != nil {
				log.Error("Ошибка при обновлении метрики", "error", err)
			}
			log.Debug("Успешное обновление метрики", "name", metric.ID, "type", metric.MType)
			w.WriteHeader(http.StatusOK)
		} else if metric.MType == models.Counter {
			err := h.storage.Update(r.Context(), log, &metric)
			if err != nil {
				log.Error("Ошибка при обновлении метрики", "error", err)
			}
			log.Debug("Успешное обновление метрики", "name", metric.ID, "type", metric.MType)
			w.WriteHeader(http.StatusOK)
		} else {
			log.Error("Неизвестный тип метрики из запроса", "type", metric.MType)
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}
		h.auditor.BroadCast(&models.AuditEvent{
			TS:        time.Now().Unix(),
			Metrics:   []string{metric.ID},
			IPAddress: strings.Split(r.RemoteAddr, ":")[0],
		})
	}
}

// UpdateValues обработчик запроса на запись / обновление набора метрик.
func (h *UpdateHandlerImpl) UpdateValues(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []*models.Metrics
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&metrics); err != nil {
			log.Error("Ошибка при конвертации тела запрос в JSON")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := h.storage.Updates(r.Context(), log, metrics); err != nil {
			log.Error("Ошибка при обновлении метрик", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Debug("Обновлены метрики", "metrics", metrics)
		w.WriteHeader(http.StatusOK)
		names := make([]string, 0, len(metrics))
		for _, metric := range metrics {
			names = append(names, metric.ID)
		}
		h.auditor.BroadCast(&models.AuditEvent{
			TS:        time.Now().Unix(),
			Metrics:   names,
			IPAddress: strings.Split(r.RemoteAddr, ":")[0],
		})
	}
}
