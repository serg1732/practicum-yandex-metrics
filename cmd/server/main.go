package main

import (
	"net/http"

	"github.com/serg1732/practicum-yandex-metrics/internal/handler"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /update/{metricType}/{metricName}/{metricValue}", handler.UpdateHandler)
	mux.HandleFunc("POST /update/{metricType}/{metricValue}", http.NotFound)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
