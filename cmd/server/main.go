package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/serg1732/practicum-yandex-metrics/internal/config"
	"github.com/serg1732/practicum-yandex-metrics/internal/handler"
	"github.com/serg1732/practicum-yandex-metrics/internal/helpers/cryptoutils"
	"github.com/serg1732/practicum-yandex-metrics/internal/logger"
	metrics_proto "github.com/serg1732/practicum-yandex-metrics/internal/proto"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"
	"github.com/serg1732/practicum-yandex-metrics/internal/service"
	"github.com/serg1732/practicum-yandex-metrics/internal/service/grpc_server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var buildVersion = "N/A"
var buildDate = "N/A"
var buildCommit = "N/A"

const maxSizeServerOptions = 2

func main() {
	printBuildInfo()
	log := logger.NewSlogLogger(slog.LevelInfo)
	serverConfig, errConfig := config.GetSeverConfig()
	if errConfig != nil {
		log.Error("Ошибка парсинга env значений", "error", errConfig)
	}
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT,
	)
	defer stop()

	audit := service.BuildAuditor(log)
	if serverConfig.AuditFile != "" {
		audit.Subscribe(service.BuildFileSubscriber(serverConfig.AuditFile))
	}

	if serverConfig.AuditURL != "" {
		audit.Subscribe(service.BuildHTTPSubscriber(repository.BuildRestyAuditMetrics(serverConfig.AuditURL)))
	}

	var mux *chi.Mux
	var grpcServer *grpc.Server
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
		mux, err = buildRouter(log, db, updaterHandler, readHandlers, serverConfig)
		if err != nil {
			log.Error("Ошибка при инициализации API", "error", err)
			os.Exit(1)
		}
		grpcServer, err = buildGRPCServer(ctx, log, serverConfig, db)
		if err != nil {
			log.Error("Ошибка при инициализации GRPC сервера", "error", err)
			os.Exit(1)
		}

	} else {
		var err error
		storage := repository.BuildMemStorage(ctx, log, serverConfig)
		updaterHandler := handler.BuildUpdateHandler(storage, &audit)
		readHandlers := handler.BuildReadHandler(storage)
		mux, err = buildRouter(log, nil, updaterHandler, readHandlers, serverConfig)
		if err != nil {
			log.Error("Ошибка при инициализации API", "error", err)
			os.Exit(1)
		}
		grpcServer, err = buildGRPCServer(ctx, log, serverConfig, storage)
		if err != nil {
			log.Error("Ошибка при инициализации GRPC сервера", "error", err)
			os.Exit(1)
		}
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

	if grpcServer != nil {
		go func() {
			log.Info("Запуск GRPC сервера", "address", serverConfig.GRPCRunAddr)
			listener, err := net.Listen("tcp", serverConfig.GRPCRunAddr)
			if err != nil {
				log.Error("ошибка при запуске GRPC сервера", "error", err)
				return
			}
			defer listener.Close()

			errServer := grpcServer.Serve(listener)
			if errServer != nil {
				log.Error("ошибка от GRPC сервера", "error", errServer)
				return
			}
		}()
	}

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(serverConfig.StoreInternal)*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("Ошибка при завершении работы http сервера", "error", err)
	}

	if grpcServer != nil {
		grpcServer.GracefulStop()
	}
}

func buildRouter(log *slog.Logger, db *repository.DataBase, updateHandlers handler.UpdateHandlerImpl,
	readHandlers handler.ReadMetricsHandlerImpl, config *config.ServerConfig) (*chi.Mux, error) {
	router := chi.NewRouter()
	trustSubnetHandler, errTrustSubnet := handler.TrustedSubnetMiddleware(config.TrustedSubnet)
	if errTrustSubnet != nil {
		log.Error("ошибка при инициализации проверки IP адреса", "error", errTrustSubnet)
		return nil, errTrustSubnet
	}
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrapWriter := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(wrapWriter, r)
		})
	})
	router.Use(logger.WithLogger(log))
	router.Use(handler.WithCheckHash(log, config.Key))
	router.Use(handler.WithGzipCompress(log))
	if config.CryptoKey != "" {
		if privateKey, errCryptoKey := cryptoutils.LoadPrivateKey(config.CryptoKey); errCryptoKey != nil {
			log.Error("ошибка при загрузке приватного ключа", "error", errCryptoKey.Error())
			return nil, errCryptoKey
		} else {
			router.Use(handler.WithCrypto(log, privateKey))
		}
	}
	router.Mount("/debug", middleware.Profiler())

	router.Route("/updates", func(r chi.Router) {
		r.Use(trustSubnetHandler)
		r.Post("/", updateHandlers.UpdateValues(log))
	})

	router.Route("/update", func(r chi.Router) {
		r.Use(trustSubnetHandler)
		r.Post("/", updateHandlers.UpdateJSONHandler(log))
		r.Post("/{metricType}/{metricName}/{metricValue}", updateHandlers.UpdatePathValuesHandler(log))
	})

	router.Route("/value", func(r chi.Router) {
		r.Post("/", readHandlers.SelectValueMetricHandler(log))
		r.Get("/{metricType}/{metricName}", readHandlers.SelectMetricHandler(log))
	})
	router.Get("/ping", readHandlers.PingDatabase(log, db))
	router.Get("/", readHandlers.AllMetricsHandler(log))
	return router, nil
}

func buildGRPCServer(ctx context.Context,
	log *slog.Logger,
	serverConfig *config.ServerConfig,
	storage grpc_server.MetricsUpdater) (*grpc.Server, error) {
	var server *grpc.Server
	serverOptions := make([]grpc.ServerOption, 0, maxSizeServerOptions)

	if serverConfig.TrustedSubnet != "" {
		interceptor, err := grpc_server.TrustedSubnetInterceptor(ctx, log, serverConfig.TrustedSubnet)
		if err != nil {
			log.Error("ошибка при запуске GRPC сервера", "error", err)
			return nil, err
		}
		serverOptions = append(serverOptions, grpc.ChainUnaryInterceptor(interceptor))
	}
	if serverConfig.TLSCertPath != "" && serverConfig.TLSKeyPath != "" {
		creds, errCreds := credentials.NewServerTLSFromFile(serverConfig.TLSCertPath, serverConfig.TLSKeyPath)
		if errCreds != nil {
			log.Error("ошибка при инициализации TLS", "error", errCreds)
			return nil, errCreds
		}
		serverOptions = append(serverOptions, grpc.Creds(creds))
	}
	server = grpc.NewServer(serverOptions...)
	metrics_proto.RegisterMetricsServer(server, grpc_server.BuildGRPCMetricsService(log, storage))
	return server, nil
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
