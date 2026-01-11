package models

import (
	"time"

	"gorm.io/gorm"
)

type APIKey struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Name      string         `gorm:"not null" json:"name"`
	KeyHash   string         `gorm:"not null;uniqueIndex;size:255" json:"-"`
	LastUsed  *time.Time     `json:"last_used,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
