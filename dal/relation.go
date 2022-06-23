package dal

import (
	"errors"
	"gorm.io/gorm"
)

// Relation 用于维护用户关注关系，使用复合主键
// 一行数据代表 "UserA 关注了 UserB"
type Relation struct {
	UserAId int64 `gorm:"primaryKey;autoIncrement:false"`
	UserBId int64 `gorm:"primaryKey;autoIncrement:false"`
}

func AddFollow(userAId, userBId int64) error {
	// 添加记录前先查找是否存在
	if DB.Where("user_a_id = ? AND user_b_id = ?", userAId, userBId).Find(&Relation{}).RowsAffected > 0 {
		return errors.New("已经关注过")
	}
	relation := Relation{
		UserAId: userAId,
		UserBId: userBId,
	}
	// 开启数据库事务，在 relations 中添加记录，在 users 中更改关注数
	if err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&relation).Error; err != nil {
			return err
		}
		// 增加关注数、被关注数
		if err := tx.Model(&User{}).Where("id = ?", userAId).UpdateColumn("follow_count", gorm.Expr("follow_count + ?", 1)).Error; err != nil {
			return err
		}
		if err := tx.Model(&User{}).Where("id = ?", userBId).UpdateColumn("follower_count", gorm.Expr("follower_count + ?", 1)).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func DeleteFollow(userAId, userBId int64) error {
	// 删除记录前先查找是否存在
	var relation Relation
	if DB.Where("user_a_id = ? AND user_b_id", userAId, userBId).First(&relation).RowsAffected <= 0 {
		return errors.New("没有关注记录")
	}

	// 开启数据库事务
	if err := DB.Transaction(func(tx *gorm.DB) error {
		// 删除关注关系
		if err := tx.Delete(&relation).Error; err != nil {
			return err
		}
		if err := tx.Model(&User{}).Where("id = ?", userAId).UpdateColumn("follow_count", gorm.Expr("follow_count - ?", 1)).Error; err != nil {
			return err
		}
		if err := tx.Model(&User{}).Where("id = ?", userBId).UpdateColumn("follower_count", gorm.Expr("follower_count - ?", 1)).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// GetFollowList 获取查询用户的所有关注的 id
func GetFollowList(userId int64) ([]int64, error) {
	var followList []Relation
	if err := DB.Where("user_a_id = ?", userId).Find(&followList).Error; err != nil {
		return []int64{}, err
	}
	followIdList := make([]int64, len(followList))
	for i, f := range followList {
		followIdList[i] = f.UserAId
	}
	return followIdList, nil
}

// GetFollowerList 获取查询用户的所有粉丝的 id
func GetFollowerList(userId int64) ([]int64, error) {
	var followList []Relation
	if err := DB.Where("user_b_id = ?", userId).Find(&followList).Error; err != nil {
		return []int64{}, err
	}
	followerIdList := make([]int64, len(followList))
	for i, f := range followList {
		followerIdList[i] = f.UserAId
	}
	return followerIdList, nil
}
