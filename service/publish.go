package service

import (
	"github.com/gin-gonic/gin"
	"github.com/zenpk/mini-douyin-ex/cache"
	"github.com/zenpk/mini-douyin-ex/dal"
	"github.com/zenpk/mini-douyin-ex/util"
	"log"
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
		log.Println(err)
		ResponseFailed(c, "读取视频失败")
		return
	}
	filename := filepath.Base(data.Filename)
	if video, err := dal.Publish(c, userId, title, filename, data); err != nil {
		log.Println(err)
		ResponseFailed(c, "上传失败")
	} else {
		// 将视频写入 Redis，如果失败也不需要回滚，下次重新读取即可
		if err := cache.AddVideo(video); err != nil {
			log.Println(err)
		}
		ResponseSuccess(c, "上传成功")
	}
}

// PublishList 获取当前用户的视频列表
func PublishList(c *gin.Context) {
	userAId := util.GetTokenUserId(c)
	userBId := util.QueryUserId(c)
	if videoList, err := cache.ReadPublishList(userAId, userBId); err != nil {
		log.Println(err)
		c.JSON(http.StatusOK, VideoListResponse{
			Response: Response{
				StatusCode: StatusFailed,
				StatusMsg:  "获取失败",
			},
		})
	} else {
		c.JSON(http.StatusOK, VideoListResponse{
			Response: Response{
				StatusCode: StatusSuccess,
			},
			VideoList: videoList,
		})
	}
}
