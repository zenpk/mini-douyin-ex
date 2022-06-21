package handler

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/zenpk/mini-douyin-ex/config"
)

// GenToken 根据用户 id 生成 token，为避免重复登录带来的麻烦，暂不加入过期时间
func GenToken(userId int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": userId,
	})
	signedToken, err := token.SignedString(config.Secret)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}
