package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/zenpk/mini-douyin-ex/dal"
	"github.com/zenpk/mini-douyin-ex/util"
	"net/http"
	"strconv"
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
		c.JSON(http.StatusOK, CommentListResponse{
			Response: Response{StatusCode: StatusFailed, StatusMsg: err.Error()},
		})
		return
	}
	userId := util.GetTokenUserId(c)
	videoId := util.QueryId(c, "video_id")
	commentId := util.QueryId(c, "comment_id")
	commentText := c.Query("comment_text")
	if actionType == ActionAddComment {
		if comment, err := dal.AddComment(userId, videoId, commentText); err != nil {
			c.JSON(http.StatusOK, CommentListResponse{
				Response: Response{StatusCode: StatusFailed, StatusMsg: err.Error()},
			})
		} else {
			c.JSON(http.StatusOK, CommentListResponse{
				Response:    Response{StatusCode: StatusSuccess, StatusMsg: "评论成功"},
				CommentList: []dal.Comment{comment},
			})
		}
	} else if actionType == ActionDeleteComment {
		if err := dal.DeleteComment(userId, videoId, commentId); err != nil {
			c.JSON(http.StatusOK, CommentListResponse{
				Response: Response{StatusCode: StatusFailed, StatusMsg: err.Error()},
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
	if commentList, err := dal.GetCommentList(videoId); err != nil {
		c.JSON(http.StatusOK, CommentListResponse{
			Response: Response{StatusCode: StatusSuccess, StatusMsg: err.Error()},
		})
	} else {
		c.JSON(http.StatusOK, CommentListResponse{
			Response:    Response{StatusCode: StatusSuccess},
			CommentList: commentList,
		})
	}
}
