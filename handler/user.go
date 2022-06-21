package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/zenpk/mini-douyin-ex/dal"
	"github.com/zenpk/mini-douyin-ex/util"
	"log"
	"net/http"
)

type UserLoginResponse struct {
	Response
	UserId int64  `json:"user_id"`
	Token  string `json:"token"`
}

type UserResponse struct {
	Response
	User dal.User `json:"user"`
}

func Register(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")

	// 调用数据层函数
	if id, err := dal.Register(username, password); err != nil {
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: StatusFailed, StatusMsg: err.Error()},
		})
	} else {
		// 根据用户 id 生成 token
		token, err := GenToken(id)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusOK, UserLoginResponse{
				Response: Response{StatusCode: StatusFailed, StatusMsg: "token 生成失败"},
			})
		} else {
			c.JSON(http.StatusOK, UserLoginResponse{
				Response: Response{StatusCode: StatusSuccess, StatusMsg: "注册成功"},
				UserId:   id,
				Token:    token,
			})
		}
	}
}

func Login(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")

	if id, err := dal.Login(username, password); err != nil {
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: StatusFailed, StatusMsg: err.Error()},
		})
	} else {
		if token, err := GenToken(id); err != nil {
			log.Println(err)
			c.JSON(http.StatusOK, UserLoginResponse{
				Response: Response{StatusCode: StatusFailed, StatusMsg: "token 生成失败"},
			})
		} else {
			c.JSON(http.StatusOK, UserLoginResponse{
				Response: Response{StatusCode: StatusSuccess, StatusMsg: "登录成功"},
				UserId:   id,
				Token:    token,
			})
		}
	}
}

func UserInfo(c *gin.Context) {
	userIdA := util.GetTokenUserId(c) // token 中包含的 id 是当前登录用户的 id
	userIdB := util.QueryUserId(c)    // 请求中的 user_id 是查看用户信息的 id
	// 查询数据库判断是否关注了此用户
	if userB, err := dal.GetUserById(userIdB); err != nil {
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: StatusFailed, StatusMsg: "未查询到该用户"},
		})
	} else {
		userB.IsFollow = dal.IsFollowed(userIdA, userIdB)
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: StatusSuccess},
			User:     userB,
		})
	}
}
