package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/zenpk/mini-douyin-ex/service"
)

// InitRouter 初始化 Gin 路由
func InitRouter(r *gin.Engine) {
	// public directory is used to serve static resources
	r.Static("/static", "./public")

	apiRouter := r.Group("/douyin")

	// 分模块路由，每个模块中区分是否要使用 JWT 中间件

	// user
	apiRouter.POST("/user/register/", service.Register)
	apiRouter.POST("/user/login/", service.Login)
	apiRouter.GET("/user/", AuthMiddleware(), service.UserInfo)

	// video
	apiRouter.GET("/feed/", AuthMiddlewareAlt(), service.Feed) // 视频流比较特殊，是否登录需要做不同处理
	apiRouter.POST("/publish/action/", AuthMiddleware(), service.Publish)
	apiRouter.GET("/publish/list/", AuthMiddleware(), service.PublishList) // ?

	// 以下功能均需使用 JWT 中间件
	authRouter := apiRouter.Group("/")
	authRouter.Use(AuthMiddleware())
	{
		// favorite
		authRouter.POST("/favorite/action/", service.FavoriteAction)
		authRouter.GET("/favorite/list/", service.FavoriteList)

		// comment
		authRouter.POST("/comment/action/", service.CommentAction)
		authRouter.GET("/comment/list/", service.CommentList)

		// relation
		authRouter.POST("/relation/action/", service.RelationAction)
		authRouter.GET("/relation/follow/list/", service.FollowList)
		authRouter.GET("/relation/follower/list/", service.FollowerList)
	}

}
