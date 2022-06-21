package dal

type Comment struct {
	Id         int64  `json:"id,omitempty" gorm:"primaryKey"`
	User       User   `json:"user" gorm:"foreignKey:UserId"`
	UserId     int64  `gorm:"not null"` // 评论对应的用户 Id
	Video      Video  `gorm:"foreignKey:VideoId"`
	VideoId    int64  `gorm:"not null"` // 评论对应的视频 Id
	Content    string `json:"content,omitempty" gorm:"not null"`
	CreateDate string `json:"create_date,omitempty" gorm:"not null"`
}
