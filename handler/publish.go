package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/zenpk/mini-douyin-ex/dal"
	"github.com/zenpk/mini-douyin-ex/util"
	"net/http"
	"path/filepath"
)

type VideoListResponse struct {
	Response
	VideoList []dal.Video `json:"video_list"`
}

// Publish 前端传入视频、token
func Publish(c *gin.Context) {
	// 上传者 id
	userId := util.GetTokenUserId(c)
	// 视频标题
	title := c.PostForm("title")
	// 读取视频
	data, err := c.FormFile("data")
	if err != nil {
		ResponseFailed(c, err.Error())
		return
	}
	filename := filepath.Base(data.Filename)
	if err := dal.Publish(c, userId, title, filename, data); err != nil {
		ResponseFailed(c, "上传失败，原因："+err.Error())
	} else {
		ResponseSuccess(c, "上传成功")
	}
}

// PublishList 显示当前用户投稿的所有视频，目前的实现方式是直接从所有视频的表里根据 user_id 选取
// 效率可能较低？
func PublishList(c *gin.Context) {
	userId := util.GetTokenUserId(c)
	videoList := dal.GetPublishList(userId)
	c.JSON(http.StatusOK, VideoListResponse{
		Response: Response{
			StatusCode: StatusSuccess,
		},
		VideoList: videoList,
	})
}
