package models

import (
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// User 表：用户
type User struct {
	gorm.Model
	Username string `gorm:"unique"` // 用户名，唯一
	Password string // 密码
}

// Room 表：聊天室
type Room struct {
	gorm.Model
	Name   string // 聊天室名称
	UserID uint   // 创建者ID，外键关联 User
}

// Message 表：消息
type Message struct {
	gorm.Model
	Content string // 消息内容
	RoomID  uint   // 房间ID，外键关联 Room
	UserID  uint   // 用户ID，外键关联 User
}

func InitDB() {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:Bixilong5201!@tcp(localhost:3306)/chat_app?charset=utf8mb4&parseTime=True&loc=Local"

		//dsn = "root:Bixilong5201!@tcp(mysql:3306)/chat_app?charset=utf8mb4&parseTime=True&loc=Local"
	}
	log.Println("Connecting to MySQL with DSN:", dsn)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("[error] failed to initialize database, got error %v", err)
		panic("无法连接数据库")
	}
	DB = db
	log.Println("Successfully connected to MySQL")

	// 自动迁移，创建或更新表结构
	err = DB.AutoMigrate(&User{}, &Room{}, &Message{})
	if err != nil {
		log.Printf("[error] failed to migrate database, got error %v", err)
		panic("无法迁移数据库")
	}
	log.Println("Database migration completed")
}
