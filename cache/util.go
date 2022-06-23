package cache

import (
	"github.com/zenpk/mini-douyin-ex/dal"
	"reflect"
	"strconv"
)

const (
	TagName = "redistructhash"
	NoHash  = "no"
)

// convertCase - from CamelCase to camel_case
func convertCase(in string) string {
	var out []rune
	for i, c := range in {
		if c >= 'A' && c <= 'Z' {
			c += 32
			if i > 0 {
				out = append(out, '_')
			}
		}
		out = append(out, c)
	}
	return string(out)
}

// RedisStructHash - Automatically create hash from struct
func RedisStructHash(t interface{}, key string) error {
	ref := reflect.ValueOf(t)
	for i := 0; i < ref.NumField(); i++ {
		tag := ref.Type().Field(i).Tag.Get(TagName)
		if tag == NoHash {
			continue
		}
		fieldName := ref.Type().Field(i).Name
		dbFieldName := convertCase(fieldName)
		if err := RDB.HSet(CTX, key, dbFieldName, ref.Field(i).Interface()).Err(); err != nil {
			return err
		}
	}
	return nil
}

func hGetInt64(key, field string) (int64, error) {
	str, err := RDB.HGet(CTX, key, field).Result()
	if err != nil {
		return 0, err
	}
	id, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ReadVideoFromHash 从 Redis 的 hash 中读取视频信息
func ReadVideoFromHash(key string) (dal.Video, error) {
	var video dal.Video
	var err error

	video.Id, err = hGetInt64(key, "id")
	if err != nil {
		return dal.Video{}, err
	}
	video.UserId, err = hGetInt64(key, "user_id")
	if err != nil {
		return dal.Video{}, err
	}
	video.PlayUrl, err = RDB.HGet(CTX, key, "play_url").Result()
	if err != nil {
		return dal.Video{}, err
	}
	video.CoverUrl, err = RDB.HGet(CTX, key, "cover_url").Result()
	if err != nil {
		return dal.Video{}, err
	}
	video.FavoriteCount, err = hGetInt64(key, "favorite_count")
	if err != nil {
		return dal.Video{}, err
	}
	video.CommentCount, err = hGetInt64(key, "comment_count")
	if err != nil {
		return dal.Video{}, err
	}
	video.Title, err = RDB.HGet(CTX, key, "title").Result()
	if err != nil {
		return dal.Video{}, err
	}
	video.CreateTime, err = hGetInt64(key, "create_time")
	if err != nil {
		return dal.Video{}, err
	}
	return video, nil
}

// ReadUserFromHash 从 Redis 的 hash 中读取用户信息
func ReadUserFromHash(key string) (dal.User, error) {
	var user dal.User
	var err error

	user.Id, err = hGetInt64(key, "id")
	if err != nil {
		return dal.User{}, err
	}
	user.Name, err = RDB.HGet(CTX, key, "name").Result()
	if err != nil {
		return dal.User{}, err
	}
	user.FollowCount, err = hGetInt64(key, "follow_count")
	if err != nil {
		return dal.User{}, err
	}
	user.FollowerCount, err = hGetInt64(key, "follower_count")
	if err != nil {
		return dal.User{}, err
	}
	return user, nil
}

// ReadCommentFromHash 从 Redis 的 hash 中读取评论信息
func ReadCommentFromHash(key string) (dal.Comment, error) {
	var comment dal.Comment
	var err error

	comment.Id, err = hGetInt64(key, "id")
	if err != nil {
		return dal.Comment{}, err
	}
	comment.UserId, err = hGetInt64(key, "user_id")
	if err != nil {
		return dal.Comment{}, err
	}
	comment.VideoId, err = hGetInt64(key, "video_id")
	if err != nil {
		return dal.Comment{}, err
	}
	comment.Content, err = RDB.HGet(CTX, key, "content").Result()
	if err != nil {
		return dal.Comment{}, err
	}
	comment.CreateDate, err = RDB.HGet(CTX, key, "create_date").Result()
	if err != nil {
		return dal.Comment{}, err
	}
	return comment, nil
}

func UserKey(userId int64) string {
	return "user:" + strconv.FormatInt(userId, 10)
}

func VideoKey(videoId int64) string {
	return "video:" + strconv.FormatInt(videoId, 10)
}

func CommentListKey(videoId int64) string {
	return "comment_list:" + strconv.FormatInt(videoId, 10)
}

func CommentKey(commentId int64) string {
	return "comment:" + strconv.FormatInt(commentId, 10)
}

func FollowKey(userId int64) string {
	return "follow:" + strconv.FormatInt(userId, 10)
}

func FollowerKey(userId int64) string {
	return "follower:" + strconv.FormatInt(userId, 10)
}

func FavoriteKey(userId, videoId int64) string {
	return "favorite:" + strconv.FormatInt(userId, 10) + ":" + strconv.FormatInt(videoId, 10)
}
