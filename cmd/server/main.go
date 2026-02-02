package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/serg1732/practicum-yandex-metrics/internal/config"
	"github.com/serg1732/practicum-yandex-metrics/internal/handler"
	"github.com/serg1732/practicum-yandex-metrics/internal/logger"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	var serverConfig config.ServerConfig
	parseFlags(&serverConfig)
	parseEnvs(&serverConfig)
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()
	storage := repository.BuildMemStorage(&serverConfig, ctx)
	updaterHandler := handler.BuildUpdateHandler(storage)
	readHandlers := handler.BuildReadHandler(storage)
	mux := buildRouter(updaterHandler, readHandlers)
	log.Printf("Starting server on address %s", serverConfig.RunAddr)
	srv := &http.Server{
		Addr:    serverConfig.RunAddr,
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(serverConfig.StoreInternal)*time.Second)
	defer cancel()

	_ = srv.Shutdown(shutdownCtx)
}

func buildRouter(updateHandlers handler.UpdateHandler, readHandlers handler.ReadMetricsHandler) *chi.Mux {
	router := chi.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrapWriter := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(wrapWriter, r)
		})
	})
	router.Use(logger.WithLogger())
	router.Use(handler.WithGzipCompress())
	router.Route("/update", func(r chi.Router) {
		r.Post("/", updateHandlers.UpdateJSONHandler)
		r.Post("/{metricType}/{metricName}/{metricValue}", updateHandlers.UpdatePathValuesHandler)
	})

	router.Route("/value", func(r chi.Router) {
		r.Post("/", readHandlers.SelectValueMetricHandler)
		r.Get("/{metricType}/{metricName}", readHandlers.SelectMetricHandler)
	})

	router.Get("/", readHandlers.AllMetricsHandler)
	return router
}
