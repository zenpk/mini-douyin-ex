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
	id, _ := strconv.ParseInt(str, 10, 64)
	return id
}

func QueryVideoId(c *gin.Context) int64 {
	str := c.Query("video_id")
	if str == "" {
		str = c.PostForm("video_id")
	}
	id, _ := strconv.ParseInt(str, 10, 64)
	return id
}

func QueryCommentId(c *gin.Context) int64 {
	str := c.Query("comment_id")
	if str == "" {
		str = c.PostForm("comment_id")
	}
	id, _ := strconv.ParseInt(str, 10, 64)
	return id
}
