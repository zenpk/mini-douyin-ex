package cache

import (
	"github.com/zenpk/mini-douyin-ex/config"
	"github.com/zenpk/mini-douyin-ex/dal"
	"strconv"
)

func WriteCommentFromFeed(videoIdList []int64) error {
	// 数据库读取评论信息
	commentList, err := dal.GetCommentByVideoIdList(videoIdList)
	if err != nil {
		return err
	}
	// 写入 Redis
	for _, comment := range commentList {
		key := CommentKey(comment.Id)
		if err := RedisStructHash(comment, key); err != nil {
			return err
		}
		if err := RDB.Expire(CTX, key, config.RedisExp).Err(); err != nil {
			return err
		}
	}
	return nil
}

// ReadComment 读取评论，没有则从数据库写入
// 还需要同时读取用户信息
func ReadComment(videoId int64) ([]dal.Comment, error) {
	var commentList []dal.Comment
	// 用正则表达式获取全部评论的 key
	commentKeyList, err := RDB.Keys(CTX, "comment:*").Result()
	if err != nil {
		return []dal.Comment{}, err
	}
	for _, commentKey := range commentKeyList {
		videoIdHashStr, err := RDB.HGet(CTX, commentKey, "video_id").Result()
		if err != nil {
			return []dal.Comment{}, err
		}
		videoIdHash, err := strconv.ParseInt(videoIdHashStr, 10, 64)
		if err != nil {
			return []dal.Comment{}, err
		}
		if videoIdHash == videoId { // 找到了该视频的评论
			comment, err := ReadCommentFromHash(commentKey)
			if err != nil {
				return []dal.Comment{}, err
			}
			commentList = append(commentList, comment)
		}
	}
	// 获取评论的操作一定是在视频读取之后，因此可以很快速地获取到视频相关信息
	video, err := ReadVideo(videoId)
	if err != nil {
		return []dal.Comment{}, err
	}
	if int64(len(commentList)) < video.CommentCount { // 评论数不够，由于无法判断少了哪条评论，因此只能重新从数据库中获取
		commentList, err = dal.GetCommentByVideoId(videoId)
		if err != nil {
			return []dal.Comment{}, err
		}
		// 将未放入缓存的评论写入缓存
		for _, comment := range commentList {
			key := CommentKey(comment.Id)
			n, err := RDB.Exists(CTX, key).Result()
			if err != nil {
				return []dal.Comment{}, err
			}
			if n <= 0 { // 没有找到记录，写入缓存
				if err := RedisStructHash(comment, key); err != nil {
					return []dal.Comment{}, err
				}
			}
		}
	}
	// 对每条评论读取用户信息
	for _, comment := range commentList {
		user, err := ReadUser(comment.UserId)
		if err != nil {
			return []dal.Comment{}, err
		}
		comment.User = user
	}
	return commentList, nil
}

// AddComment 有新评论时，先写入 MySQL 再写入 Redis
// 需要删除 Redis 中涉及到的视频，采用延迟双删
func AddComment(comment dal.Comment) error {
	// 第一次删除 Redis 中的视频
	if err := DeleteVideo(comment.VideoId); err != nil {
		return err
	}
	// 写入 MySQL
	commentId, err := dal.AddComment(comment)
	if err != nil {
		return err
	}
	// 第二次删除 Redis 中的视频
	if err := DeleteVideo(comment.VideoId); err != nil {
		return err
	}
	// 写入 Redis
	key := CommentKey(commentId)
	comment.Id = commentId
	if err := RedisStructHash(comment, key); err != nil {
		return err
	}
	if err := RDB.Expire(CTX, key, config.RedisExp).Err(); err != nil {
		return err
	}
	return nil
}

// DeleteComment 删除评论时，采用延迟双删确保一致性
func DeleteComment(userId, videoId, commentId int64) error {
	// Redis 第一次删除评论
	key := CommentKey(commentId)
	if err := RDB.Del(CTX, key).Err(); err != nil {
		return err
	}
	// Redis 第一次删除视频
	if err := DeleteVideo(videoId); err != nil {
		return err
	}
	// MySQL 删除
	if err := dal.DeleteComment(userId, videoId, commentId); err != nil {
		return err
	}
	// Redis 第一次删除视频
	if err := DeleteVideo(videoId); err != nil {
		return err
	}
	// Redis 第二次删除评论
	if err := RDB.Del(CTX, key).Err(); err != nil {
		return err
	}
	return nil
}
