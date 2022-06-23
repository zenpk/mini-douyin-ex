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
		log.Println(err)
		ResponseFailed(c, "操作出错")
		return
	}
	userId := util.GetTokenUserId(c)
	toUserId := util.QueryId(c, "to_user_id")
	// 先检查是否已关注
	isFollow, err := cache.ReadRelation(userId, toUserId)
	if err != nil {
		log.Println(err)
		ResponseFailed(c, "操作出错")
		return
	}
	if action == ActionFollow {
		if isFollow {
			ResponseFailed(c, "已关注过该用户")
		} else if err := cache.AddFollow(userId, toUserId); err != nil {
			log.Println(err)
			ResponseFailed(c, "关注失败")
		} else {
			ResponseSuccess(c, "关注成功")
		}
	} else if action == ActionUnfollow {
		if !isFollow {
			ResponseFailed(c, "未关注过该用户")
		} else if err := cache.DeleteFollow(userId, toUserId); err != nil {
			log.Println(err)
			ResponseFailed(c, "取消关注失败")
		} else {
			ResponseSuccess(c, "取消关注成功")
		}
	} else {
		ResponseFailed(c, "不支持的操作")
	}
}

// FollowList 展示查询用户的关注列表
func FollowList(c *gin.Context) {
	userAId := util.GetTokenUserId(c)
	userBId := util.QueryUserId(c)
	// 先查 Redis，未命中再查数据库，查询过程中应更新 isFollow 信息
	if followList, err := cache.ReadFollow(userAId, userBId); err != nil {
		log.Println(err)
		c.JSON(http.StatusOK, UserListResponse{
			Response: Response{StatusCode: StatusFailed, StatusMsg: "查询失败"},
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
	userAId := util.GetTokenUserId(c)
	userBId := util.QueryUserId(c)
	// 先查 Redis，未命中再查数据库，查询过程中应更新 isFollow 信息
	if followList, err := cache.ReadFollower(userAId, userBId); err != nil {
		log.Println(err)
		c.JSON(http.StatusOK, UserListResponse{
			Response: Response{StatusCode: StatusFailed, StatusMsg: "查询失败"},
		})
	} else {
		c.JSON(http.StatusOK, UserListResponse{
			Response: Response{StatusCode: StatusSuccess},
			UserList: followList,
		})
	}
}
