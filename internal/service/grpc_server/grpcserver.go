package grpc_server

import (
	"context"
	"log/slog"

	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
	metrics_proto "github.com/serg1732/practicum-yandex-metrics/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricsUpdater interface {
	Updates(ctx context.Context, log *slog.Logger, Data []*models.Metrics) error
}

type GRPCMetricsService struct {
	metrics_proto.UnimplementedMetricsServer
	storage MetricsUpdater
	log     *slog.Logger
}

func BuildGRPCMetricsService(log *slog.Logger, storage MetricsUpdater) *GRPCMetricsService {
	return &GRPCMetricsService{
		storage: storage,
		log:     log,
	}
}

func (s *GRPCMetricsService) UpdateMetrics(
	ctx context.Context,
	req *metrics_proto.UpdateMetricsRequest) (*metrics_proto.UpdateMetricsResponse, error) {
	metrics := make([]*models.Metrics, 0, len(req.GetMetrics()))
	s.log.Debug("Обновление метрик по GRPC", "metrics", req.GetMetrics())

	for _, m := range req.GetMetrics() {
		if m.GetId() == "" {
			return nil, status.Error(codes.InvalidArgument, "идентификатор метрики отсутствует")
		}

		switch m.GetType() {
		case metrics_proto.Metric_GAUGE:
			value := m.GetValue()

			metrics = append(metrics, &models.Metrics{
				ID:    m.GetId(),
				MType: models.Gauge,
				Value: &value,
			})

		case metrics_proto.Metric_COUNTER:
			delta := m.GetDelta()

			metrics = append(metrics, &models.Metrics{
				ID:    m.GetId(),
				MType: models.Counter,
				Delta: &delta,
			})

		default:
			return nil, status.Error(codes.InvalidArgument, "неизвестный тип метрики")
		}
	}

	if len(metrics) == 0 {
		return &metrics_proto.UpdateMetricsResponse{}, nil
	}

	if err := s.storage.Updates(ctx, s.log, metrics); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &metrics_proto.UpdateMetricsResponse{}, nil
}
