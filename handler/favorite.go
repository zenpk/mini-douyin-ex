package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/zenpk/mini-douyin-ex/dal"
	"github.com/zenpk/mini-douyin-ex/util"
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
		ResponseFailed(c, err.Error())
		return
	}
	userId := util.GetTokenUserId(c)
	videoId := util.QueryId(c, "video_id")
	if actionType == ActionFav {
		if err := dal.AddFavorite(userId, videoId); err != nil {
			ResponseFailed(c, err.Error())
		} else {
			ResponseSuccess(c, "点赞成功")
		}
	} else if actionType == ActionFav {
		if err := dal.DeleteFavorite(userId, videoId); err != nil {
			ResponseFailed(c, err.Error())
		} else {
			ResponseSuccess(c, "取消点赞成功")
		}
	} else {
		ResponseFailed(c, "不支持的操作")
	}

}

// FavoriteList 获取点赞视频列表
func FavoriteList(c *gin.Context) {
	userId := util.QueryId(c, "user_id")
	if videoList, err := dal.GetFavoriteList(userId); err != nil {
		c.JSON(http.StatusOK, FavoriteListResponse{
			Response: Response{StatusCode: StatusFailed, StatusMsg: err.Error()},
		})
	} else {
		c.JSON(http.StatusOK, FavoriteListResponse{
			Response:  Response{StatusCode: StatusSuccess},
			VideoList: videoList,
		})
	}
}
