package dal

type Video struct {
	Id            int64  `json:"id,omitempty" gorm:"primaryKey"`
	Author        User   `json:"author" gorm:"foreignKey:UserId"`
	UserId        int64  `gorm:"not null"` // 视频对应的用户 Id
	PlayUrl       string `json:"play_url" json:"play_url,omitempty" gorm:"not null"`
	CoverUrl      string `json:"cover_url,omitempty" gorm:"not null"`
	FavoriteCount int64  `json:"favorite_count,omitempty"`
	CommentCount  int64  `json:"comment_count,omitempty"`
	IsFavorite    bool   `json:"is_favorite,omitempty" gorm:"-:all"` // IsFavorite 是根据 favorites 表查询得到的，不需要存储
	Title         string `json:"title,omitempty"`                    // demo 里没有 title
}
