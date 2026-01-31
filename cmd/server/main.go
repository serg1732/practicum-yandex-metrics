package main

import (
	"log"
	"net/http"

	"github.com/serg1732/practicum-yandex-metrics/internal/config"
	"github.com/serg1732/practicum-yandex-metrics/internal/handler"
	"github.com/serg1732/practicum-yandex-metrics/internal/logger"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"

	"github.com/go-chi/chi/v5"
)

func main() {
	storage := repository.BuildMemStorage()
	updaterHandler := handler.BuildUpdateHandler(storage)
	readHandlers := handler.BuildReadHandler(storage)
	mux := buildRouter(updaterHandler, readHandlers)
	var serverConfig config.ServerConfig
	parseFlags(&serverConfig)
	parseEnvs(&serverConfig)
	log.Printf("Starting server on address %s", serverConfig.RunAddr)
	err := http.ListenAndServe(serverConfig.RunAddr, mux)
	if err != nil {
		panic(err)
	}
}

func buildRouter(updateHandlers handler.UpdateHandler, readHandlers handler.ReadMetricsHandler) *chi.Mux {
	router := chi.NewRouter()
	router.Use(logger.WithLogger())
	router.Post("/update/{metricType}/{metricName}/{metricValue}", updateHandlers.UpdatePathValuesHandler)
	router.Post("/update/", updateHandlers.UpdateJSONHandler)
	router.Post("/value/", readHandlers.SelectValueMetricHandler)
	router.Get("/", readHandlers.AllMetricsHandler)
	router.Get("/value/{metricType}/{metricName}", readHandlers.SelectMetricHandler)
	return router
}
