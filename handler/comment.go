package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

type CommentListResponse struct {
	Response
	CommentList []Comment `json:"comment_list,omitempty"`
}

// CommentAction 发表评论、删除评论
func CommentAction(c *gin.Context) {
	user, tokenValid := GetUserByToken(c)
	//判断用户是否登录
	if !tokenValid {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  "You haven't logged in yet",
		})
		return
	}
	action := c.Query("action_type")
	actionType, err := strconv.Atoi(action)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}

	if actionType == 1 {
		PostComment(c, user)
	} else {
		DeleteComment(c, user)
	}

}

// DeleteComment 删除评论
func DeleteComment(c *gin.Context, user User) {
	// 获取需要删除的评论 Id 和对应视频 Id
	commentId := GetId(c, "comment_id")
	videoId := GetId(c, "video_id")

	// 改进：判断该评论是否属于此用户，不属于则返回删除失败
	var comment Comment
	DB.Where("id=?", commentId).First(&comment)
	if comment.UserId != user.Id {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  "you cannot delete this comment",
		})
		return
	}

	//（1）直接在数据库中删除（2）comment中设置一个deleted 列，用bool表示是否删除
	//目前实现的是第一种
	//开启数据库事务，在comments中添加记录，在videos中更改评论数目
	DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id=?", commentId).Delete(&Comment{}).Error; err != nil {
			// 返回任何错误都会回滚事务
			return err
		}
		tx.Model(&Video{}).Where("id=?", videoId).UpdateColumn("comment_count", gorm.Expr("comment_count - ?", 1))
		// 返回 nil 提交事务
		return nil
	})
	c.JSON(http.StatusOK, Response{
		StatusCode: 0,
		StatusMsg:  "comment deleted successfully",
	})
}

// PostComment 发表评论
func PostComment(c *gin.Context, user User) {
	//读取评论内容
	context := c.Query("comment_text")
	//用户未输入评论内容
	if context == "" {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  "请输入评论内容",
		})
		return
	}
	videoId := GetId(c, "video_id")
	//日期MM-DD
	timeFormat := time.Now().Format("01-02")
	comment := Comment{
		User:       user,
		VideoId:    videoId,
		UserId:     user.Id,
		Content:    context,
		CreateDate: timeFormat,
	}
	//开启数据库事务，在comments中添加记录，在videos中更改评论数目
	DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&comment).Error; err != nil {
			// 返回任何错误都会回滚事务
			return err
		}
		tx.Model(&Video{}).Where("id=?", videoId).UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1))
		// 返回 nil 提交事务
		return nil
	})

	//文档中标明不需要拉取评论列表，数据库中的自增id无法获取
	//目前默认每次处理一条comment，所以数组只存入一条评论数据
	commentList := []Comment{comment}

	c.JSON(http.StatusOK, CommentListResponse{
		Response: Response{
			StatusCode: 0,
			StatusMsg:  "comment posted successfully"},
		CommentList: commentList,
	})
}

// CommentList 获取评论列表
func CommentList(c *gin.Context) {
	token := c.Query("token")
	//判断用户是否登录
	if token == "" {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  "You haven't logged in yet",
		})
		return
	}

	videoId := GetId(c, "video_id")

	var commentList []Comment
	DB.Preload("User").Where("video_id=?", videoId).Find(&commentList)

	c.JSON(http.StatusOK, CommentListResponse{
		Response: Response{
			StatusCode: 0,
			StatusMsg:  "success",
		},
		CommentList: commentList,
	})
}
