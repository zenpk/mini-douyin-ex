package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zenpk/mini-douyin-ex/controller"
	"github.com/zenpk/mini-douyin-ex/dal"
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
			Response: Response{StatusCode: StatusFailed, StatusMsg: fmt.Sprintln(err)},
		})
	} else {
		// 根据用户 id 生成 token
		token, err := controller.GenToken(id)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusOK, Response{StatusCode: StatusFailed, StatusMsg: "token 生成失败"})
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
		c.JSON(http.StatusOK, Response{StatusCode: StatusFailed, StatusMsg: fmt.Sprintln(err)})
	} else {
		token, err := controller.GenToken(id)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusOK, Response{StatusCode: StatusFailed, StatusMsg: "token 生成失败"})
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
	userA, tokenValid := GetUserByToken(c)
	if !tokenValid {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  "You haven't logged in yet",
		})
	}
	userB, _ := GetUserById(c)
	// 查询数据库判断是否关注了此用户
	rows := DB.Find(&Relation{}).Where("user_a_id=?", userA.Id).Where("user_b_id=?", userB.Id).RowsAffected
	userB.IsFollow = rows > 0 // 查询到关注记录，则返回 true
	c.JSON(http.StatusOK, UserResponse{
		Response: Response{StatusCode: 0},
		User:     userB,
	})
}
