package config

import "time"

// 保存一些常量

const (
	//IP               = "192.168.50.163"            // 本地 IP

	IP               = "101.43.179.27"             // 服务器 IP
	Port             = "10240"                     // 服务器端口
	ServerAddr       = "http://" + IP + ":" + Port // 完整服务器地址
	DBName           = "douyin"                    // MySQL 数据库名
	DBUserPass       = "root:root"                 // MySQL 数据库用户名:密码
	DBAddr           = "localhost:3306"            // MySQL 地址
	RedisAddr        = "localhost:6379"            // Redis 地址
	MaxFeedSize      = 30                          // 单次视频流请求最多推送个数
	MaxFeedSizeRedis = 10000                       // 从 MySQL 将视频流读入 Redis 时的最多推送个数
	RedisExp         = 24 * time.Hour              // Redis 数据过期时间
)

var (
	Secret = []byte("mini-douyin") // JWT token 加密
)
