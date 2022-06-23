package cache

import (
	"github.com/go-redis/redis/v8"
	"github.com/zenpk/mini-douyin-ex/config"
	"github.com/zenpk/mini-douyin-ex/dal"
	"strconv"
)

// WriteFeed 将视频流 **首次** 写入 Redis
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
	for _, video := range videoList {
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
		// 读取视频对应评论并写入 Redis
		if _, err := WriteCommentList(video.Id); err != nil {
			return err
		}
		// 读取视频作者信息并写入 Redis
		if _, err := WriteUser(video.UserId); err != nil {
			return err
		}
		// 读取视频作者的关注粉丝信息并写入 Redis
		if err := WriteRelation(video.UserId); err != nil {
			return err
		}
		// 读取视频作者的投稿信息并写入 Redis
		if _, err := WritePublishList(video.UserId); err != nil {
			return err
		}
		// 读取视频作者的点赞信息并写入 Redis
		if _, err := WriteFavoriteList(video.UserId); err != nil {
			return err
		}
	}
	return nil
}

// WritePublishList 根据用户 id 从 MySQL 中读取投稿信息
// 根据用户 id 建立 set
func WritePublishList(userId int64) ([]dal.Video, error) {
	// 数据库读取投稿信息
	videoList, err := dal.GetPublishList(userId)
	if err != nil {
		return []dal.Video{}, err
	}
	// listKey 值是 userId 决定的，这样才能方便地查询每个用户的投稿视频
	listKey := PublishListKey(userId)
	for _, video := range videoList {
		if err := RDB.SAdd(CTX, listKey, video.Id).Err(); err != nil {
			return []dal.Video{}, err
		}
		// 同时还需要将每个 video 单独存储在 hash 中
		key := VideoKey(video.Id)
		if err := RedisStructHash(video, key); err != nil {
			return []dal.Video{}, err
		}
		if err := RDB.Expire(CTX, key, config.RedisExp).Err(); err != nil {
			return []dal.Video{}, err
		}
	}
	// set 整体设置一次过期时间即可
	if err := RDB.Expire(CTX, listKey, config.RedisExp).Err(); err != nil {
		return []dal.Video{}, err
	}

	return videoList, nil
}

// WriteVideo 从 MySQL 中读取视频信息写入 Redis
func WriteVideo(videoId int64) (dal.Video, error) {
	key := VideoKey(videoId)
	video, err := dal.GetVideoById(videoId)
	if err != nil {
		return dal.Video{}, err
	}
	// 写入 Redis
	if err := RedisStructHash(video, key); err != nil {
		return dal.Video{}, err
	}
	if err := RDB.Expire(CTX, key, config.RedisExp).Err(); err != nil {
		return dal.Video{}, err
	}
	return video, nil
}

// ReadVideo 先在 Redis 中查找视频信息，若无则从 MySQL 中读取
func ReadVideo(videoId int64) (dal.Video, error) {
	var video dal.Video
	key := VideoKey(videoId)
	n, err := RDB.Exists(CTX, key).Result()
	if err != nil {
		return dal.Video{}, err
	}
	if n <= 0 { // 没有此视频的缓存，从 MySQL 中读取并写入
		video, err = WriteVideo(videoId)
		if err != nil {
			return dal.Video{}, err
		}
	} else { // 有此缓存
		video, err = ReadVideoFromHash(key)
		if err != nil {
			return dal.Video{}, err
		}
		// 更新过期时间
		if err := RDB.Expire(CTX, key, config.RedisExp).Err(); err != nil {
			return dal.Video{}, err
		}
	}
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
	if err := RDB.Expire(CTX, "feed", config.RedisExp).Err(); err != nil {
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

// ReadPublishList 读取用户投稿视频
// userA 是当前登录用户，userB 是查看的用户
func ReadPublishList(userAId, userBId int64) ([]dal.Video, error) {
	key := PublishListKey(userBId)
	n, err := RDB.Exists(CTX, key).Result()
	if err != nil {
		return []dal.Video{}, err
	}
	var videoList []dal.Video
	if n <= 0 { // 未命中，先从数据库中提取用户的投稿记录并写入
		videoList, err = WritePublishList(userBId)
		if err != nil {
			return []dal.Video{}, err
		}
		for i, video := range videoList {
			// 查找当前登录用户是否点过赞
			videoList[i].IsFavorite, err = ReadFavorite(userAId, video.Id)
			if err != nil {
				return []dal.Video{}, err
			}
			// 查找视频对应的用户
			videoList[i].Author, err = ReadUser(videoList[i].UserId)
			if err != nil {
				return []dal.Video{}, err
			}
			// 查找是否关注了这个用户
			videoList[i].Author.IsFollow, err = ReadRelation(userAId, videoList[i].UserId)
			if err != nil {
				return []dal.Video{}, err
			}
		}
	} else { // 命中
		videoIdStrList, err := RDB.SMembers(CTX, key).Result()
		if err != nil {
			return []dal.Video{}, err
		}
		// 更新过期时间
		if err := RDB.Expire(CTX, key, config.RedisExp).Err(); err != nil {
			return []dal.Video{}, err
		}
		for _, videoIdStr := range videoIdStrList {
			videoId, err := strconv.ParseInt(videoIdStr, 10, 64)
			if err != nil {
				return []dal.Video{}, err
			}
			// 根据 id 查找视频，先查 Redis 再查 MySQL
			video, err := ReadVideo(videoId)
			if err != nil {
				return []dal.Video{}, err
			}
			// 查找当前登录用户是否点过赞
			video.IsFavorite, err = ReadFavorite(userAId, video.Id)
			if err != nil {
				return []dal.Video{}, err
			}
			// 查找视频对应的用户
			video.Author, err = ReadUser(video.UserId)
			if err != nil {
				return []dal.Video{}, err
			}
			// 查找是否关注了这个用户
			video.Author.IsFollow, err = ReadRelation(userAId, video.UserId)
			if err != nil {
				return []dal.Video{}, err
			}
			videoList = append(videoList, video)
		}
	}
	return videoList, nil
}

// AddVideo 将新发布的视频分别写入 Redis 的 feed 和视频 hash 中
// 同时还需要写入用户的投稿列表中
// 由于发布视频的用户一定是登录了的用户，因此不用重新向 Redis 中写入作者
func AddVideo(video dal.Video) error {
	// 写入 feed
	if err := RDB.ZAdd(CTX, "feed", &redis.Z{Score: float64(video.CreateTime), Member: video.Id}).Err(); err != nil {
		return err
	}
	if err := RDB.Expire(CTX, "feed", config.RedisExp).Err(); err != nil {
		return err
	}
	// 写入 hash
	key := VideoKey(video.Id)
	if err := RedisStructHash(video, key); err != nil {
		return err
	}
	if err := RDB.Expire(CTX, key, config.RedisExp).Err(); err != nil {
		return err
	}
	// 写入 set
	listKey := PublishListKey(video.UserId)
	if err := RDB.SAdd(CTX, listKey, video.Id).Err(); err != nil {
		return err
	}
	if err := RDB.Expire(CTX, listKey, config.RedisExp).Err(); err != nil {
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
