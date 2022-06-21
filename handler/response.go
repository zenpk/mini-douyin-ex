package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	StatusCode int32  `json:"status_code"`
	StatusMsg  string `json:"status_msg,omitempty"`
}

const (
	StatusSuccess = 0
	StatusFailed  = 1
)

func ResponseFailed(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{StatusCode: StatusFailed, StatusMsg: msg})
}

func ResponseSuccess(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{StatusCode: StatusSuccess, StatusMsg: msg})
}
