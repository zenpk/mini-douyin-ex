package main

import (
	"github.com/gin-gonic/gin"
	"github.com/zenpk/mini-douyin-ex/config"
	"github.com/zenpk/mini-douyin-ex/controller"
	"github.com/zenpk/mini-douyin-ex/dal"
	"log"
)

func main() {
	r := gin.Default()

	dal.ConnectDB() // 连接 MySQL 数据库并创建表格
	controller.InitRouter(r)

	if err := r.Run("0.0.0.0:" + config.Port); err != nil {
		log.Fatalln(err)
	}
}
