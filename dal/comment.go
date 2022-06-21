package dal

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type Comment struct {
	Id         int64  `json:"id" gorm:"primaryKey"`
	User       User   `json:"user" gorm:"foreignKey:UserId"`
	UserId     int64  `gorm:"not null"` // 评论对应的用户 Id
	Video      Video  `gorm:"foreignKey:VideoId"`
	VideoId    int64  `gorm:"not null"` // 评论对应的视频 Id
	Content    string `json:"content" gorm:"not null"`
	CreateDate string `json:"create_date" gorm:"not null"`
}

// AddComment 发表评论
func AddComment(userId, videoId int64, content string) (Comment, error) {
	// 日期 MM-DD
	timeFormat := time.Now().Format("01-02 15:04:05")
	user, err := GetUserById(userId)
	if err != nil {
		return Comment{}, err
	}
	comment := Comment{
		User:       user,
		VideoId:    videoId,
		UserId:     user.Id,
		Content:    content,
		CreateDate: timeFormat,
	}
	// 开启数据库事务，在 comments 中添加记录，在 videos 中更改评论数目
	if err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&comment).Error; err != nil {
			return err
		}
		if err := tx.Model(&Video{}).Where("id=?", videoId).UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).Error; err != nil {
			return err
		}
		// 返回 nil 提交事务
		return nil
	}); err != nil {
		return Comment{}, err
	}
	return comment, nil
}

// DeleteComment 删除评论
func DeleteComment(userId, videoId, commentId int64) error {
	// 检查是否有该用户对应的评论
	if DB.Find(&Comment{}).Where("id=?", commentId).Where("user_id=?", userId).RowsAffected == 0 {
		return errors.New("无法删除评论")
	}
	// 可选删除方案：1. 直接在数据库中删除; 2. 软删除：comments 中设置一个 deleted 列，用 bool 表示是否删除
	// 目前实现的是第一种
	// 开启数据库事务
	if err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id=?", commentId).Delete(&Comment{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&Video{}).Where("id=?", videoId).UpdateColumn("comment_count", gorm.Expr("comment_count - ?", 1)).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// GetCommentList 根据视频 id 读取评论列表
func GetCommentList(videoId int64) ([]Comment, error) {
	var commentList []Comment
	if err := DB.Preload("User").Where("video_id=?", videoId).Find(&commentList).Error; err != nil {
		return []Comment{Comment{}}, err
	}
	return commentList, nil
}
