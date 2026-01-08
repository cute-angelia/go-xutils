package iredisV2

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

// 建议：所有的 c.ctx 替换为 context.Background() 或从参数传入

func (c *RedisMgr) Expire(key string, timeDur time.Duration) {
	c.opts.client.Expire(context.Background(), key, timeDur)
}

func (c *RedisMgr) Set(key string, val string, expire time.Duration) error {
	// SetEX 是原子操作，2026 年依然推荐使用
	return c.opts.client.SetEX(context.Background(), key, val, expire).Err()
}

// Get 移除 SingleFlight，Redis 自身足够快
func (c *RedisMgr) Get(key string) string {
	val, err := c.opts.client.Get(context.Background(), key).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Printf("[%s] Get Error Key:%s, Err:%v", name, key, err)
		}
		return ""
	}
	return val
}

// GetDataWithSfg 真正的 SingleFlight 使用方式：Redis 不存在时保护下游数据库
// 1. 查 Redis，命中直接返回
// 2. Redis 不存在，查数据库（SingleFlight 确保只有一个请求打到数据库）
// 3. 数据库返回结果，回写 Redis 并返回
func (c *RedisMgr) GetDataWithSfg(key string, fetchFn func() (string, error), expire time.Duration) (string, error) {
	val, err, _ := c.sfg.Do(key, func() (interface{}, error) {
		// 1. 查 Redis
		res, err := c.opts.client.Get(context.Background(), key).Result()
		if err == nil {
			return res, nil
		}
		if !errors.Is(err, redis.Nil) {
			return "", err
		}

		// 2. Redis 不存在，去查数据库（受 SingleFlight 保护，只会有一个请求打到数据库）
		data, err := fetchFn()
		if err != nil {
			return "", err
		}

		// 3. 回写 Redis
		c.opts.client.SetEX(context.Background(), key, data, expire)
		return data, nil
	})
	if err != nil {
		return "", err
	}
	return val.(string), nil
}

func (c *RedisMgr) HMGet(key string, fields ...string) (map[string]interface{}, error) {
	result := make(map[string]interface{}, len(fields))
	vals, err := c.opts.client.HMGet(context.Background(), key, fields...).Result()
	if err != nil {
		return result, err
	}
	for i, field := range fields {
		// 容错处理：即使 field 不存在，vals[i] 也是 nil，map 会安全存储
		result[field] = vals[i]
	}
	return result, nil
}

// LRange 修正：stop 位置
// Redis 的 stop 是包含在内的，所以 stop 应该是 start + perpage - 1
func (c *RedisMgr) LRange(key string, page int64, perpage int64) ([]string, error) {
	if page < 1 {
		page = 1
	}
	start := (page - 1) * perpage
	stop := start + perpage - 1 // 修正：包含末尾坐标

	val, err := c.opts.client.LRange(context.Background(), key, start, stop).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			return []string{}, err
		}
		return []string{}, nil
	}
	return val, nil
}

// LTrimLimit 修正：n 的边界
func (c *RedisMgr) LTrimLimit(key string, n int64) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("limit n must > 0")
	}
	// 保留最新的 n 个元素，通常配合 LPUSH 使用
	return c.opts.client.LTrim(context.Background(), key, 0, n-1).Result()
}
