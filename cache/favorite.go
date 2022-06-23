package cache

import (
	"github.com/zenpk/mini-douyin-ex/config"
	"github.com/zenpk/mini-douyin-ex/dal"
)

// WriteFavoriteFromFeed 从视频流的用户 id 中读取点赞信息
// 点赞信息比较特殊，用 favorite:userId:videoId 表示，value 无意义
func WriteFavoriteFromFeed(userIdList []int64) error {
	// 数据库读取点赞信息
	favoriteList, err := dal.GetFavoriteListByUserIdList(userIdList)
	if err != nil {
		return err
	}
	// 写入 Redis
	for _, favorite := range favoriteList {
		key := FavoriteKey(favorite.UserId, favorite.VideoId)
		if err := RDB.Set(CTX, key, 1, config.RedisExp).Err(); err != nil {
			return err
		}
	}
	return nil
}

// getAndWriteFavorite 从 MySQL 中读取用户的点赞记录并写入 Redis
func getAndWriteFavorite(userId int64) error {
	// 查找点赞记录
	favoriteList, err := dal.GetFavoriteListByUserId(userId)
	if err != nil {
		return err
	}
	// 写入 Redis
	for _, favorite := range favoriteList {
		key := FavoriteKey(favorite.UserId, favorite.VideoId)
		if err := RDB.Set(CTX, key, 1, config.RedisExp).Err(); err != nil {
			return err
		}
	}
	return nil
}

// ReadFavorite 查询用户是否点过赞，未命中则从 MySQL 中读取
func ReadFavorite(userId, videoId int64) (bool, error) {
	key := FavoriteKey(userId, videoId)
	n, err := RDB.Exists(CTX, key).Result()
	if err != nil {
		return false, err
	}
	// 未命中，先从数据库中提取用户的点赞记录
	if n <= 0 {
		if err := getAndWriteRelation(userId); err != nil {
			return false, err
		}
	}
	// 再查询是否存在点赞记录
	n, err = RDB.Exists(CTX, key).Result()
	if err != nil {
		return false, err
	} else {
		return n > 0, nil
	}
}

// AddFavorite 有新点赞时，先写入 MySQL 再写入 Redis
// 需要删除 Redis 中涉及到的视频，采用延迟双删
func AddFavorite(userId, videoId int64) error {
	// 第一次删除 Redis 中的视频
	if err := DeleteVideo(videoId); err != nil {
		return err
	}
	// 写入 MySQL
	if err := dal.AddFavorite(userId, videoId); err != nil {
		return err
	}
	// 第二次删除 Redis 中的视频
	if err := DeleteVideo(videoId); err != nil {
		return err
	}
	// 写入 Redis
	key := FavoriteKey(userId, videoId)
	if err := RDB.Set(CTX, key, 1, config.RedisExp).Err(); err != nil {
		return err
	}
	return nil
}

// DeleteFavorite 取消点赞时，采用延迟双删确保一致性
func DeleteFavorite(userId, videoId int64) error {
	// Redis 第一次删除点赞
	key := FavoriteKey(userId, videoId)
	if err := RDB.Del(CTX, key).Err(); err != nil {
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
	// Redis 第一次删除视频
	if err := DeleteVideo(videoId); err != nil {
		return err
	}
	// Redis 第二次删除点赞
	if err := RDB.Del(CTX, key).Err(); err != nil {
		return err
	}
	return nil
}
