//go:generate mockgen -source=read.go -destination=mocks/mock_read.go -package=mocks
package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"text/template"

	"github.com/go-chi/chi/v5"
	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"
)

type ReadStorage interface {
	GetCounter(name string) (*models.Metrics, bool)
	GetGauge(name string) (*models.Metrics, bool)
	GetAllCounters() map[string]*models.Metrics
	GetAllGauges() map[string]*models.Metrics
}

func BuildReadHandler(storage ReadStorage) ReadMetricsHandlerImpl {
	return ReadMetricsHandlerImpl{storage: storage}
}

type ReadMetricsHandlerImpl struct {
	storage      ReadStorage
	templateHTML *template.Template
}

func (h *ReadMetricsHandlerImpl) AllMetricsHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		templateHTML, errParseTemplate := template.New("All").Parse(getTemplate())
		if errParseTemplate != nil {
			log.Error("Ошибка при парсинге шаблона html", "error", errParseTemplate)
			http.Error(w, errParseTemplate.Error(), http.StatusInternalServerError)
			return
		}

		data := struct {
			GaugeMap   map[string]*models.Metrics
			CounterMap map[string]*models.Metrics
		}{
			GaugeMap:   h.storage.GetAllGauges(),
			CounterMap: h.storage.GetAllCounters(),
		}
		err := templateHTML.Execute(w, data)
		if err != nil {
			log.Error("Ошибка при заполнении шаблона html", "error", err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (h *ReadMetricsHandlerImpl) SelectMetricHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		if metricType == models.Counter {
			val, isExist := h.storage.GetCounter(metricName)
			if !isExist {
				log.Error("Метрика не найдена", "metric_type", metricType, "metric_name", metricName)
				w.WriteHeader(http.StatusNotFound)
				return
			}
			fmt.Fprintln(w, *val.Delta)
		} else if metricType == models.Gauge {
			val, isExist := h.storage.GetGauge(metricName)
			if !isExist {
				log.Error("Метрика не найдена", "name", metricName, "type", metricType)
				w.WriteHeader(http.StatusNotFound)
				return
			}
			fmt.Fprintln(w, *val.Value)
		} else {
			log.Error("Неизвестный тип метрики", "name", metricName, "type", metricType)
			w.WriteHeader(http.StatusNotFound)
		}
		log.Info("Метрика найдена", "name", metricName, "type", metricType)
		w.WriteHeader(http.StatusOK)
	}
}

func (h *ReadMetricsHandlerImpl) SelectValueMetricHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		var metric models.Metrics
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&metric); err != nil {
			log.Error("Ошибка конвертации JSON из тела запроса", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if metric.MType == models.Counter {
			val, isExist := h.storage.GetCounter(metric.ID)
			if !isExist {
				log.Error("Метрика не найдена", "name", metric.ID, "type", metric.MType)
				w.WriteHeader(http.StatusNotFound)
				return
			}
			metric.Delta = val.Delta
			enc := json.NewEncoder(w)
			if err := enc.Encode(metric); err != nil {
				log.Error("Ошибка при конвертации в JSON данных для отправки", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else if metric.MType == models.Gauge {
			val, isExist := h.storage.GetGauge(metric.ID)
			if !isExist {
				log.Error("Метрика не найдена", "name", metric.ID, "type", metric.MType)
				w.WriteHeader(http.StatusNotFound)
				return
			}
			metric.Value = val.Value
			enc := json.NewEncoder(w)
			if err := enc.Encode(metric); err != nil {
				log.Error("Ошибка при конвертации в JSON данных для отправки", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func (h *ReadMetricsHandlerImpl) PingDatabase(log *slog.Logger, db *repository.DataBase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db.Ping() != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
}

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
