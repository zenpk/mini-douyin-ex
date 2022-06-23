package cache

import (
	"github.com/zenpk/mini-douyin-ex/config"
	"github.com/zenpk/mini-douyin-ex/dal"
	"strconv"
)

// WriteRelationFromFeed 将视频流相关用户的关注粉丝关系写入 Redis
func WriteRelationFromFeed(userIdList []int64) error {
	// Redis 中存储的 Relation 比较特殊，需要分成关注和粉丝两种 key
	for _, userId := range userIdList {
		if err := getAndWriteRelation(userId); err != nil {
			return err
		}
	}
	return nil
}

// getAndWriteRelation 从数据库中读取关注粉丝列表并写入
func getAndWriteRelation(userId int64) error {
	// 查找关注列表
	followList, err := dal.GetFollowList(userId)
	if err != nil {
		return err
	}
	// 写入 Redis
	for _, id := range followList {
		key := FollowKey(userId)
		if err := RDB.SAdd(CTX, key, id).Err(); err != nil {
			return err
		}
		RDB.Expire(CTX, key, config.RedisExp)
	}
	// 查找粉丝列表
	followerList, err := dal.GetFollowerList(userId)
	if err != nil {
		return err
	}
	// 写入 Redis
	for _, id := range followerList {
		key := FollowerKey(userId)
		if err := RDB.SAdd(CTX, key, id).Err(); err != nil {
			return err
		}
		RDB.Expire(CTX, key, config.RedisExp)
	}
	return nil
}

// ReadRelation 查找是否存在某条关注信息，不存在则从数据库写入
// userA 一般是当前登录用户，因此优先查询和写入
func ReadRelation(userAId, userBId int64) (bool, error) {
	key := FollowKey(userAId)
	n, err := RDB.Exists(CTX, key).Result()
	if err != nil {
		return false, err
	}
	// 未命中，先从数据库中提取 A 的关注粉丝记录
	if n <= 0 {
		if err := getAndWriteRelation(userAId); err != nil {
			return false, err
		}
	}
	// 再查询是否存在 A 关注 B 的记录
	isFollow, err := RDB.SIsMember(CTX, key, userBId).Result()
	if err != nil {
		return false, err
	} else {
		RDB.Expire(CTX, key, config.RedisExp)
		return isFollow, nil
	}
}

// ReadFollow 读取用户关注列表，并判断列表中用户是否被关注
func ReadFollow(userAId, userBId int64) ([]dal.User, error) {
	var followList []dal.User
	key := FollowKey(userBId) // 查看的是用户 B 的信息
	n, err := RDB.Exists(CTX, key).Result()
	if err != nil {
		return []dal.User{}, err
	}
	// 未命中，先从数据库中提取 B 的关注粉丝记录
	if n <= 0 {
		if err := getAndWriteRelation(userBId); err != nil {
			return []dal.User{}, err
		}
	}
	// 再查询所有关注用户的信息
	followIdStrList, err := RDB.SMembers(CTX, key).Result()
	if err != nil {
		return []dal.User{}, err
	}
	for _, followIdStr := range followIdStrList {
		followId, err := strconv.ParseInt(followIdStr, 10, 64)
		if err != nil {
			return []dal.User{}, err
		}
		user, err := ReadUser(followId)
		if err != nil {
			return []dal.User{}, err
		}
		// 查找当前登录用户（userA）是否关注了该用户
		user.IsFollow, err = ReadRelation(userAId, userBId)
		if err != nil {
			return []dal.User{}, err
		}
		followList = append(followList, user)
	}
	return followList, nil
}

// ReadFollower 读取用户粉丝列表，并判断列表中用户是否被关注
func ReadFollower(userAId, userBId int64) ([]dal.User, error) {
	var followList []dal.User
	key := FollowerKey(userBId) // 查看的是用户 B 的信息
	n, err := RDB.Exists(CTX, key).Result()
	if err != nil {
		return []dal.User{}, err
	}
	// 未命中，先从数据库中提取 B 的关注粉丝记录
	if n <= 0 {
		if err := getAndWriteRelation(userBId); err != nil {
			return []dal.User{}, err
		}
	}
	// 再查询所有粉丝的信息
	followIdStrList, err := RDB.SMembers(CTX, key).Result()
	if err != nil {
		return []dal.User{}, err
	}
	for _, followIdStr := range followIdStrList {
		followId, err := strconv.ParseInt(followIdStr, 10, 64)
		if err != nil {
			return []dal.User{}, err
		}
		user, err := ReadUser(followId)
		if err != nil {
			return []dal.User{}, err
		}
		// 查找当前登录用户（userA）是否关注了该用户
		user.IsFollow, err = ReadRelation(userAId, userBId)
		if err != nil {
			return []dal.User{}, err
		}
		followList = append(followList, user)
	}
	return followList, nil
}

// AddFollow 有新关注时，先写入 MySQL 再写入 Redis
// 需要删除 Redis 中涉及到的用户，采用延迟双删
// 和上面一样，把主体放在 userA 上
func AddFollow(userAId, userBId int64) error {
	// 第一次删除 Redis 中的用户
	if err := DeleteUser(userAId); err != nil {
		return err
	}
	if err := DeleteUser(userBId); err != nil {
		return err
	}
	// 写入 MySQL
	if err := dal.AddFollow(userAId, userBId); err != nil {
		return err
	}
	// 第二次删除 Redis 中的用户
	if err := DeleteUser(userAId); err != nil {
		return err
	}
	if err := DeleteUser(userBId); err != nil {
		return err
	}
	// 写入 Redis
	key := FollowKey(userAId)
	if err := RDB.SAdd(CTX, key, userBId).Err(); err != nil {
		return err
	}
	if err := RDB.Expire(CTX, key, config.RedisExp).Err(); err != nil {
		return err
	}
	return nil
}

// DeleteFollow 删除关注时，采用延迟双删确保一致性
func DeleteFollow(userAId, userBId int64) error {
	// Redis 第一次删除关注
	keyA := FollowKey(userAId)
	if err := RDB.SRem(CTX, keyA, userBId).Err(); err != nil {
		return err
	}
	keyB := FollowerKey(userBId)
	if err := RDB.SRem(CTX, keyB, userAId).Err(); err != nil {
		return err
	}
	// Redis 第一次删除用户
	if err := DeleteUser(userAId); err != nil {
		return err
	}
	if err := DeleteUser(userBId); err != nil {
		return err
	}
	// MySQL 删除
	if err := dal.DeleteFollow(userAId, userBId); err != nil {
		return err
	}
	// Redis 第二次删除用户
	if err := DeleteUser(userAId); err != nil {
		return err
	}
	if err := DeleteUser(userBId); err != nil {
		return err
	}
	// Redis 第二次删除关注
	if err := RDB.SRem(CTX, keyA, userBId).Err(); err != nil {
		return err
	}
	if err := RDB.SRem(CTX, keyB, userAId).Err(); err != nil {
		return err
	}
	return nil
}
