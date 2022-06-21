package dal

// Relation 用于维护用户关注关系，待之后完善
// 一行数据代表 "UserA 关注了 UserB"
type Relation struct {
	Id      int64 `gorm:"primaryKey"`
	UserA   User  `gorm:"foreignKey:UserAId"`
	UserAId int64 `gorm:"notnull"`
	UserB   User  `gorm:"foreignKey:UserBId"`
	UserBId int64 `gorm:"notnull"`
}
