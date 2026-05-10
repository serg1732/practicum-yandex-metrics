//go:generate mockgen -source=read.go -destination=mocks/mock_read.go -package=mocks
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"text/template"

	"github.com/go-chi/chi/v5"
	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"
)

// ReadStorage представляет интерфейс, отражающий функционал по работе с хранилищем.
type ReadStorage interface {
	GetCounter(ctx context.Context, name string) (*models.Metrics, error)
	GetGauge(ctx context.Context, name string) (*models.Metrics, error)
	GetAllCounters(ctx context.Context) (map[string]*models.Metrics, error)
	GetAllGauges(ctx context.Context) (map[string]*models.Metrics, error)
}

// BuildReadHandler создание обработчика получения метрик.
func BuildReadHandler(storage ReadStorage) ReadMetricsHandlerImpl {
	return ReadMetricsHandlerImpl{storage: storage}
}

// ReadMetricsHandlerImpl обработчик получения метрик.
type ReadMetricsHandlerImpl struct {
	storage      ReadStorage
	templateHTML *template.Template
}

// AllMetricsHandler обработчик получения всех метрик из хранилища.
func (h *ReadMetricsHandlerImpl) AllMetricsHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		templateHTML, errParseTemplate := template.New("All").Parse(getTemplate())
		if errParseTemplate != nil {
			log.Error("Ошибка при парсинге шаблона html", "error", errParseTemplate)
			http.Error(w, errParseTemplate.Error(), http.StatusInternalServerError)
			return
		}

		gauges, errGauges := h.storage.GetAllGauges(r.Context())
		if errGauges != nil {
			log.Error("Ошибка при получении всех метрик gauges", "error", errGauges)
		}

		counter, errCounter := h.storage.GetAllCounters(r.Context())
		if errCounter != nil {
			log.Error("Ошибка при получении всех метрик counter", "error", errGauges)
		}

		data := struct {
			GaugeMap   map[string]*models.Metrics
			CounterMap map[string]*models.Metrics
		}{
			GaugeMap:   gauges,
			CounterMap: counter,
		}
		err := templateHTML.Execute(w, data)
		if err != nil {
			log.Error("Ошибка при заполнении шаблона html", "error", err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// SelectMetricHandler обработчик получения метрики по имени и типу.
func (h *ReadMetricsHandlerImpl) SelectMetricHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		if metricType == models.Counter {
			val, errCounter := h.storage.GetCounter(r.Context(), metricName)
			if errCounter != nil {
				if errors.Is(errCounter, repository.ErrorMetricNotFound) {
					log.Debug("Метрика не найдена", "name", metricName, "type", metricType)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				log.Error("Ошибка при получении метрики", "metric_type", metricType, "metric_name", metricName)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			fmt.Fprintln(w, *val.Delta)
		} else if metricType == models.Gauge {
			val, errGauge := h.storage.GetGauge(r.Context(), metricName)
			if errGauge != nil {
				if errors.Is(errGauge, repository.ErrorMetricNotFound) {
					log.Debug("Метрика не найдена", "name", metricName, "type", metricType)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				log.Error("Ошибка при получении метрики", "metric_type", metricType, "metric_name", metricName)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			fmt.Fprintln(w, *val.Value)
		} else {
			log.Debug("Неизвестный тип метрики", "name", metricName, "type", metricType)
			w.WriteHeader(http.StatusNotFound)
		}
		log.Debug("Метрика найдена", "name", metricName, "type", metricType)
		w.WriteHeader(http.StatusOK)
	}
}

// SelectValueMetricHandler обработчик получения метрики.
// Данные имени метрики и типа указываются в теле запроса в формате JSON.
func (h *ReadMetricsHandlerImpl) SelectValueMetricHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		var metric models.Metrics
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&metric); err != nil {
			log.Error("Ошибка конвертации JSON из тела запроса", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Debug("Поиск метрики", "name", metric.ID, "type", metric.MType)

		if metric.MType == models.Counter {
			val, errCounter := h.storage.GetCounter(r.Context(), metric.ID)
			if errCounter != nil {
				if errors.Is(errCounter, repository.ErrorMetricNotFound) {
					log.Debug("Метрика не найдена", "name", metric.ID, "type", metric.MType)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				log.Error("Ошибка при получении метрики", "name", metric.ID, "type", metric.MType, "error", errCounter)
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else if val == nil {
				log.Debug("Метрика не найдена", "name", metric.ID, "type", metric.MType)
				w.WriteHeader(http.StatusNotFound)
				return
			}
			metric.Delta = val.Delta
			enc := json.NewEncoder(w)
			if err := enc.Encode(metric); err != nil {
				log.Error("Ошибка при конвертации в JSON данных для отправки", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			log.Debug("Найдена метрика", "name", metric.ID, "type", metric.MType, "delta", metric.Delta)
		} else if metric.MType == models.Gauge {
			val, errGauge := h.storage.GetGauge(r.Context(), metric.ID)
			if errGauge != nil {
				if errors.Is(errGauge, repository.ErrorMetricNotFound) {
					log.Debug("Метрика не найдена", "name", metric.ID, "type", metric.MType)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				log.Error("Ошибка при получении метрики", "name", metric.ID, "type", metric.MType, "error", errGauge)
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else if val == nil {
				log.Debug("Метрика не найдена", "name", metric.ID, "type", metric.MType)
				w.WriteHeader(http.StatusNotFound)
				return
			}
			metric.Value = val.Value
			enc := json.NewEncoder(w)
			if err := enc.Encode(metric); err != nil {
				log.Error("Ошибка при конвертации в JSON данных для отправки", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			log.Debug("Найдена метрика", "name", metric.ID, "type", metric.MType, "value", metric.Value)
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// PingDatabase обработчик запроса на проверку работоспособности БД.
func (h *ReadMetricsHandlerImpl) PingDatabase(log *slog.Logger, db *repository.DataBase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db != nil && db.Ping(r.Context()) != nil {
			log.Error("Ошибка при проверки подключения к БД (строка пустая или нет подключения)")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// getTemplate шаблон для вывода всех метрик.
func getTemplate() string {
	return `
	<!doctype html>
	<html>
		<head>
    	<meta charset="utf-8">
    	<title>Metrics</title>
		</head>
		<body>
			<h3>gauge</h3>
				<pre>
					{{range $k, $v := .GaugeMap}}
					{{$k}}={{$v.Value}}
					{{end}}
				</pre>

			<h3>counter</h3>
		<pre>
			{{range $k, $v := .CounterMap}}
			{{$k}}={{$v.Delta}}
			{{end}}
		</pre>
		</body>
	</html>
`
}
