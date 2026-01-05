package models

import (
	"time"

	"gorm.io/gorm"
)

type Assignment struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `json:"description"`
	Subject     string         `json:"subject"`
	Priority    string         `gorm:"not null;default:medium" json:"priority"` // low, medium, high
	DueDate     time.Time      `gorm:"not null" json:"due_date"`
	IsCompleted bool           `gorm:"default:false" json:"is_completed"`
	IsArchived  bool           `gorm:"default:false;index" json:"is_archived"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	// Reminder notification settings (one-time)
	ReminderEnabled bool       `gorm:"default:false" json:"reminder_enabled"`
	ReminderAt      *time.Time `json:"reminder_at,omitempty"`
	ReminderSent    bool       `gorm:"default:false;index" json:"reminder_sent"`
	// Urgent reminder settings (repeating until completed)
	UrgentReminderEnabled   bool       `gorm:"default:true" json:"urgent_reminder_enabled"`
	LastUrgentReminderSent  *time.Time `json:"last_urgent_reminder_sent,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
	DeletedAt               gorm.DeletedAt `gorm:"index" json:"-"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (a *Assignment) IsOverdue() bool {
	return !a.IsCompleted && time.Now().After(a.DueDate)
}

func (a *Assignment) IsDueToday() bool {
	now := time.Now()
	return a.DueDate.Year() == now.Year() &&
		a.DueDate.Month() == now.Month() &&
		a.DueDate.Day() == now.Day()
}

func (a *Assignment) IsDueThisWeek() bool {
	now := time.Now()
	weekLater := now.AddDate(0, 0, 7)
	return a.DueDate.After(now) && a.DueDate.Before(weekLater)
}
