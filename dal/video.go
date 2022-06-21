package dal

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zenpk/mini-douyin-ex/config"
	"mime/multipart"
	"os/exec"
	"path/filepath"
)

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

// Publish - 自动生成视频名称、路径，上传视频，生成封面
func Publish(c *gin.Context, userId int64, title string, filename string, data *multipart.FileHeader) error {
	playUrl := config.ServerAddr + "/static/videos/"
	coverUrl := config.ServerAddr + "/static/covers/"
	user, err := GetUserById(userId)
	if err != nil {
		return err
	}
	// 因为存储的文件名需要包含 videoId，所以先保存到数据库，利用 Id 自增特性获取 videoId
	// 相关信息存入数据库
	video := Video{
		Author:   user,
		UserId:   user.Id,
		PlayUrl:  playUrl,
		CoverUrl: coverUrl,
		Title:    title,
	}
	DB.Create(&video)
	finalName := fmt.Sprintf("%d_%d_%s", user.Id, video.Id, filename) // 保存的文件名，为防止文件名冲突，增加一项 videoId
	saveFile := filepath.Join("./public/videos/", finalName)
	// 将视频存入本地
	if err := c.SaveUploadedFile(data, saveFile); err != nil {
		return err
	}
	// 调用 ffmpeg 获取封面（第一帧老是黑屏，所以这里获取第 300 帧）
	// 当然更好的实践是先读取总共有多少帧，再获取中间的某一帧，这里为了简便实现就先这样了
	cmd := exec.Command(
		"ffmpeg", "-i", "./public/videos/"+finalName,
		"-vf", "select=eq(n\\, 300)", "-frames", "1",
		"./public/covers/"+finalName+".jpg",
	)
	//cmd.Stderr = os.Stderr // 输出错误信息
	if err := cmd.Run(); err != nil {
		return err
	}
	// 更新数据库中的视频和封面链接
	video.PlayUrl += finalName
	video.CoverUrl += finalName + ".jpg"
	DB.Save(&video)
	return nil
}

func GetPublishList(userId int64) []Video {
	var videoList []Video
	DB.Where("user_id=?", userId).Find(&videoList)
	return videoList
}

// GetFeed 获取时间倒序前 30 个视频
func GetFeed(userId int64) []Video {
	var videoList []Video
	DB.Preload("Author").Order("id desc").Limit(30).Find(&videoList)
	for i, v := range videoList {
		// 查找是否存在一条当前用户给该视频点赞的记录
		rows := DB.Where("video_id=?", v.Id).Where("user_id=?", userId).Find(&Favorite{}).RowsAffected
		videoList[i].IsFavorite = rows > 0 // 查找到了点赞记录
	}
	return videoList
}
