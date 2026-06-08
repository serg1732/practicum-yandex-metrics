package grpc_server

import (
	"context"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const realIPMetadataKey = "x-real-ip"

func TrustedSubnetInterceptor(ctx context.Context, log *slog.Logger, trustedSubnet string) (grpc.UnaryServerInterceptor, error) {
	subnet, err := parseTrustedSubnet(trustedSubnet)
	if err != nil {
		return nil, err
	}

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if subnet == nil {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Error("отсутствует metadata в grpc запросе")
			return nil, status.Error(codes.PermissionDenied, "отсутствует metadata")
		}

		values := md.Get(realIPMetadataKey)
		if len(values) == 0 {
			log.Error("отсутствует x-real-ip в grpc запросе")
			return nil, status.Error(codes.PermissionDenied, "отсутствует x-real-ip")
		}

		ip := values[0]
		if net.ParseIP(ip) == nil {
			log.Error("некорретный x-real-ip в grpc запросе")
			return nil, status.Error(codes.PermissionDenied, "некорретный x-real-ip")
		}

		if !checkIPInSubnet(ip, subnet) {
			log.Error("IP адрес не в списке разрешенных")
			return nil, status.Error(codes.PermissionDenied, "IP адрес не в списке разрешенных")
		}

		return handler(ctx, req)
	}, nil
}

func parseTrustedSubnet(value string) (*net.IPNet, error) {
	if value == "" {
		return nil, nil
	}

	_, subnet, err := net.ParseCIDR(value)
	if err != nil {
		return nil, err
	}

	return subnet, nil
}

func checkIPInSubnet(ipStr string, subnet *net.IPNet) bool {
	if subnet == nil {
		return true
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	return subnet.Contains(ip)
}
