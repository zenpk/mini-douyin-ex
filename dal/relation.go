package dal

import (
	"errors"
	"gorm.io/gorm"
)

// Relation 用于维护用户关注关系，待之后完善
// 一行数据代表 "UserA 关注了 UserB"
type Relation struct {
	Id      int64 `gorm:"primaryKey"`
	UserA   User  `gorm:"foreignKey:UserAId"`
	UserAId int64 `gorm:"notnull"`
	UserB   User  `gorm:"foreignKey:UserBId"`
	UserBId int64 `gorm:"notnull"`
}

// IsFollowed 查询用户 A 是否关注了用户 B
func IsFollowed(userIdA, userIdB int64) bool {
	return DB.Find(&Relation{}).Where("user_a_id=?", userIdA).Where("user_b_id=?", userIdB).RowsAffected > 0
}

func Follow(userAId, userBId int64) error {
	// 添加记录前先查找是否存在
	if DB.Find(&Relation{}).Where("user_a_id=?", userAId).Where("user_b_id", userBId).RowsAffected > 0 {
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
		if err := tx.Model(&User{}).Where("id=?", userAId).UpdateColumn("follow_count", gorm.Expr("follow_count + ?", 1)).Error; err != nil {
			return err
		}
		if err := tx.Model(&User{}).Where("id=?", userBId).UpdateColumn("follower_count", gorm.Expr("follower_count + ?", 1)).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func Unfollow(userAId, userBId int64) error {
	// 删除记录前先查找是否存在
	if DB.Find(&Relation{}).Where("user_a_id=?", userAId).Where("user_b_id", userBId).RowsAffected == 0 {
		return errors.New("没有关注记录")
	}

	// 开启数据库事务
	if err := DB.Transaction(func(tx *gorm.DB) error {
		// 删除关注关系
		if err := tx.Where("user_a_id=?", userAId).Where("user_b_id=?", userBId).Delete(&Relation{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&User{}).Where("id=?", userAId).UpdateColumn("follow_count", gorm.Expr("follow_count - ?", 1)).Error; err != nil {
			return err
		}
		if err := tx.Model(&User{}).Where("id=?", userBId).UpdateColumn("follower_count", gorm.Expr("follower_count - ?", 1)).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// GetFollowList 获取查询用户的关注列表
func GetFollowList(userId int64) ([]User, error) {
	var followList []Relation
	// 加载 UserB 即加载当前用户关注的用户
	DB.Preload("UserB").Where("user_a_id=?", userId).Find(&followList)
	// 这里直接暴力复制了，不知道 Go 语言有无更好的方法可以提取结构体数组中的元素
	followUserList := make([]User, len(followList))
	for i, f := range followList {
		followUserList[i] = f.UserB
	}
	return followUserList, nil
}

// GetFollowerList 获取查询用户的粉丝列表
func GetFollowerList(userId int64) ([]User, error) {
	var followerList []Relation
	// 加载 UserA 即加载当前用户的粉丝
	DB.Preload("UserA").Where("user_b_id=?", userId).Find(&followerList)
	// 这里直接暴力复制了，不知道 Go 语言有无更好的方法可以提取结构体数组中的元素
	followerUserList := make([]User, len(followerList))
	for i, f := range followerList {
		followerUserList[i] = f.UserA
	}
	return followerUserList, nil
}
