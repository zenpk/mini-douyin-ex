package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/zenpk/mini-douyin-ex/config"
)

var RDB *redis.Client
var CTX = context.Background()

// ConnectRDB 连接 Redis
func ConnectRDB() error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	if _, err := RDB.Ping(CTX).Result(); err != nil {
		return err
	}
	if err := RDB.FlushAll(CTX).Err(); err != nil { // 初始化，方便测试，实际应用中可删除此行
		return err
	}
	return nil
}
