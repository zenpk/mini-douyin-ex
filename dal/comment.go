package dal

import (
	"errors"
	"gorm.io/gorm"
)

type Comment struct {
	Id         int64  `json:"id" gorm:"primaryKey"`
	User       User   `json:"user" gorm:"-:all" redistructhash:"no"` // 不使用外键
	UserId     int64  `gorm:"not null"`
	VideoId    int64  `gorm:"not null"`
	Content    string `json:"content" gorm:"not null"`
	CreateDate string `json:"create_date" gorm:"not null"`
}

// AddComment 发表评论，返回生成的自增 id
func AddComment(comment Comment) (int64, error) {
	// 开启数据库事务，在 comments 中添加记录，在 videos 中更改评论数目
	if err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&comment).Error; err != nil {
			return err
		}
		if err := tx.Model(&Video{}).Where("id = ?", comment.VideoId).UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).Error; err != nil {
			return err
		}
		// 返回 nil 提交事务
		return nil
	}); err != nil {
		return 0, err
	}
	return comment.Id, nil
}

// DeleteComment 删除评论
func DeleteComment(userId, videoId, commentId int64) error {
	// 检查是否有该用户对应的评论
	if DB.Find(&Comment{}).Where("id = ?", commentId).Where("user_id = ?", userId).RowsAffected == 0 {
		return errors.New("无法删除评论")
	}
	// 可选删除方案：1. 直接在数据库中删除; 2. 软删除：comments 中设置一个 deleted 列，用 bool 表示是否删除
	// 目前实现的是第一种
	// 开启数据库事务
	if err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", commentId).Delete(&Comment{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&Video{}).Where("id = ?", videoId).UpdateColumn("comment_count", gorm.Expr("comment_count - ?", 1)).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func GetCommentByVideoId(videoId int64) ([]Comment, error) {
	var commentList []Comment
	err := DB.Where("video_id = ?", videoId).Find(&commentList).Error
	return commentList, err
}

func GetCommentByVideoIdList(videoIdList []int64) ([]Comment, error) {
	var commentList []Comment
	err := DB.Where("video_id IN ?", videoIdList).Find(&commentList).Error
	return commentList, err
}
