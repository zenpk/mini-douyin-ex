package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/zenpk/mini-douyin-ex/dal"
	"github.com/zenpk/mini-douyin-ex/util"
	"net/http"
	"strconv"
)

type UserListResponse struct {
	Response
	UserList []dal.User `json:"user_list"`
}

const (
	ActionFollow   = 1
	ActionUnfollow = 2
)

// RelationAction 关注取关操作
func RelationAction(c *gin.Context) {
	actionStr := c.Query("action_type")
	action, err := strconv.Atoi(actionStr)
	if err != nil {
		ResponseFailed(c, err.Error())
		return
	}
	userId := util.GetTokenUserId(c)
	toUserId := util.QueryId(c, "to_user_id")
	if action == ActionFollow {
		if err := dal.Follow(userId, toUserId); err != nil {
			ResponseFailed(c, err.Error())
		}
		ResponseSuccess(c, "关注成功")
	} else if action == ActionUnfollow {
		if err := dal.Unfollow(userId, toUserId); err != nil {
			ResponseFailed(c, err.Error())
		}
		ResponseSuccess(c, "取消关注成功")
	} else {
		ResponseFailed(c, "不支持的操作")
	}
}

// FollowList 展示查询用户的关注列表
func FollowList(c *gin.Context) {
	userId := util.QueryUserId(c)
	if followList, err := dal.GetFollowList(userId); err != nil {
		c.JSON(http.StatusOK, UserListResponse{
			Response: Response{StatusCode: StatusFailed, StatusMsg: err.Error()},
		})
	} else {
		c.JSON(http.StatusOK, UserListResponse{
			Response: Response{StatusCode: StatusSuccess},
			UserList: followList,
		})
	}
}

// FollowerList 展示查询用户的粉丝列表
func FollowerList(c *gin.Context) {
	userId := util.QueryUserId(c)
	if followerList, err := dal.GetFollowerList(userId); err != nil {
		c.JSON(http.StatusOK, UserListResponse{
			Response: Response{StatusCode: StatusFailed, StatusMsg: err.Error()},
		})
	} else {
		c.JSON(http.StatusOK, UserListResponse{
			Response: Response{StatusCode: StatusSuccess},
			UserList: followerList,
		})
	}
}
