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

// IsIPv4 功能与 net.ParseIP 类似，
// 但不检查 IPv6 的情况，并且不返回 net.IP 切片，因此 IsIPv4 无需分配内存。
func IsIPv4(s string) bool {
	for i := 0; i < net.IPv4len; i++ {
		// 检查字符串是否为空
		if len(s) == 0 {
			return false
		}

		// 如果不是第一个字节组，则检查并跳过点分隔符
		if i > 0 {
			if s[0] != '.' {
				return false
			}
			s = s[1:]
		}

		n, ci := 0, 0

		// 解析十进制数，构建当前字节的值
		for ci = 0; ci < len(s) && '0' <= s[ci] && s[ci] <= '9'; ci++ {
			n = n*10 + int(s[ci]-'0')
			// 如果超出 0xFF（255）范围，返回 false
			if n > 0xFF {
				return false
			}
		}

		// 检查解析是否合法（没有前导零等问题）
		if ci == 0 || (ci > 1 && s[0] == '0') {
			return false
		}

		// 移动到下一个部分
		s = s[ci:]
	}

	// 确保解析完成后字符串为空
	return len(s) == 0
}

// IsIPv6 功能与 net.ParseIP 类似，
// 但不检查 IPv4 的情况，并且不返回 net.IP 切片，因此 IsIPv6 无需分配内存。
func IsIPv6(s string) bool {
	ellipsis := -1 // 记录“省略符号”（双冒号）的位置

	// 检查是否有开头的省略符号
	if len(s) >= 2 && s[0] == ':' && s[1] == ':' {
		ellipsis = 0
		s = s[2:]
		// 如果只有省略号，则是合法的 IPv6 地址
		if len(s) == 0 {
			return true
		}
	}

	// 循环解析每个十六进制部分
	i := 0
	for i < net.IPv6len {
		n, ci := 0, 0

		// 解析十六进制数
		for ci = 0; ci < len(s); ci++ {
			if '0' <= s[ci] && s[ci] <= '9' {
				n = n*16 + int(s[ci]-'0')
			} else if 'a' <= s[ci] && s[ci] <= 'f' {
				n = n*16 + int(s[ci]-'a') + 10
			} else if 'A' <= s[ci] && s[ci] <= 'F' {
				n = n*16 + int(s[ci]-'A') + 10
			} else {
				break
			}
			if n > 0xFFFF {
				return false
			}
		}

		// 检查解析的合法性
		if ci == 0 || n > 0xFFFF {
			return false
		}

		// 检查 IPv4 尾部
		if ci < len(s) && s[ci] == '.' {
			if ellipsis < 0 && i != net.IPv6len-net.IPv4len {
				return false
			}
			if i+net.IPv4len > net.IPv6len {
				return false
			}

			// 调用 IsIPv4 判断尾部是否是有效的 IPv4 地址
			if !IsIPv4(s) {
				return false
			}

			s = ""
			i += net.IPv4len
			break
		}

		// 保存当前 16 位块
		i += 2

		// 检查是否到达字符串末尾
		s = s[ci:]
		if len(s) == 0 {
			break
		}

		// 检查下一个字符是否为冒号
		if s[0] != ':' || len(s) == 1 {
			return false
		}
		s = s[1:]

		// 检查省略号（双冒号）
		if s[0] == ':' {
			if ellipsis >= 0 { // 如果已存在一个省略号
				return false
			}
			ellipsis = i
			s = s[1:]
			if len(s) == 0 { // 省略号可以在末尾
				break
			}
		}
	}

	// 检查是否使用了完整的字符串
	if len(s) != 0 {
		return false
	}

	// 如果省略了部分，需要展开省略号
	if i < net.IPv6len {
		if ellipsis < 0 {
			return false
		}
	} else if ellipsis >= 0 {
		// 省略号必须至少表示一个 0 组
		return false
	}
	return true
}
