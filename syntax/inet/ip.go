package inet

import (
	"encoding/binary"
	"errors"
	"net"
	"strconv"
)

// AssignRandPort 修正版
func AssignRandPort(ip ...string) (int, error) {
	host := "0.0.0.0"
	if len(ip) > 0 && ip[0] != "" {
		host = ip[0]
	}

	// 1. 使用 JoinHostPort 确保 IPv6 兼容性 (例如 [::1]:0)
	addr := net.JoinHostPort(host, "0")

	// 2. 尝试监听
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return 0, err
	}

	// 3. 立即释放
	defer listener.Close()

	// 4. 类型断言确保安全获取端口
	if tcpAddr, ok := listener.Addr().(*net.TCPAddr); ok {
		return tcpAddr.Port, nil
	}

	return 0, errors.New("failed to get tcp port")
}

// ParseAddr 优化版：增强了对 IPv6 和私有地址的处理
func ParseAddr(addr string) (listenAddr, exposeAddr string, err error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		// 处理仅有端口或非法格式的情况
		if addr != "" && !contains(addr, ":") {
			host = ""
			port = addr
		} else {
			return "", "", err
		}
	}

	if port == "" || port == "0" {
		p, err := AssignRandPort(host)
		if err != nil {
			return "", "", err
		}
		port = strconv.Itoa(p)
	}

	// 2026 实践：增加对 IPv6 通配符的识别
	isWildcard := host == "" || host == "0.0.0.0" || host == "[::]" || host == "::"

	if !isWildcard {
		listenAddr = net.JoinHostPort(host, port)
		exposeAddr = listenAddr
	} else {
		ip, err := InternalIP()
		if err != nil {
			// 如果找不到内网IP，回退到本地环回
			ip = "127.0.0.1"
		}
		listenAddr = net.JoinHostPort(host, port)
		exposeAddr = net.JoinHostPort(ip, port)
	}
	return
}

// InternalIP 2026 推荐写法
func InternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		// 排除掉未开启或环回接口
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ipv4 := ip.To4()
			// 2026 实践：直接使用 IsPrivate 判断私有地址
			if ipv4 != nil && ipv4.IsPrivate() {
				return ipv4.String(), nil
			}
		}
	}
	return "", errors.New("no private ip address found")
}

// IP2Long 修正：确保 ParseIP 结果有效
func IP2Long(ipStr string) uint32 {
	parsedIP := net.ParseIP(ipStr)
	if parsedIP == nil {
		return 0
	}
	v := parsedIP.To4()
	if len(v) == 0 {
		return 0
	}
	return binary.BigEndian.Uint32(v)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s != "" // 简单辅助逻辑
}
