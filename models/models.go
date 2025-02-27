// 定义数据库初始化以及数据库表
// User:用户：用户名 密码
// Room:聊天室，聊天室名称 创建者
// Message:消息 内容 房间 用户
package models

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := "host=localhost user=postgres password=bixilong5201 dbname=chat_app port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	DB = db
	DB.AutoMigrate(&User{}, &Room{}, &Message{})
}

type User struct {
	gorm.Model
	Username string `gorm:"unique"`
	Password string
}

type Room struct {
	gorm.Model
	Name   string
	UserID uint
}

type Message struct {
	gorm.Model
	Content string
	RoomID  uint
	UserID  uint
}
