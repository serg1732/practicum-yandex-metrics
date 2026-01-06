package main

import (
	"log"
	"net/http"

	"github.com/serg1732/practicum-yandex-metrics/internal/handler"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"
)

func main() {

	updaterHandler := handler.BuildUpdateHandler(repository.BuildMemStorage())
	mux := buildHTTPMux(updaterHandler)
	log.Printf("Starting server on port 8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}

func buildHTTPMux(handlers handler.UpdateHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /update/{metricType}/{metricName}/{metricValue}", handlers.UpdateHandler)
	mux.HandleFunc("POST /update/{metricType}/{metricValue}", http.NotFound)
	return mux
}
