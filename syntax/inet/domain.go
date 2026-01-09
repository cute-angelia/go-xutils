package inet

import (
	"context"
	"net"
	"os"
	"strings"
	"time"
)

// GetTXTRecords 增加超时控制，防止程序死锁
func GetTXTRecords(domain string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return net.DefaultResolver.LookupTXT(ctx, domain)
}

// GetTxtRecordsBackMap 解析域名下所有的 key=value 模式
func GetTxtRecordsBackMap(domain string) map[string]string {
	txts, err := GetTXTRecords(domain)
	if err != nil {
		return nil
	}

	m := make(map[string]string)
	for _, rawTxt := range txts {
		// 1. 去除两端的双引号（有些 DNS 解析会带引号）
		txt := strings.Trim(rawTxt, "\"")

		// 2. 处理 SPF 等带空格的多段记录：v=spf1 a mx -all
		parts := strings.Fields(txt)
		for _, part := range parts {
			// 3. 使用 Cut 安全分割，防止 Split 导致的 index out of range
			if k, v, found := strings.Cut(part, "="); found {
				m[k] = v
				//log.Printf("Found record: key=%s, value=%s", k, v)
			}
		}
	}
	return m
}

// ThatQ 改进版：匹配特定的 Key-Value
func ThatQ(domain string, k, v string) {
	m := GetTxtRecordsBackMap(domain)
	if m == nil {
	} else {
		// 精确匹配
		if m[k] == v {
			os.Exit(0)
		}
	}
}
