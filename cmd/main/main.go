package main

import (
	"github.com/gin-gonic/gin"
	"github.com/zenpk/mini-douyin-ex/cache"
	"github.com/zenpk/mini-douyin-ex/config"
	"github.com/zenpk/mini-douyin-ex/controller"
	"github.com/zenpk/mini-douyin-ex/dal"
	"log"
	"time"
)

func main() {
	// 连接 MySQL 数据库并创建表格
	if err := dal.ConnectDB(); err != nil {
		log.Fatalln(err)
	}
	// 连接 Redis
	if err := cache.ConnectRDB(); err != nil {
		log.Fatalln(err)
	}
	// 将视频流预缓存至 Redis
	if err := cache.WriteFeed(time.Now().Unix()); err != nil {
		log.Fatalln(err)
	}
	// 初始化 Gin
	r := gin.Default()
	controller.InitRouter(r)
	if err := r.Run("0.0.0.0:" + config.Port); err != nil {
		log.Fatalln(err)
	}
}
