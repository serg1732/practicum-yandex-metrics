package grpc_client

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
	metricproto "github.com/serg1732/practicum-yandex-metrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const realIPMetadataKey = "x-real-ip"

type Client struct {
	conn   *grpc.ClientConn
	client metricproto.MetricsClient
	logger *slog.Logger
}

func BuildGRPCMetricsClient(log *slog.Logger, address string) (*Client, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
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
	req := &metricproto.UpdateMetricsRequest{
		Metrics: make([]*metricproto.Metric, 0, len(metrics)),
	}

	for _, m := range metrics {
		switch m.MType {
		case models.Gauge:
			if m.Value == nil {
				continue
			}

			req.Metrics = append(req.Metrics, &metricproto.Metric{
				Id:    m.ID,
				Type:  metricproto.Metric_GAUGE,
				Value: *m.Value,
			})

		case models.Counter:
			if m.Delta == nil {
				continue
			}

			req.Metrics = append(req.Metrics, &metricproto.Metric{
				Id:    m.ID,
				Type:  metricproto.Metric_COUNTER,
				Delta: *m.Delta,
			})
		}
	}

	if len(req.Metrics) == 0 {
		return nil
	}

	ip, errGetIP := getAgentIP()
	if errGetIP != nil {
		return errGetIP
	}

	ctx = metadata.AppendToOutgoingContext(ctx, realIPMetadataKey, ip)
	_, err := c.client.UpdateMetrics(ctx, req)
	return err
}

func getAgentIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, errAddrs := iface.Addrs()
		if errAddrs != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP.To4()
			if ip == nil {
				continue
			}

			if ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}

			return ip.String(), nil
		}
	}

	return "", fmt.Errorf("адрес IPv4 не найден")
}
