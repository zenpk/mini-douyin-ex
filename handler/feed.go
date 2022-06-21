package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/zenpk/mini-douyin-ex/dal"
	"github.com/zenpk/mini-douyin-ex/util"
	"net/http"
	"time"
)

type FeedResponse struct {
	Response
	VideoList []dal.Video `json:"video_list"`
	NextTime  int64       `json:"next_time"`
}

// Feed 推送最新的 30 个视频
func Feed(c *gin.Context) {
	userId := util.GetTokenUserId(c)
	videoList := dal.GetFeed(userId)
	c.JSON(http.StatusOK, FeedResponse{
		Response:  Response{StatusCode: StatusSuccess},
		VideoList: videoList,
		NextTime:  time.Now().Unix(),
	})
}
