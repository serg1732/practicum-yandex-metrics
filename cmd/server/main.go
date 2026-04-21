package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/serg1732/practicum-yandex-metrics/internal/config"
	"github.com/serg1732/practicum-yandex-metrics/internal/handler"
	"github.com/serg1732/practicum-yandex-metrics/internal/logger"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"
	"github.com/serg1732/practicum-yandex-metrics/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	log := logger.NewSlogLogger(slog.LevelInfo)
	serverConfig, errConfig := config.GetSeverConfig()
	if errConfig != nil {
		log.Error("Ошибка парсинга env значений", "error", errConfig)
	}
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	audit := service.BuildAuditor(log)
	if serverConfig.AuditFile != "" {
		audit.Subscribe(service.BuildFileSubscriber(serverConfig.AuditFile))
	}

	if serverConfig.AuditUrl != "" {
		audit.Subscribe(service.BuildHttpSubscriber(repository.BuildRestyAuditMetrics(serverConfig.AuditUrl)))
	}

	var mux *chi.Mux
	if serverConfig.DSN != "" {
		db, err := repository.BuildDataBase(ctx, log, serverConfig)
		if err != nil {
			log.Error("Ошибка подключения к БД", "error", err)
			os.Exit(1)
		}

		errMigrate := repository.MigrateDataBase(log, serverConfig)
		if errMigrate != nil {
			log.Error("Ошибка при миграции", "error", errMigrate)
			os.Exit(1)
		}
		updaterHandler := handler.BuildUpdateHandler(db, &audit)
		readHandlers := handler.BuildReadHandler(db)
		mux = buildRouter(log, db, updaterHandler, readHandlers, serverConfig.Key)

	} else {
		storage := repository.BuildMemStorage(ctx, log, serverConfig)
		updaterHandler := handler.BuildUpdateHandler(storage, &audit)
		readHandlers := handler.BuildReadHandler(storage)
		mux = buildRouter(log, nil, updaterHandler, readHandlers, serverConfig.Key)
	}

	log.Info("Запуск http сервера", "address", serverConfig.RunAddr)
	srv := &http.Server{
		Addr:    serverConfig.RunAddr,
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Ошибка в http сервере", "error", err)
			return
		}
		log.Info("Завершение работы http сервера")
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(serverConfig.StoreInternal)*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("Ошибка при завершении работы http сервера", "error", err)
	}
}

func buildRouter(log *slog.Logger, db *repository.DataBase, updateHandlers handler.UpdateHandlerImpl,
	readHandlers handler.ReadMetricsHandlerImpl, key string) *chi.Mux {
	router := chi.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrapWriter := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(wrapWriter, r)
		})
	})
	router.Use(logger.WithLogger(log))
	router.Use(handler.WithCheckHash(log, key))
	router.Use(handler.WithGzipCompress(log))
	router.Mount("/debug", middleware.Profiler())

	router.Route("/updates", func(r chi.Router) {
		r.Post("/", updateHandlers.UpdateValues(log))
	})

	router.Route("/update", func(r chi.Router) {
		r.Post("/", updateHandlers.UpdateJSONHandler(log))
		r.Post("/{metricType}/{metricName}/{metricValue}", updateHandlers.UpdatePathValuesHandler(log))
	})

	router.Route("/value", func(r chi.Router) {
		r.Post("/", readHandlers.SelectValueMetricHandler(log))
		r.Get("/{metricType}/{metricName}", readHandlers.SelectMetricHandler(log))
	})
	router.Get("/ping", readHandlers.PingDatabase(log, db))
	router.Get("/", readHandlers.AllMetricsHandler(log))
	return router
}
