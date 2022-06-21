package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/zenpk/mini-douyin-ex/handler"
)

func InitRouter(r *gin.Engine) {
	// public directory is used to serve static resources
	r.Static("/static", "./public")

	apiRouter := r.Group("/douyin")

	// 分模块路由，每个模块中区分是否要使用 JWT 中间件

	// user
	apiRouter.POST("/user/register/", handler.Register)
	apiRouter.POST("/user/login/", handler.Login)
	apiRouter.GET("/user/", AuthMiddleware(), handler.UserInfo)

	// video
	apiRouter.GET("/feed/", AuthMiddlewareAlt(), handler.Feed) // 视频流比较特殊，是否登录需要做不同处理
	apiRouter.POST("/publish/action/", AuthMiddleware(), handler.Publish)
	apiRouter.GET("/publish/list/", AuthMiddleware(), handler.PublishList) // ?

	// 以下功能均需使用 JWT 中间件
	authRouter := apiRouter.Group("/")
	authRouter.Use(AuthMiddleware())
	{
		// favorite
		authRouter.POST("/favorite/action/", handler.FavoriteAction)
		authRouter.GET("/favorite/list/", handler.FavoriteList)

		// comment
		authRouter.POST("/comment/action/", handler.CommentAction)
		authRouter.GET("/comment/list/", handler.CommentList)

		// relation
		authRouter.POST("/relation/action/", handler.RelationAction)
		authRouter.GET("/relation/follow/list/", handler.FollowList)
		authRouter.GET("/relation/follower/list/", handler.FollowerList)
	}

}
