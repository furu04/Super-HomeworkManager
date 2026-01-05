package models

import (
	"time"

	"gorm.io/gorm"
)

// UserNotificationSettings stores user's notification preferences
type UserNotificationSettings struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	UserID          uint           `gorm:"uniqueIndex;not null" json:"user_id"`
	TelegramEnabled bool           `gorm:"default:false" json:"telegram_enabled"`
	TelegramChatID  string         `json:"telegram_chat_id"`
	LineEnabled     bool           `gorm:"default:false" json:"line_enabled"`
	LineNotifyToken string         `json:"-"` // Hide token from JSON
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	User *User `gorm:"foreignKey:UserID" json:"-"`
}
