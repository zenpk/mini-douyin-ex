package dal

import (
	"github.com/zenpk/mini-douyin-ex/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDB 连接 MySQL
func ConnectDB() error {
	dsn := config.DBUserPass + "@tcp(" + config.DBAddr + ")/" + config.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	// 创建 User, Video, Comment, Favorite, Relation 表
	if err := DB.AutoMigrate(&User{}); err != nil {
		return err
	}
	if err := DB.AutoMigrate(&Video{}); err != nil {
		return err
	}
	if err := DB.AutoMigrate(&Comment{}); err != nil {
		return err
	}
	if err := DB.AutoMigrate(&Favorite{}); err != nil {
		return err
	}
	if err := DB.AutoMigrate(&Relation{}); err != nil {
		return err
	}
	return nil
}
