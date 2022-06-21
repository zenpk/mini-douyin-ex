package dal

import (
	"github.com/zenpk/mini-douyin-ex/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

// 连接数据库，并创建所有表

var DB *gorm.DB

func ConnectDB() {
	// 连接数据库
	dsn := config.DBUserPass + "@tcp(localhost:3306)/" + config.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	// 创建 User, Video, Comment, Favorite, Relation 表
	if err := DB.AutoMigrate(&User{}); err != nil {
		log.Fatal(err)
	}
	if err := DB.AutoMigrate(&Video{}); err != nil {
		log.Fatal(err)
	}
	if err := DB.AutoMigrate(&Comment{}); err != nil {
		log.Fatal(err)
	}
	if err := DB.AutoMigrate(&Favorite{}); err != nil {
		log.Fatal(err)
	}
	if err := DB.AutoMigrate(&Relation{}); err != nil {
		log.Fatal(err)
	}
}
