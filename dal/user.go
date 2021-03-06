package dal

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id            int64  `json:"id" gorm:"primaryKey"`
	Name          string `json:"name" gorm:"unique; not null"`
	Password      string `gorm:"not null" redistructhash:"no"`
	FollowCount   int64  `json:"follow_count"`
	FollowerCount int64  `json:"follower_count"`
	IsFollow      bool   `json:"is_follow" gorm:"-:all"` // IsFollow 是根据 relations 表查询得到的，不需要存储
}

// bCryptPassword 对密码加密
func bCryptPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func Register(name, password string) (user User, err error) {
	// 根据用户名的唯一性，查找是否存在该用户，如果不存在则将用户信息存入数据库中
	if DB.Where("name = ?", name).Find(&User{}).RowsAffected > 0 {
		return User{}, errors.New("用户名已存在")
	} else {
		passwordHash, _ := bCryptPassword(password) // 将密码加密
		newUser := User{
			Name:     name,
			Password: passwordHash,
		}
		DB.Create(&newUser) // 存入数据库
		return newUser, nil
	}
}

func Login(name, password string) (User, error) {
	// 查找数据库中对应的用户名，并检查密码
	var user User
	if DB.Where("name = ?", name).First(&user).RowsAffected > 0 {
		passwordHashByte := []byte(user.Password)
		passwordByte := []byte(password)
		// 检查密码是否正确，使用 BCrypt 内置的比较函数
		if err := bcrypt.CompareHashAndPassword(passwordHashByte, passwordByte); errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return User{}, errors.New("密码错误")
		} else { // 密码正确
			return user, nil
		}
	} else {
		return User{}, errors.New("用户不存在")
	}
}

func GetUserById(id int64) (User, error) {
	var user User
	err := DB.Find(&user, id).Error
	return user, err
}
