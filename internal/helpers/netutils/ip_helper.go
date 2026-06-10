package netutils

import (
	"fmt"
	"net"
	"strings"
)

func GetAgentIPWithServerIP(serverAddress string) (string, error) {
	if strings.Contains(serverAddress, "localhost") ||
		strings.Contains(serverAddress, "127.0.0.1") {
		return "127.0.0.1", nil
	}
	return GetAgentIP()
}

func GetAgentIP() (string, error) {
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
