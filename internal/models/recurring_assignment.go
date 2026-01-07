package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	RecurrenceNone    = "none"
	RecurrenceDaily   = "daily"
	RecurrenceWeekly  = "weekly"
	RecurrenceMonthly = "monthly"
)

const (
	EndTypeNever = "never"
	EndTypeCount = "count"
	EndTypeDate  = "date"
)

const (
	EditBehaviorThisOnly      = "this_only"
	EditBehaviorThisAndFuture = "this_and_future"
	EditBehaviorAll           = "all"
)

type RecurringAssignment struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	UserID      uint   `gorm:"not null;index" json:"user_id"`
	Title       string `gorm:"not null" json:"title"`
	Description string `json:"description"`
	Subject     string `json:"subject"`
	Priority    string `gorm:"not null;default:medium" json:"priority"`

	RecurrenceType     string `gorm:"not null;default:none" json:"recurrence_type"`
	RecurrenceInterval int    `gorm:"not null;default:1" json:"recurrence_interval"`
	RecurrenceWeekday  *int   `json:"recurrence_weekday,omitempty"`
	RecurrenceDay      *int   `json:"recurrence_day,omitempty"`
	DueTime            string `gorm:"not null" json:"due_time"`

	EndType        string     `gorm:"not null;default:never" json:"end_type"`
	EndCount       *int       `json:"end_count,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	GeneratedCount int        `gorm:"default:0" json:"generated_count"`
	EditBehavior   string     `gorm:"not null;default:this_only" json:"edit_behavior"`

	ReminderEnabled       bool           `gorm:"default:false" json:"reminder_enabled"`
	ReminderOffset        *int           `json:"reminder_offset,omitempty"`
	UrgentReminderEnabled bool           `gorm:"default:true" json:"urgent_reminder_enabled"`
	IsActive              bool           `gorm:"default:true" json:"is_active"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`

	User        *User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Assignments []Assignment `gorm:"foreignKey:RecurringAssignmentID" json:"assignments,omitempty"`
}

func (r *RecurringAssignment) ShouldGenerateNext() bool {
	if !r.IsActive || r.RecurrenceType == RecurrenceNone {
		return false
	}

	switch r.EndType {
	case EndTypeCount:
		if r.EndCount != nil && r.GeneratedCount >= *r.EndCount {
			return false
		}
	case EndTypeDate:
		if r.EndDate != nil && time.Now().After(*r.EndDate) {
			return false
		}
	}

	return true
}

func (r *RecurringAssignment) CalculateNextDueDate(lastDueDate time.Time) time.Time {
	var nextDate time.Time

	switch r.RecurrenceType {
	case RecurrenceDaily:
		nextDate = lastDueDate.AddDate(0, 0, r.RecurrenceInterval)
	case RecurrenceWeekly:
		nextDate = lastDueDate.AddDate(0, 0, 7*r.RecurrenceInterval)
	case RecurrenceMonthly:
		nextDate = lastDueDate.AddDate(0, r.RecurrenceInterval, 0)
		if r.RecurrenceDay != nil {
			day := *r.RecurrenceDay
			lastDayOfMonth := time.Date(nextDate.Year(), nextDate.Month()+1, 0, 0, 0, 0, 0, nextDate.Location()).Day()
			if day > lastDayOfMonth {
				day = lastDayOfMonth
			}
			nextDate = time.Date(nextDate.Year(), nextDate.Month(), day, nextDate.Hour(), nextDate.Minute(), 0, 0, nextDate.Location())
		}
	default:
		return lastDueDate
	}

	return nextDate
}
