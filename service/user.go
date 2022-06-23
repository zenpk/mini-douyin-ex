package service

import (
	"github.com/gin-gonic/gin"
	"github.com/zenpk/mini-douyin-ex/cache"
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
	if user, err := dal.Register(username, password); err != nil {
		log.Println(err)
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: StatusFailed, StatusMsg: "注册失败"},
		})
	} else {
		// 根据用户 id 生成 token
		token, err := GenToken(user.Id)
		// 用户登录后，将用户信息写入缓存
		if err := cache.RegisterLoginUser(user); err != nil {
			log.Println(err)
		}
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusOK, UserLoginResponse{
				Response: Response{StatusCode: StatusFailed, StatusMsg: "token 生成失败"},
			})
		} else {
			c.JSON(http.StatusOK, UserLoginResponse{
				Response: Response{StatusCode: StatusSuccess, StatusMsg: "注册成功"},
				UserId:   user.Id,
				Token:    token,
			})
		}
	}
}

func Login(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")
	if user, err := dal.Login(username, password); err != nil {
		log.Println(err)
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: StatusFailed, StatusMsg: "登录失败"},
		})
	} else {
		if token, err := GenToken(user.Id); err != nil {
			log.Println(err)
			c.JSON(http.StatusOK, UserLoginResponse{
				Response: Response{StatusCode: StatusFailed, StatusMsg: "token 生成失败"},
			})
		} else {
			// 用户登录后，将用户信息写入缓存
			if err := cache.RegisterLoginUser(user); err != nil {
				log.Println(err)
			}
			c.JSON(http.StatusOK, UserLoginResponse{
				Response: Response{StatusCode: StatusSuccess, StatusMsg: "登录成功"},
				UserId:   user.Id,
				Token:    token,
			})
		}
	}
}

// UserInfo 查看某用户的信息
func UserInfo(c *gin.Context) {
	userAId := util.GetTokenUserId(c) // token 中包含的 id 是当前登录用户的 id
	userBId := util.QueryUserId(c)    // 请求中的 user_id 是查看用户信息的 id
	// 先查用户
	userB, err := cache.ReadUser(userBId)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: StatusFailed, StatusMsg: "查看失败"},
		})
	}
	// 再查是否关注
	userB.IsFollow, err = cache.ReadRelation(userAId, userBId)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: StatusFailed, StatusMsg: "查看失败"},
		})
	}
	c.JSON(http.StatusOK, UserResponse{
		Response: Response{StatusCode: StatusSuccess},
		User:     userB,
	})
}
