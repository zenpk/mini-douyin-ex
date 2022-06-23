package dal

import (
	"errors"
	"gorm.io/gorm"
)

// Favorite 记录用户点赞的视频，使用复合主键
type Favorite struct {
	UserId  int64 `gorm:"primaryKey;autoIncrement:false"`
	VideoId int64 `gorm:"primaryKey;autoIncrement:false"`
}

// AddFavorite 点赞操作，通过数据库事务保证数据一致性
func AddFavorite(userId, videoId int64) error {
	favorite := Favorite{
		UserId:  userId,
		VideoId: videoId,
	}
	// 检查是否已存在点赞记录
	if DB.Find(&Favorite{}).Where("user_id = ?", userId).Where("video_id = ?", videoId).RowsAffected > 0 {
		return errors.New("已经点赞过")
	}
	// 开启数据库事务，在 favorites 中添加记录，在 videos 中更改点赞数目
	if err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&favorite).Error; err != nil {
			return err
		}
		if err := tx.Model(&Video{}).Where("id = ?", videoId).UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", 1)).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// DeleteFavorite 取消点赞操作
func DeleteFavorite(userId, videoId int64) error {
	// 检查是否存在点赞记录
	if DB.Find(&Favorite{}).Where("user_id = ?", userId).Where("video_id = ?", videoId).RowsAffected == 0 {
		return errors.New("不存在点赞记录")
	}
	// 开启数据库事务，在 favorites 中添加记录，在 videos 中更改点赞数目
	if err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userId).Where("video_id = ?", videoId).Delete(&Favorite{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&Video{}).Where("id = ?", videoId).UpdateColumn("favorite_count", gorm.Expr("favorite_count - ?", 1)).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func GetFavoriteByUserId(userId int64) ([]Favorite, error) {
	var favoriteList []Favorite
	err := DB.Where("user_id = ?", userId).Find(&favoriteList).Error
	return favoriteList, err
}
