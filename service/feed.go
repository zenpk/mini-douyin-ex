package service

import (
	"github.com/gin-gonic/gin"
	"github.com/zenpk/mini-douyin-ex/cache"
	"github.com/zenpk/mini-douyin-ex/dal"
	"github.com/zenpk/mini-douyin-ex/util"
	"log"
	"net/http"
	"time"
)

type FeedResponse struct {
	Response
	VideoList []dal.Video `json:"video_list"`
	NextTime  int64       `json:"next_time"`
}

// Feed 获取视频流，总体分为三步：获取视频信息（包含作者信息）、获取点赞信息、获取作者关注信息
// 其中每步还需要先从 Redis 查询，未命中再查询 MySQL
func Feed(c *gin.Context) {
	userId := util.GetTokenUserId(c)
	latestTime := util.QueryId(c, "latest_time")
	// 先从 Redis 获取，未命中的部分查找 MySQL
	videoList, err := cache.ReadFeed(latestTime)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusOK, FeedResponse{
			Response: Response{StatusCode: StatusFailed, StatusMsg: "视频流获取失败"},
		})
		return
	}
	if userId != 0 { // 用户已登录，则需要进一步查询点赞信息和关注信息
		for i, video := range videoList {
			// 是否点过赞
			videoList[i].IsFavorite, err = cache.ReadFavorite(userId, video.Id)
			if err != nil {
				log.Println(err)
				c.JSON(http.StatusOK, FeedResponse{
					Response: Response{StatusCode: StatusFailed, StatusMsg: "视频流获取失败"},
				})
				return
			}
			// 是否关注作者
			videoList[i].Author.IsFollow, err = cache.ReadRelation(userId, video.Author.Id)
			if err != nil {
				log.Println(err)
				c.JSON(http.StatusOK, FeedResponse{
					Response: Response{StatusCode: StatusFailed, StatusMsg: "视频流获取失败"},
				})
				return
			}
		}
	}
	c.JSON(http.StatusOK, FeedResponse{
		Response:  Response{StatusCode: StatusSuccess},
		VideoList: videoList,
		NextTime:  time.Now().Unix(),
	})
}
