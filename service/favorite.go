package service

import (
	"github.com/gin-gonic/gin"
	"github.com/zenpk/mini-douyin-ex/cache"
	"github.com/zenpk/mini-douyin-ex/dal"
	"github.com/zenpk/mini-douyin-ex/util"
	"log"
	"net/http"
	"strconv"
)

type FavoriteListResponse struct {
	Response
	VideoList []dal.Video `json:"video_list"`
}

const (
	ActionFav   = 1
	ActionUnFav = 2
)

// FavoriteAction 点赞操作
func FavoriteAction(c *gin.Context) {
	// 获取操作
	action := c.Query("action_type")
	actionType, err := strconv.Atoi(action)
	if err != nil {
		log.Println(err)
		ResponseFailed(c, "操作失败")
		return
	}
	userId := util.GetTokenUserId(c)
	videoId := util.QueryId(c, "video_id")
	if actionType == ActionFav {
		if err := cache.AddFavorite(userId, videoId); err != nil {
			log.Println(err)
			ResponseFailed(c, "点赞失败")
		} else {
			ResponseSuccess(c, "点赞成功")
		}
	} else if actionType == ActionUnFav {
		if err := cache.DeleteFavorite(userId, videoId); err != nil {
			log.Println(err)
			ResponseFailed(c, "取消点赞失败")
		} else {
			ResponseSuccess(c, "取消点赞成功")
		}
	} else {
		ResponseFailed(c, "不支持的操作")
	}

}

// FavoriteList 获取点赞视频列表
// 由于前端无法从点赞列表中查看视频详情，因此无需考虑作者等信息
func FavoriteList(c *gin.Context) {
	var videoList []dal.Video
	userId := util.QueryId(c, "user_id")
	favoriteList, err := dal.GetFavoriteListByUserId(userId)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusOK, FavoriteListResponse{
			Response: Response{StatusCode: StatusFailed, StatusMsg: "获取点赞视频列表失败"},
		})
		return
	}
	for _, favorite := range favoriteList {
		video, err := cache.ReadVideo(favorite.VideoId)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusOK, FavoriteListResponse{
				Response: Response{StatusCode: StatusFailed, StatusMsg: "获取点赞视频列表失败"},
			})
			return
		}
		videoList = append(videoList, video)
	}
	c.JSON(http.StatusOK, FavoriteListResponse{
		Response:  Response{StatusCode: StatusSuccess},
		VideoList: videoList,
	})
}
