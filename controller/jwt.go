package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zenpk/mini-douyin-ex/handler"
	"github.com/zenpk/mini-douyin-ex/util"
	"net/http"
)

var (
	Secret = []byte("mini-douyin")
)

// GenToken 根据用户 id 生成 token，为避免重复登录带来的麻烦，暂不加入过期时间
func GenToken(userId int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": userId,
	})
	signedToken, err := token.SignedString(Secret)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

// parseToken 验证 token 并提取 token 中的用户 id
func parseToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return Secret, nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["id"].(int64), nil
	} else {
		return 0, err
	}
}

// AuthMiddleware 用于 token 校验
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := util.QueryToken(c)
		userId, err := parseToken(token)
		if err != nil || userId == int64(0) {
			c.JSON(http.StatusOK, handler.Response{
				StatusCode: handler.StatusFailed,
				StatusMsg:  "token 校验失败",
			})
			c.Abort()
		} else {
			c.Set("token_user_id", userId) // 避免和 "user_id" 冲突
			c.Next()
		}
	}
}
