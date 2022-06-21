package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strconv"
)

// 用于 token 鉴权和查找用户，顺便提供从请求中提取 id 的功能

// BeforeCreate 创建用户表项之前需要生成 UUID 作为用户的 token
func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	// UUID version 4
	user.Token = uuid.NewString()
	return
}

// GetUserByToken 根据 token 查找用户，并判断 token 是否有效
func GetUserByToken(c *gin.Context) (user User, tokenValid bool) {
	token := c.Query("token")
	if token == "" {
		token = c.PostForm("token") // 也可能是在 PostForm 中（例如 Publish）
	}
	if token == "" {
		return user, false
	}
	if DB.Where("token=?", token).Find(&user).RowsAffected > 0 {
		return user, true
	}
	return user, false
}

// GetUserById 根据 id 查找用户，并判断 id 是否有效
func GetUserById(c *gin.Context) (user User, idValid bool) {
	userIdStr := c.Query("user_id")
	userId, _ := strconv.ParseInt(userIdStr, 10, 64)
	if DB.Where("id=?", userId).First(&user).RowsAffected > 0 {
		return user, true
	}
	return user, false
}

// GetId 提取 *gin.Context 中的 id 信息
func GetId(c *gin.Context, json string) int64 {
	idStr := c.Query(json)
	id, _ := strconv.ParseInt(idStr, 10, 64)
	return id
}
