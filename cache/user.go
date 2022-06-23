package cache

import (
	"github.com/zenpk/mini-douyin-ex/config"
	"github.com/zenpk/mini-douyin-ex/dal"
)

// WriteUserFromFeed 根据视频流中的作者 id 将作者信息写入 Redis
func WriteUserFromFeed(userIdList []int64) error {
	// 数据库读取用户信息
	userList, err := dal.GetUserListById(userIdList)
	if err != nil {
		return err
	}
	// 写入 Redis
	for _, user := range userList {
		key := UserKey(user.Id)
		if err := RedisStructHash(user, key); err != nil {
			return err
		}
		if err := RDB.Expire(CTX, key, config.RedisExp).Err(); err != nil {
			return err
		}
	}
	return nil
}

// WriteUser 用户登录时写入缓存
func WriteUser(user dal.User) error {
	key := UserKey(user.Id)
	// 将用户信息写入 Redis 前需要先查询是否已存在
	if n, err := RDB.Exists(CTX, key).Result(); n > 0 || err != nil {
		return err
	}
	if err := RedisStructHash(user, key); err != nil {
		return err
	}
	if err := RDB.Expire(CTX, key, config.RedisExp).Err(); err != nil {
		return err
	}
	return nil
}

// ReadUser 从 Redis 中查找用户，不存在则读 MySQL 写入
func ReadUser(userId int64) (dal.User, error) {
	var user dal.User
	key := UserKey(userId)
	n, err := RDB.Exists(CTX, key).Result()
	if err != nil {
		return dal.User{}, err
	}
	if n <= 0 { // 未命中，读取 MySQL
		user, err = dal.GetUserById(userId)
		if err != nil {
			return dal.User{}, err
		}
		// 写入 Redis
		if err := RedisStructHash(user, key); err != nil {
			return dal.User{}, err
		}
	} else { // 命中，直接读取
		user, err = ReadUserFromHash(key)
		if err != nil {
			return dal.User{}, err
		}
	}
	RDB.Expire(CTX, key, config.RedisExp)
	return user, nil
}

// DeleteUser 涉及到 FollowCount 和 FollowerCount 变化时要删除用户
func DeleteUser(userId int64) error {
	key := UserKey(userId)
	err := RDB.Del(CTX, key).Err()
	return err
}
