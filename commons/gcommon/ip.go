package gcommon

import (
	"errors"
	"net"
	"net/http"
	"strings"
)

// Define http headers.
const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
	XClientIP     = "x-client-ip"
)

// GetLocalIP 获取本地IP地址
func GetLocalIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() {
				continue
			}

			ip4 := ipNet.IP.To4()
			if ip4 != nil && !ip4.IsLoopback() {
				if !ip4.IsMulticast() {
					return ip4.String(), nil
				}
			}
		}
	}
	// If no suitable IP is found, return an error
	return "", errors.New("no suitable local IP address found")
}

// RemoteIP 返回客户端的IP地址
func RemoteIP(req *http.Request) string {
	if ip := req.Header.Get(XClientIP); ip != "" {
		return ip
	} else if ip = req.Header.Get(XRealIP); ip != "" {
		return ip
	} else if ip = req.Header.Get(XForwardedFor); ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}

	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		ip = req.RemoteAddr
	}

	if ip == "::1" {
		return "127.0.0.1"
	}
	return ip
}
