package dal

import (
	"errors"
	"gorm.io/gorm"
)

// Favorite 记录用户点赞的视频
type Favorite struct {
	Id      int64 `gorm:"primaryKey"`
	User    User  `gorm:"foreignKey:UserId"`
	UserId  int64 `gorm:"not null"`
	Video   Video `gorm:"foreignKey:VideoId"`
	VideoId int64 `gorm:"not null"`
}

// AddFavorite 点赞操作，通过数据库事务保证数据一致性
func AddFavorite(userId, videoId int64) error {
	favorite := Favorite{
		UserId:  userId,
		VideoId: videoId,
	}
	// 检查是否已存在点赞记录
	if DB.Find(&Favorite{}).Where("user_id=?", userId).Where("video_id=?", videoId).RowsAffected > 0 {
		return errors.New("已经点赞过")
	}
	// 开启数据库事务，在 favorites 中添加记录，在 videos 中更改点赞数目
	if err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&favorite).Error; err != nil {
			return err
		}
		if err := tx.Model(&Video{}).Where("id=?", videoId).UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", 1)).Error; err != nil {
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
	if DB.Find(&Favorite{}).Where("user_id=?", userId).Where("video_id=?", videoId).RowsAffected == 0 {
		return errors.New("不存在点赞记录")
	}
	// 开启数据库事务，在 favorites 中添加记录，在 videos 中更改点赞数目
	if err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id=?", userId).Where("video_id=?", videoId).Delete(&Favorite{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&Video{}).Where("id=?", videoId).UpdateColumn("favorite_count", gorm.Expr("favorite_count - ?", 1)).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func GetFavoriteList(userId int64) ([]Video, error) {
	var favoriteList []Favorite
	var videoList []Video
	DB.Where("user_id=?", userId).Find(&favoriteList)

	DB.Table("favorites").Select("favorites.video_id,videos.*").
		Where("favorites.user_id=?", userId).
		Joins("LEFT JOIN videos ON favorites.video_id = videos.id").
		Find(&videoList)
	return videoList, nil
}
