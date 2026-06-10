package grpc_client

import (
	"context"
	"log/slog"

	"github.com/serg1732/practicum-yandex-metrics/internal/helpers/netutils"
	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
	metricproto "github.com/serg1732/practicum-yandex-metrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const realIPMetadataKey = "x-real-ip"

type Client struct {
	conn   *grpc.ClientConn
	client metricproto.MetricsClient
	logger *slog.Logger
}

func BuildGRPCMetricsClient(log *slog.Logger, address string, creds credentials.TransportCredentials) (*Client, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		logger: log,
		client: metricproto.NewMetricsClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) SendMetrics(ctx context.Context, metrics []models.Metrics) error {
	metricsProto := make([]*metricproto.Metric, 0, len(metrics))
	for _, m := range metrics {
		switch m.MType {
		case models.Gauge:
			if m.Value == nil {
				continue
			}

			metricsProto = append(metricsProto, metricproto.Metric_builder{
				Id:    m.ID,
				Type:  metricproto.Metric_GAUGE,
				Value: *m.Value,
			}.Build())

		case models.Counter:
			if m.Delta == nil {
				continue
			}

			metricsProto = append(metricsProto, metricproto.Metric_builder{
				Id:    m.ID,
				Type:  metricproto.Metric_COUNTER,
				Delta: *m.Delta,
			}.Build())
		}
	}

	req := metricproto.UpdateMetricsRequest_builder{
		Metrics: metricsProto,
	}.Build()

	if len(req.GetMetrics()) == 0 {
		return nil
	}

	ip, errGetIP := netutils.GetAgentIP()
	if errGetIP != nil {
		return errGetIP
	}

	ctx = metadata.AppendToOutgoingContext(ctx, realIPMetadataKey, ip)
	_, err := c.client.UpdateMetrics(ctx, req)
	return err
}
