package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type FeedResponse struct {
	Response
	VideoList []Video `json:"video_list,omitempty"`
	NextTime  int64   `json:"next_time,omitempty"`
}

// Feed 推送最新的 30 个视频
func Feed(c *gin.Context) {
	user, _ := GetUserByToken(c) // 未登录也可推送，因此不判断 token 有效性
	var videoList []Video
	DB.Preload("Author").Order("id desc").Limit(30).Find(&videoList)
	for i, v := range videoList {
		// 查找是否存在一条当前用户给该视频点赞的记录
		rows := DB.Where("video_id=?", v.Id).Where("user_id=?", user.Id).Find(&Favorite{}).RowsAffected
		videoList[i].IsFavorite = rows > 0 // 查找到了点赞记录
	}
	c.JSON(http.StatusOK, FeedResponse{
		Response:  Response{StatusCode: 0},
		VideoList: videoList,
		NextTime:  time.Now().Unix(),
	})
}
