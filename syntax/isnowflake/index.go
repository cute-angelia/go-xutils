package isnowflake

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/sony/sonyflake"
)

var (
	sf   *sonyflake.Sonyflake
	once sync.Once
)

type SnowRest struct {
	Id uint64
}

func (f SnowRest) Int64() int64   { return int64(f.Id) }
func (f SnowRest) String() string { return strconv.FormatUint(f.Id, 10) }

func GetSnowflake() *sonyflake.Sonyflake {
	once.Do(func() {
		settings := sonyflake.Settings{
			MachineID: func() (uint16, error) {
				// 1. 环境变量优先
				if envID := os.Getenv("MACHINE_ID"); envID != "" {
					id, err := strconv.ParseUint(envID, 10, 16)
					if err != nil {
						return 0, fmt.Errorf("invalid MACHINE_ID env: %w", err)
					}
					return uint16(id), nil
				}

				// 2. 自动获取 IP
				return getMachineIDByIP()
			},
		}

		sf = sonyflake.NewSonyflake(settings)
		if sf == nil {
			panic("sonyflake 启动失败：请检查系统时钟")
		}
	})
	return sf
}

func getMachineIDByIP() (uint16, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return 0, err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip := ipnet.IP.To4(); ip != nil && len(ip) >= 4 {
				// 取 IP 后 16 位
				return uint16(ip[2])<<8 | uint16(ip[3]), nil
			}
		}
	}
	return 0, errors.New("无法获取有效的本地 IPv4 地址")
}

// NewSnowId 修正：将生成的 uint64 包装进 SnowRest 结构体
func NewSnowId() (SnowRest, error) {
	id, err := GetSnowflake().NextID()
	if err != nil {
		return SnowRest{}, err
	}
	// 这样返回后，调用者可以直接使用 .String() 或 .Int64()
	return SnowRest{Id: id}, nil
}
