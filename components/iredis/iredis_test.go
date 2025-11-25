package iredis

import (
	"context"
	"errors"
	"github.com/cute-angelia/go-xutils/utils/conf"
	"github.com/go-redis/redis/v8"
	"log"
	"testing"
	"time"
)

func startConnectRedis() {
	conf.LoadConfigFile("./config.toml")
	Load("redis")
}

func TestRedis(t *testing.T) {

	// start
	startConnectRedis()

	redisKey := "abc"

	rdb, _ := GetRedis("cache")
	ctx := context.Background()
	rdb.Set(ctx, redisKey, "world", time.Minute*1)
	val, err := rdb.Get(ctx, redisKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Println(err)
	} else {
		log.Println(redisKey, val)
	}

	t.Log("Exists:", rdb.Exists(ctx, redisKey+"x").Val())
}
