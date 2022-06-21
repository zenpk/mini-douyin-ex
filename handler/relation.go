package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

type UserListResponse struct {
	Response
	UserList []User `json:"user_list"`
}

// RelationAction 关注取关操作
func RelationAction(c *gin.Context) {
	user, tokenValid := GetUserByToken(c)
	if !tokenValid {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  "You haven't logged in yet",
		})
		return
	}
	actionStr := c.Query("action_type") // 1:关注，2:取关
	action, _ := strconv.Atoi(actionStr)
	if action == 1 {
		Follow(c, user)
	} else {
		Unfollow(c, user)
	}
}

func Follow(c *gin.Context, user User) {
	// 因为前端没有正确传入 user_id，因此不获取 user_id

	userIdTo := GetId(c, "to_user_id") // UserB

	relation := Relation{
		UserAId: user.Id,
		UserBId: userIdTo,
	}
	// 开启数据库事务，在 relations 中添加记录，在 users 中更改关注数
	DB.Transaction(func(tx *gorm.DB) error {
		// 创建关注关系
		if err := tx.Create(&relation).Error; err != nil {
			// 返回任何错误都会回滚事务
			return err
		}
		// 增加关注数、被关注数
		tx.Model(&User{}).Where("id=?", user.Id).UpdateColumn("follow_count", gorm.Expr("follow_count + ?", 1))
		tx.Model(&User{}).Where("id=?", userIdTo).UpdateColumn("follower_count", gorm.Expr("follower_count + ?", 1))
		// 返回 nil 提交事务
		return nil
	})
	c.JSON(http.StatusOK, Response{
		StatusCode: 1,
		StatusMsg:  "Successfully followed",
	})
}

func Unfollow(c *gin.Context, user User) {
	// 因为前端没有正确传入 user_id，因此不获取 user_id

	userIdTo := GetId(c, "to_user_id") // UserB

	// 开启数据库事务，在 relations 中添加记录，在 users 中更改关注数
	DB.Transaction(func(tx *gorm.DB) error {
		// 删除关注关系
		if err := tx.Where("user_a_id=?", user.Id).Where("user_b_id=?", userIdTo).Delete(&Relation{}).Error; err != nil {
			// 返回任何错误都会回滚事务
			return err
		}
		// 减少关注数、被关注数
		tx.Model(&User{}).Where("id=?", user.Id).UpdateColumn("follow_count", gorm.Expr("follow_count - ?", 1))
		tx.Model(&User{}).Where("id=?", userIdTo).UpdateColumn("follower_count", gorm.Expr("follower_count - ?", 1))
		// 返回 nil 提交事务
		return nil
	})
	c.JSON(http.StatusOK, Response{
		StatusCode: 1,
		StatusMsg:  "Successfully unfollowed",
	})
}

// FollowList 展示查询用户的关注列表
func FollowList(c *gin.Context) {
	user, tokenValid := GetUserByToken(c)
	if !tokenValid {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  "You haven't logged in yet",
		})
		return
	}
	var followList []Relation
	// 加载 UserB 即加载当前用户关注的用户
	DB.Preload("UserB").Where("user_a_id=?", user.Id).Find(&followList)
	// 这里直接暴力复制了，不知道 Go 语言有无更好的方法可以提取结构体数组中的元素
	followUserList := make([]User, len(followList))
	for i, f := range followList {
		followUserList[i] = f.UserB
	}
	c.JSON(http.StatusOK, UserListResponse{
		Response: Response{
			StatusCode: 0,
		},
		UserList: followUserList,
	})
}

// FollowerList 展示查询用户的粉丝列表
func FollowerList(c *gin.Context) {
	user, tokenValid := GetUserByToken(c)
	if !tokenValid {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  "You haven't logged in yet",
		})
		return
	}
	var followerList []Relation
	// 加载 UserA 即加载当前用户的粉丝
	DB.Preload("UserA").Where("user_b_id=?", user.Id).Find(&followerList)
	// 这里直接暴力复制了，不知道 Go 语言有无更好的方法可以提取结构体数组中的元素
	followerUserList := make([]User, len(followerList))
	for i, f := range followerList {
		followerUserList[i] = f.UserA
	}
	c.JSON(http.StatusOK, UserListResponse{
		Response: Response{
			StatusCode: 0,
		},
		UserList: followerUserList,
	})
}
