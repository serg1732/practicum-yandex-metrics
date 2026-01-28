package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"
)

type ReadMetricsHandler interface {
	AllMetricsHandler(w http.ResponseWriter, r *http.Request)
	SelectMetricHandler(w http.ResponseWriter, r *http.Request)
}

func BuildReadHandler(storage repository.MemStorage) ReadMetricsHandler {
	return &ReadMetricsHandlerImpl{storage: storage}
}

type ReadMetricsHandlerImpl struct {
	storage repository.MemStorage
}

func (h *ReadMetricsHandlerImpl) AllMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprintln(w, "<!doctype html><html><body>")
	fmt.Fprintln(w, "<h3>gauge</h3><pre>")
	for k, v := range h.storage.GetAllGauges() {
		fmt.Fprintf(w, "%s=%v\n", k, *v)
	}
	fmt.Fprintln(w, "</pre>")

	fmt.Fprintln(w, "<h3>counter</h3><pre>")
	for k, v := range h.storage.GetAllCounters() {
		fmt.Fprintf(w, "%s=%v\n", k, *v)
	}
	fmt.Fprintln(w, "</pre></body></html>")
	w.WriteHeader(http.StatusOK)
}

func (h *ReadMetricsHandlerImpl) SelectMetricHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	if metricType == models.Counter {
		val, isExist := h.storage.GetCounter(metricName)
		if !isExist {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		fmt.Fprintln(w, *val)
	} else if metricType == models.Gauge {
		val, isExist := h.storage.GetGauge(metricName)
		if !isExist {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		fmt.Fprintln(w, *val)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
	w.WriteHeader(http.StatusOK)
}
