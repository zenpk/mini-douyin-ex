package util

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

func QueryToken(c *gin.Context) string {
	str := c.Query("token")
	if str == "" {
		str = c.PostForm("token")
	}
	return str
}

func QueryUserId(c *gin.Context) int64 {
	str := c.Query("user_id")
	if str == "" {
		str = c.PostForm("user_id")
	}
	id, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func GetTokenUserId(c *gin.Context) int64 {
	value, exists := c.Get("token_user_id")
	if !exists {
		return 0
	} else {
		id := value.(int64)
		return id
	}

}

func QueryId(c *gin.Context, key string) int64 {
	str := c.Query(key)
	if str == "" {
		str = c.PostForm(key)
	}
	id, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return id
}
