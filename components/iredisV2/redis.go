package iredisV2

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"golang.org/x/sync/singleflight"
	"log"
	"sync"
	"time"
)

const name = "redis"

var redisMgrPools sync.Map

type RedisMgr struct {
	client redis.UniversalClient
	opts   *options
	sfg    singleflight.Group
	alias  string
}

// RedisInit 初始化
func RedisInit(alias string, opts ...Option) *RedisMgr {
	// 1. 雙重檢查 (DCL) 防止併發初始化造成的連接洩漏
	if v, ok := redisMgrPools.Load(alias); ok {
		return v.(*RedisMgr)
	}

	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	if o.client == nil {
		client := redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:        o.addrs,
			DB:           o.db,
			Username:     o.username,
			Password:     o.password,
			MaxRetries:   o.maxRetries,
			PoolSize:     o.poolSize,
			MinIdleConns: o.minIdleConns,
			DialTimeout:  o.dialTimeout,
		})

		// 使用一個短暫的超時 Context 進行 Ping 測試
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := client.Ping(ctx).Err(); err != nil {
			log.Printf("[%s] ❌ Redis Alias:%s 連接失敗: %v", name, alias, err)
			panic(err) // 啟動時連不上通常需要報警
		}
		o.client = client
	}

	l := &RedisMgr{
		client: o.client,
		opts:   o,
		alias:  alias,
	}

	// 存入池中
	actual, loaded := redisMgrPools.LoadOrStore(alias, l)
	mgr := actual.(*RedisMgr)

	if loaded {
		// 如果是被其他協程搶先存入了，則關閉當前重複創建的連接
		if l.client != mgr.client {
			l.client.Close()
		}
		mgr.printRedisPool("命中池", mgr.client.PoolStats())
	} else {
		mgr.printRedisPool("初始化成功", mgr.client.PoolStats())
	}

	return mgr
}

// GetClient 直接獲取原生客戶端
func GetClient(alias string) (redis.UniversalClient, error) {
	v, ok := redisMgrPools.Load(alias)
	if !ok {
		return nil, fmt.Errorf("redis: alias %s 未初始化", alias)
	}
	return v.(*RedisMgr).client, nil
}

func (c *RedisMgr) printRedisPool(msg string, stats *redis.PoolStats) {
	log.Printf("[%s] Alias:%s, Addrs:%v, TotalConns:%d, IdleConns:%d, %s \n",
		name,
		c.alias,
		c.opts.addrs,
		stats.TotalConns,
		stats.IdleConns,
		msg,
	)
}
