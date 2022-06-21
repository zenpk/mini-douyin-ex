package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zenpk/mini-douyin-ex/config"
	"github.com/zenpk/mini-douyin-ex/handler"
	"github.com/zenpk/mini-douyin-ex/util"
)

// parseToken 验证 token 并提取 token 中的用户 id
func parseToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return config.Secret, nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return int64(claims["id"].(float64)), nil
	} else {
		return 0, err
	}
}

// AuthMiddleware 用于 token 校验
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := util.QueryToken(c)
		if token == "" {
			handler.ResponseFailed(c, "请先登录或注册")
			return
		}
		userId, err := parseToken(token)
		if err != nil || userId == int64(0) {
			handler.ResponseFailed(c, "token 校验失败")
			c.Abort()
		} else {
			c.Set("token_user_id", userId) // 避免和 "user_id" 冲突
			c.Next()
		}
	}
}

// AuthMiddlewareAlt 无论是否登录均可以通过验证（专门用于处理 feed 请求）
func AuthMiddlewareAlt() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := util.QueryToken(c)
		if token == "" { // 未登录
			c.Next()
			return
		}
		userId, err := parseToken(token)
		if err != nil || userId == int64(0) { // 登录了但校验失败
			handler.ResponseFailed(c, "token 校验失败")
			c.Abort()
		} else {
			c.Set("token_user_id", userId) // 避免和 "user_id" 冲突
			c.Next()
		}
	}
}
