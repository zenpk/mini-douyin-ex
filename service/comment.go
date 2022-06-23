package service

import (
	"github.com/gin-gonic/gin"
	"github.com/zenpk/mini-douyin-ex/cache"
	"github.com/zenpk/mini-douyin-ex/dal"
	"github.com/zenpk/mini-douyin-ex/util"
	"log"
	"net/http"
	"strconv"
	"time"
)

type CommentListResponse struct {
	Response
	CommentList []dal.Comment `json:"comment_list"`
}

const (
	ActionAddComment    = 1
	ActionDeleteComment = 2
)

// CommentAction 发表评论、删除评论
func CommentAction(c *gin.Context) {
	// 获取操作
	action := c.Query("action_type")
	actionType, err := strconv.Atoi(action)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusOK, CommentListResponse{
			Response: Response{StatusCode: StatusFailed, StatusMsg: "操作出错"},
		})
		return
	}
	userId := util.GetTokenUserId(c)
	videoId := util.QueryId(c, "video_id")
	commentId := util.QueryId(c, "comment_id")
	commentText := c.Query("comment_text")
	comment := dal.Comment{
		UserId:     userId,
		VideoId:    videoId,
		Content:    commentText,
		CreateDate: time.Now().Format("01-02 15:04:05"),
	}
	if actionType == ActionAddComment {
		if err := cache.AddComment(comment); err != nil {
			log.Println(err)
			c.JSON(http.StatusOK, CommentListResponse{
				Response: Response{StatusCode: StatusFailed, StatusMsg: "评论失败"},
			})
		} else {
			c.JSON(http.StatusOK, CommentListResponse{
				Response:    Response{StatusCode: StatusSuccess, StatusMsg: "评论成功"},
				CommentList: []dal.Comment{comment},
			})
		}
	} else if actionType == ActionDeleteComment {
		if err := cache.DeleteComment(userId, videoId, commentId); err != nil {
			log.Println(err)
			c.JSON(http.StatusOK, CommentListResponse{
				Response: Response{StatusCode: StatusFailed, StatusMsg: "删除评论失败"},
			})
		} else {
			c.JSON(http.StatusOK, CommentListResponse{
				Response: Response{StatusCode: StatusSuccess, StatusMsg: "删除评论成功"},
			})
		}
	} else {
		c.JSON(http.StatusOK, CommentListResponse{
			Response: Response{StatusCode: StatusFailed, StatusMsg: "不支持的操作"},
		})
	}
}

// CommentList 获取评论列表
func CommentList(c *gin.Context) {
	videoId := util.QueryId(c, "video_id")
	if commentList, err := cache.ReadCommentList(videoId); err != nil {
		// TODO 进一步读取关注信息？
		log.Println(err)
		c.JSON(http.StatusOK, CommentListResponse{
			Response: Response{StatusCode: StatusSuccess, StatusMsg: "获取评论列表失败"},
		})
	} else {
		c.JSON(http.StatusOK, CommentListResponse{
			Response:    Response{StatusCode: StatusSuccess},
			CommentList: commentList,
		})
	}
}
