package dal

// Favorite 记录用户点赞的视频
type Favorite struct {
	Id      int64 `gorm:"primaryKey"`
	User    User  `gorm:"foreignKey:UserId"`
	UserId  int64 `gorm:"not null"`
	Video   Video `gorm:"foreignKey:VideoId"`
	VideoId int64 `gorm:"not null"`
}
