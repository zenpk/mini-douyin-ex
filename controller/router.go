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
	apiRouter.GET("/feed/", handler.Feed)
	apiRouter.POST("/publish/action/", AuthMiddleware(), handler.Publish)
	apiRouter.GET("/publish/list/", AuthMiddleware(), handler.PublishList)

	// 以下功能均需使用 JWT 中间件
	authRouter := apiRouter.Group("/")
	authRouter.Use(AuthMiddleware())
	{
		// favorite
		apiRouter.POST("/favorite/action/", handler.FavoriteAction)
		apiRouter.GET("/favorite/list/", handler.FavoriteList)

		// comment
		apiRouter.POST("/comment/action/", handler.CommentAction)
		apiRouter.GET("/comment/list/", handler.CommentList)

		// relation
		apiRouter.POST("/relation/action/", handler.RelationAction)
		apiRouter.GET("/relation/follow/list/", handler.FollowList)
		apiRouter.GET("/relation/follower/list/", handler.FollowerList)
	}

}
