package cache

import (
	"github.com/go-redis/redis/v8"
	"github.com/zenpk/mini-douyin-ex/config"
	"github.com/zenpk/mini-douyin-ex/dal"
	"strconv"
)

// WriteFeed 将视频流 **首次** 写入 Redis
// 由于仅在首次启动时调用，因此该函数和其子调用在写入时均不用查询是否已存在
func WriteFeed(latestTime int64) error {
	// 检查是否已有数据
	if n, err := RDB.Exists(CTX, "feed").Result(); err != nil || n > 0 {
		return err
	}
	// 读取一定数量的视频流
	videoList, err := dal.GetFeed(latestTime, config.MaxFeedSizeRedis)
	if err != nil {
		return err
	}
	if len(videoList) == 0 {
		return nil
	}
	// 将视频 id 写入 feed，视频信息写入 hash，同时记录作者信息
	userIdList := make([]int64, len(videoList))
	videoIdList := make([]int64, len(videoList))
	for i, video := range videoList {
		if err := RDB.ZAdd(CTX, "feed", &redis.Z{Score: float64(video.CreateTime), Member: video.Id}).Err(); err != nil {
			return err
		}
		key := VideoKey(video.Id)
		if err := RedisStructHash(video, key); err != nil {
			return err
		}
		if err := RDB.Expire(CTX, key, config.RedisExp).Err(); err != nil {
			return err
		}
		userIdList[i] = video.UserId
		videoIdList[i] = video.Id
	}
	// 将视频流中的评论信息写入 Redis
	if err := WriteCommentFromFeed(videoIdList); err != nil {
		return err
	}
	// 将视频流中的作者信息写入 Redis
	if err := WriteUserFromFeed(userIdList); err != nil {
		return err
	}
	// 将视频流中作者的关注粉丝信息写入 Redis
	if err := WriteRelationFromFeed(userIdList); err != nil {
		return err
	}
	// 将视频流中作者的点赞信息写入 Redis
	err = WriteFavoriteFromFeed(userIdList)
	return err
}

// ReadVideo 先在 Redis 中查找视频信息，若无则从 MySQL 中读取
func ReadVideo(videoId int64) (dal.Video, error) {
	var video dal.Video
	key := VideoKey(videoId)
	n, err := RDB.Exists(CTX, key).Result()
	if err != nil {
		return dal.Video{}, err
	}
	if n <= 0 { // 没有此视频的缓存，从 MySQL 中读取
		video, err = dal.GetVideoById(videoId)
		if err != nil {
			return dal.Video{}, err
		}
		// 写入 Redis
		if err := RedisStructHash(video, key); err != nil {
			return dal.Video{}, err
		}
	} else { // 有此缓存
		video, err = ReadVideoFromHash(key)
		if err != nil {
			return dal.Video{}, err
		}
	}
	RDB.Expire(CTX, key, config.RedisExp)
	// 查询该视频对应的用户
	user, err := ReadUser(video.UserId)
	if err != nil {
		return dal.Video{}, err
	}
	video.Author = user
	return video, nil
}

// ReadFeed 从 Redis 中读取视频流，包括 id、视频信息、作者信息
// 没有的数据从 MySQL 中读取并写入 Redis
func ReadFeed(latestTime int64) ([]dal.Video, error) {
	// 读取视频流 id
	opt := redis.ZRangeBy{
		Min:    "0",
		Max:    strconv.FormatInt(latestTime, 10),
		Offset: 0,
		Count:  config.MaxFeedSize,
	}
	videoIdList, err := RDB.ZRevRangeByScore(CTX, "feed", &opt).Result()
	if err != nil {
		return []dal.Video{}, err
	}
	//if len(videoIdList) < config.MaxFeedSize { // 比较少见的场景，即用户请求的视频流 id 超出了缓存中的范围
	//	if !(len(videoIdList) != 0 && videoIdList[len(videoIdList)-1] == "0") { // 非视频总数不够的情况
	//	}
	//}
	videoList := make([]dal.Video, len(videoIdList))
	for i, videoIdStr := range videoIdList {
		videoId, err := strconv.ParseInt(videoIdStr, 10, 64)
		if err != nil {
			return []dal.Video{}, err
		}
		video, err := ReadVideo(videoId)
		if err != nil {
			return []dal.Video{}, err
		}
		videoList[i] = video
	}
	return videoList, nil
}

// AddVideo 将新发布的视频分别写入 Redis 的 feed 和视频 hash 中
// 由于是在 Publish 中调用，一定为新的视频，因此不用查询是否已存在
// 由于发布视频的用户一定是登录了的用户，因此不用重新向 Redis 中写入作者
func AddVideo(video dal.Video) error {
	if err := RDB.ZAdd(CTX, "feed", &redis.Z{Score: float64(video.CreateTime), Member: video.Id}).Err(); err != nil {
		return err
	}
	key := VideoKey(video.Id)
	if err := RedisStructHash(video, key); err != nil {
		return err
	}
	if err := RDB.Expire(CTX, key, config.RedisExp).Err(); err != nil {
		return err
	}
	return nil
}

// DeleteVideo 涉及到 FavoriteCount 和 CommentCount 变化时要删除视频
func DeleteVideo(videoId int64) error {
	key := VideoKey(videoId)
	err := RDB.Del(CTX, key).Err()
	return err
}