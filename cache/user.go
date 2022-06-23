package cache

import (
	"github.com/zenpk/mini-douyin-ex/config"
	"github.com/zenpk/mini-douyin-ex/dal"
)

// RegisterLoginUser 用户登录时写入缓存
func RegisterLoginUser(user dal.User) error {
	key := UserKey(user.Id)
	if err := RedisStructHash(user, key); err != nil {
		return err
	}
	if err := RDB.Expire(CTX, key, config.RedisExp).Err(); err != nil {
		return err
	}
	return nil
}

// WriteUser 从 MySQL 中读取用户并写入 Redis
func WriteUser(userId int64) (dal.User, error) {
	key := UserKey(userId)
	user, err := dal.GetUserById(userId)
	if err != nil {
		return dal.User{}, err
	}
	// 写入 Redis
	if err := RedisStructHash(user, key); err != nil {
		return dal.User{}, err
	}
	return user, nil
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
		user, err = WriteUser(userId)
		if err != nil {
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
