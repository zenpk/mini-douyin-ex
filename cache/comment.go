package cache

import (
	"github.com/zenpk/mini-douyin-ex/config"
	"github.com/zenpk/mini-douyin-ex/dal"
	"strconv"
)

// WriteCommentList 根据视频 id 从 MySQL 中读取对应的评论列表，用 set 存储
// Comment 本身的内容用 hash 存储
func WriteCommentList(videoId int64) ([]dal.Comment, error) {
	commentList, err := dal.GetCommentByVideoId(videoId)
	if err != nil {
		return []dal.Comment{}, err
	}
	listKey := CommentListKey(videoId)
	for _, comment := range commentList {
		// 写入 set
		if err := RDB.SAdd(CTX, listKey, comment.Id).Err(); err != nil {
			return []dal.Comment{}, err
		}
		// 写入 hash
		key := CommentKey(comment.Id)
		if err := RedisStructHash(comment, key); err != nil {
			return []dal.Comment{}, err
		}
		if err := RDB.Expire(CTX, key, config.RedisExp).Err(); err != nil {
			return []dal.Comment{}, err
		}
	}
	// set 整体设置一次过期时间即可
	if err := RDB.Expire(CTX, listKey, config.RedisExp).Err(); err != nil {
		return []dal.Comment{}, err
	}
	return commentList, nil
}

// ReadCommentList 读取视频的评论列表，没有则从数据库写入
// 需要同时读取用户信息
func ReadCommentList(videoId int64) ([]dal.Comment, error) {
	var commentList []dal.Comment
	listKey := CommentListKey(videoId)
	n, err := RDB.Exists(CTX, listKey).Result()
	if err != nil {
		return []dal.Comment{}, err
	}
	if n <= 0 { // 未命中，从数据库中读取并分别写入 set 和 hash
		commentList, err = WriteCommentList(videoId)
		if err != nil {
			return []dal.Comment{}, err
		}
	} else { // 命中，从 Redis 中读取
		commentIdStrList, err := RDB.SMembers(CTX, listKey).Result()
		if err != nil {
			return []dal.Comment{}, err
		}
		// 更新过期时间
		if err := RDB.Expire(CTX, listKey, config.RedisExp).Err(); err != nil {
			return []dal.Comment{}, err
		}
		for _, commentIdStr := range commentIdStrList {
			commentId, err := strconv.ParseInt(commentIdStr, 10, 64)
			if err != nil {
				return []dal.Comment{}, err
			}
			// 查找对应评论，若无则写入评论
			key := CommentKey(commentId)
			n, err := RDB.Exists(CTX, key).Result()
			if err != nil {
				return []dal.Comment{}, err
			}
			var comment dal.Comment
			if n <= 0 { // 未命中，从数据库中读取
				comment, err = dal.GetCommentById(commentId)
				if err != nil {
					return []dal.Comment{}, err
				}
			} else { // 命中
				comment, err = ReadCommentFromHash(key)
				if err != nil {
					return []dal.Comment{}, err
				}
				if err := RDB.Expire(CTX, key, config.RedisExp).Err(); err != nil {
					return []dal.Comment{}, err
				}
			}
			// 为每条评论读取用户信息
			comment.User, err = ReadUser(comment.UserId)
			if err != nil {
				return []dal.Comment{}, err
			}
			commentList = append(commentList, comment)
		}
	}
	return commentList, nil
}

// AddComment 有新评论时，先写入 MySQL 再写入 Redis
// 需要删除 Redis 中涉及到的视频，采用延迟双删
func AddComment(comment dal.Comment) error {
	// Redis 第一次删除视频
	if err := DeleteVideo(comment.VideoId); err != nil {
		return err
	}
	// 写入 MySQL
	commentId, err := dal.AddComment(comment)
	if err != nil {
		return err
	}
	// 更新生成的自增主键
	comment.Id = commentId
	// Redis 第二次删除视频
	if err := DeleteVideo(comment.VideoId); err != nil {
		return err
	}
	// 分别写入 Redis 的 set 和 hash
	// 写入 set
	listKey := CommentListKey(comment.VideoId)
	if err := RDB.SAdd(CTX, listKey, commentId).Err(); err != nil {
		return err
	}
	// 写入 Hash
	key := CommentKey(commentId)
	if err := RedisStructHash(comment, key); err != nil {
		return err
	}
	if err := RDB.Expire(CTX, key, config.RedisExp).Err(); err != nil {
		return err
	}
	return nil
}

// DeleteComment 删除评论时，采用延迟双删确保一致性
// 此处 Redis 需要删除的有：该条评论的 hash、该条评论对应的 set 中的 id、该条评论对应视频的 hash
func DeleteComment(userId, videoId, commentId int64) error {
	// Redis 第一次删除评论 hash
	key := CommentKey(commentId)
	if err := RDB.Del(CTX, key).Err(); err != nil {
		return err
	}
	// Redis 第一次删除评论 set
	listKey := CommentListKey(videoId)
	if err := RDB.SRem(CTX, listKey, commentId).Err(); err != nil {
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
	// Redis 第二次删除视频
	if err := DeleteVideo(videoId); err != nil {
		return err
	}
	// Redis 第二次删除评论 set
	if err := RDB.SRem(CTX, listKey, commentId).Err(); err != nil {
		return err
	}
	// Redis 第二次删除评论
	if err := RDB.Del(CTX, key).Err(); err != nil {
		return err
	}
	return nil
}

// Old version
//func ReadComment(videoId int64) ([]dal.Comment, error) {
//	var commentList []dal.Comment
//	// 用正则表达式获取全部评论的 key
//	commentKeyList, err := RDB.Keys(CTX, "comment:*").Result()
//	if err != nil {
//		return []dal.Comment{}, err
//	}
//	for _, commentKey := range commentKeyList {
//		videoIdHashStr, err := RDB.HGet(CTX, commentKey, "video_id").Result()
//		if err != nil {
//			return []dal.Comment{}, err
//		}
//		videoIdHash, err := strconv.ParseInt(videoIdHashStr, 10, 64)
//		if err != nil {
//			return []dal.Comment{}, err
//		}
//		if videoIdHash == videoId { // 找到了该视频的评论
//			comment, err := ReadCommentFromHash(commentKey)
//			if err != nil {
//				return []dal.Comment{}, err
//			}
//			commentList = append(commentList, comment)
//		}
//	}
//	// 获取评论的操作一定是在视频读取之后，因此可以很快速地获取到视频相关信息
//	video, err := ReadVideo(videoId)
//	if err != nil {
//		return []dal.Comment{}, err
//	}
//	if int64(len(commentList)) < video.CommentCount { // 评论数不够，由于无法判断少了哪条评论，因此只能重新从数据库中获取
//		commentList, err = dal.GetCommentByVideoId(videoId)
//		if err != nil {
//			return []dal.Comment{}, err
//		}
//		// 将未放入缓存的评论写入缓存
//		for _, comment := range commentList {
//			key := CommentKey(comment.Id)
//			n, err := RDB.Exists(CTX, key).Result()
//			if err != nil {
//				return []dal.Comment{}, err
//			}
//			if n <= 0 { // 没有找到记录，写入缓存
//				if err := RedisStructHash(comment, key); err != nil {
//					return []dal.Comment{}, err
//				}
//			}
//		}
//	}
//	// 对每条评论读取用户信息
//	for _, comment := range commentList {
//		user, err := ReadUser(comment.UserId)
//		if err != nil {
//			return []dal.Comment{}, err
//		}
//		comment.User = user
//	}
//	return commentList, nil
//}
