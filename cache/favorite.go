package cache

import (
	"github.com/zenpk/mini-douyin-ex/config"
	"github.com/zenpk/mini-douyin-ex/dal"
	"strconv"
)

// WriteFavoriteList 根据用户 id 从 MySQL 中读取点赞信息
// 根据用户 id 建立 set
func WriteFavoriteList(userId int64) ([]dal.Favorite, error) {
	favoriteList, err := dal.GetFavoriteByUserId(userId)
	if err != nil {
		return []dal.Favorite{}, err
	}
	key := FavoriteKey(userId)
	for _, favorite := range favoriteList {
		if err := RDB.SAdd(CTX, key, favorite.VideoId).Err(); err != nil {
			return []dal.Favorite{}, err
		}
	}
	// 整体设置一次过期时间
	if err := RDB.Expire(CTX, key, config.RedisExp).Err(); err != nil {
		return []dal.Favorite{}, err
	}
	return favoriteList, err
}

// ReadFavorite 查询用户是否点过赞，未命中则从 MySQL 中读取
func ReadFavorite(userId, videoId int64) (bool, error) {
	key := FavoriteKey(userId)
	n, err := RDB.Exists(CTX, key).Result()
	if err != nil {
		return false, err
	}
	// 未命中，先从数据库中提取用户的点赞记录并写入
	if n <= 0 {
		if _, err := WriteFavoriteList(userId); err != nil {
			return false, err
		}
	}
	// 在用户点赞 set 中查询是否点赞
	videoIdStr := strconv.FormatInt(videoId, 10)
	isFavorite, err := RDB.SIsMember(CTX, key, videoIdStr).Result()
	if err != nil {
		return false, err
	}
	if err := RDB.Expire(CTX, key, config.RedisExp).Err(); err != nil {
		return false, err
	}
	return isFavorite, nil
}

// ReadFavoriteList 查询用户点赞视频列表，未命中则从 MySQL 中读取
// userA 是当前登录用户 userB 是查询用户
func ReadFavoriteList(userAId, userBId int64) ([]dal.Video, error) {
	key := FavoriteKey(userBId)
	n, err := RDB.Exists(CTX, key).Result()
	if err != nil {
		return []dal.Video{}, err
	}
	var videoList []dal.Video
	// 未命中，先从数据库中提取用户的点赞记录并写入
	if n <= 0 {
		favoriteList, err := WriteFavoriteList(userBId)
		if err != nil {
			return []dal.Video{}, err
		}
		for _, favorite := range favoriteList {
			// 根据 id 查找视频，先查 Redis 再查 MySQL
			video, err := ReadVideo(favorite.VideoId)
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

// AddFavorite 有新点赞时，先写入 MySQL 再写入 Redis
// 需要删除 Redis 中涉及到的视频，采用延迟双删
func AddFavorite(userId, videoId int64) error {
	// Redis 第一次删除视频
	if err := DeleteVideo(videoId); err != nil {
		return err
	}
	// 写入 MySQL
	if err := dal.AddFavorite(userId, videoId); err != nil {
		return err
	}
	// Redis 第二次删除视频
	if err := DeleteVideo(videoId); err != nil {
		return err
	}
	// 写入 Redis
	key := FavoriteKey(userId)
	if err := RDB.SAdd(CTX, key, videoId).Err(); err != nil {
		return err
	}
	if err := RDB.Expire(CTX, key, config.RedisExp).Err(); err != nil {
		return err
	}
	return nil
}

// DeleteFavorite 取消点赞时，采用延迟双删确保一致性
func DeleteFavorite(userId, videoId int64) error {
	// Redis 第一次删除点赞
	key := FavoriteKey(userId)
	if err := RDB.SRem(CTX, key, videoId).Err(); err != nil {
		return err
	}
	// Redis 第一次删除视频
	if err := DeleteVideo(videoId); err != nil {
		return err
	}
	// MySQL 删除
	if err := dal.DeleteFavorite(userId, videoId); err != nil {
		return err
	}
	// Redis 第二次删除视频
	if err := DeleteVideo(videoId); err != nil {
		return err
	}
	// Redis 第二次删除点赞
	if err := RDB.SRem(CTX, key, videoId).Err(); err != nil {
		return err
	}
	return nil
}
