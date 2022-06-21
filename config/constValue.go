package config

// 保存一些常量

const (
	//IP = "101.43.179.27" // 服务器 IP

	IP = "192.168.50.163" // 本地 IP

	Port       = "10240"
	ServerAddr = "http://" + IP + ":" + Port
	DBName     = "douyin"    // MySQL 数据库名
	DBUserPass = "root:root" // MySQL 数据库用户名:密码
)

var (
	Secret = []byte("mini-douyin") // token 加密
)
