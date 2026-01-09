package main

import (
	"log"
	"net/http"

	"github.com/serg1732/practicum-yandex-metrics/internal/handler"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"

	"github.com/go-chi/chi/v5"
)

func main() {
	storage := repository.BuildMemStorage()
	updaterHandler := handler.BuildUpdateHandler(storage)
	readHandlers := handler.BuildReadHandler(storage)
	mux := buildRouter(updaterHandler, readHandlers)
	log.Printf("Starting server on port 8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}

func buildRouter(updateHandlers handler.UpdateHandler, readHandlers handler.ReadMetricsHandler) *chi.Mux {
	router := chi.NewRouter()
	router.Post("/update/{metricType}/{metricName}/{metricValue}", updateHandlers.UpdateHandler)
	router.Get("/", readHandlers.AllMetricsHandler)
	router.Get("/value/{metricType}/{metricName}", readHandlers.SelectMetricHandler)
	return router
}
