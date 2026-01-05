package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"homework-manager/internal/database"
	"homework-manager/internal/models"
)

// NotificationService handles Telegram and LINE notifications
type NotificationService struct {
	telegramBotToken string
}

// NewNotificationService creates a new notification service
func NewNotificationService(telegramBotToken string) *NotificationService {
	return &NotificationService{
		telegramBotToken: telegramBotToken,
	}
}

// GetUserSettings retrieves notification settings for a user
func (s *NotificationService) GetUserSettings(userID uint) (*models.UserNotificationSettings, error) {
	var settings models.UserNotificationSettings
	result := database.GetDB().Where("user_id = ?", userID).First(&settings)
	if result.Error != nil {
		// If not found, return a new empty settings object
		if result.RowsAffected == 0 {
			return &models.UserNotificationSettings{
				UserID: userID,
			}, nil
		}
		return nil, result.Error
	}
	return &settings, nil
}

// UpdateUserSettings updates notification settings for a user
func (s *NotificationService) UpdateUserSettings(userID uint, settings *models.UserNotificationSettings) error {
	settings.UserID = userID
	
	var existing models.UserNotificationSettings
	result := database.GetDB().Where("user_id = ?", userID).First(&existing)
	
	if result.RowsAffected == 0 {
		// Create new
		return database.GetDB().Create(settings).Error
	}
	
	// Update existing
	settings.ID = existing.ID
	return database.GetDB().Save(settings).Error
}

// SendTelegramNotification sends a message via Telegram Bot API
func (s *NotificationService) SendTelegramNotification(chatID, message string) error {
	if s.telegramBotToken == "" {
		return fmt.Errorf("telegram bot token is not configured")
	}
	if chatID == "" {
		return fmt.Errorf("telegram chat ID is empty")
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.telegramBotToken)
	
	payload := map[string]string{
		"chat_id":    chatID,
		"text":       message,
		"parse_mode": "HTML",
	}
	
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}
	
	return nil
}

// SendLineNotification sends a message via LINE Notify API
func (s *NotificationService) SendLineNotification(token, message string) error {
	if token == "" {
		return fmt.Errorf("LINE Notify token is empty")
	}

	apiURL := "https://notify-api.line.me/api/notify"
	
	data := url.Values{}
	data.Set("message", message)
	
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("LINE Notify API returned status %d", resp.StatusCode)
	}
	
	return nil
}

// SendAssignmentReminder sends a reminder notification for an assignment
func (s *NotificationService) SendAssignmentReminder(userID uint, assignment *models.Assignment) error {
	settings, err := s.GetUserSettings(userID)
	if err != nil {
		return err
	}

	message := fmt.Sprintf(
		"📚 課題リマインダー\n\n【%s】\n科目: %s\n期限: %s\n\n%s",
		assignment.Title,
		assignment.Subject,
		assignment.DueDate.Format("2006/01/02 15:04"),
		assignment.Description,
	)

	var errors []string

	// Send to Telegram if enabled
	if settings.TelegramEnabled && settings.TelegramChatID != "" {
		if err := s.SendTelegramNotification(settings.TelegramChatID, message); err != nil {
			errors = append(errors, fmt.Sprintf("Telegram: %v", err))
		}
	}

	// Send to LINE if enabled
	if settings.LineEnabled && settings.LineNotifyToken != "" {
		if err := s.SendLineNotification(settings.LineNotifyToken, message); err != nil {
			errors = append(errors, fmt.Sprintf("LINE: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("notification errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// SendUrgentReminder sends an urgent reminder notification for an assignment
func (s *NotificationService) SendUrgentReminder(userID uint, assignment *models.Assignment) error {
	settings, err := s.GetUserSettings(userID)
	if err != nil {
		return err
	}

	timeRemaining := time.Until(assignment.DueDate)
	var timeStr string
	if timeRemaining < 0 {
		timeStr = "期限切れ！"
	} else if timeRemaining < time.Hour {
		timeStr = fmt.Sprintf("あと%d分", int(timeRemaining.Minutes()))
	} else {
		timeStr = fmt.Sprintf("あと%d時間%d分", int(timeRemaining.Hours()), int(timeRemaining.Minutes())%60)
	}

	priorityEmoji := "📌"
	switch assignment.Priority {
	case "high":
		priorityEmoji = "🚨"
	case "medium":
		priorityEmoji = "⚠️"
	case "low":
		priorityEmoji = "📌"
	}

	message := fmt.Sprintf(
		"%s 督促通知！\n\n【%s】\n科目: %s\n期限: %s (%s)\n\n完了したらアプリで完了ボタンを押してください！",
		priorityEmoji,
		assignment.Title,
		assignment.Subject,
		assignment.DueDate.Format("2006/01/02 15:04"),
		timeStr,
	)

	var errors []string

	// Send to Telegram if enabled
	if settings.TelegramEnabled && settings.TelegramChatID != "" {
		if err := s.SendTelegramNotification(settings.TelegramChatID, message); err != nil {
			errors = append(errors, fmt.Sprintf("Telegram: %v", err))
		}
	}

	// Send to LINE if enabled
	if settings.LineEnabled && settings.LineNotifyToken != "" {
		if err := s.SendLineNotification(settings.LineNotifyToken, message); err != nil {
			errors = append(errors, fmt.Sprintf("LINE: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("notification errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// getUrgentReminderInterval returns the reminder interval based on priority
// high=10min, medium=30min, low=60min
func getUrgentReminderInterval(priority string) time.Duration {
	switch priority {
	case "high":
		return 10 * time.Minute
	case "medium":
		return 30 * time.Minute
	case "low":
		return 60 * time.Minute
	default:
		return 30 * time.Minute
	}
}

// ProcessPendingReminders checks and sends pending one-time reminders
func (s *NotificationService) ProcessPendingReminders() {
	now := time.Now()
	
	var assignments []models.Assignment
	result := database.GetDB().Where(
		"reminder_enabled = ? AND reminder_sent = ? AND reminder_at <= ? AND is_completed = ?",
		true, false, now, false,
	).Find(&assignments)
	
	if result.Error != nil {
		log.Printf("Error fetching pending reminders: %v", result.Error)
		return
	}
	
	for _, assignment := range assignments {
		if err := s.SendAssignmentReminder(assignment.UserID, &assignment); err != nil {
			log.Printf("Error sending reminder for assignment %d: %v", assignment.ID, err)
			continue
		}
		
		// Mark as sent
		database.GetDB().Model(&assignment).Update("reminder_sent", true)
		log.Printf("Sent reminder for assignment %d to user %d", assignment.ID, assignment.UserID)
	}
}

// ProcessUrgentReminders checks and sends urgent (repeating) reminders
// Starts 3 hours before deadline, repeats at interval based on priority
func (s *NotificationService) ProcessUrgentReminders() {
	now := time.Now()
	urgentStartTime := 3 * time.Hour // Start 3 hours before deadline
	
	var assignments []models.Assignment
	result := database.GetDB().Where(
		"urgent_reminder_enabled = ? AND is_completed = ? AND due_date > ?",
		true, false, now,
	).Find(&assignments)
	
	if result.Error != nil {
		log.Printf("Error fetching urgent reminders: %v", result.Error)
		return
	}
	
	for _, assignment := range assignments {
		timeUntilDue := assignment.DueDate.Sub(now)
		
		// Only send if within 3 hours of deadline
		if timeUntilDue > urgentStartTime {
			continue
		}
		
		// Check if enough time has passed since last urgent reminder
		interval := getUrgentReminderInterval(assignment.Priority)
		
		if assignment.LastUrgentReminderSent != nil {
			timeSinceLastReminder := now.Sub(*assignment.LastUrgentReminderSent)
			if timeSinceLastReminder < interval {
				continue
			}
		}
		
		// Send urgent reminder
		if err := s.SendUrgentReminder(assignment.UserID, &assignment); err != nil {
			log.Printf("Error sending urgent reminder for assignment %d: %v", assignment.ID, err)
			continue
		}
		
		// Update last sent time
		database.GetDB().Model(&assignment).Update("last_urgent_reminder_sent", now)
		log.Printf("Sent urgent reminder for assignment %d (priority: %s) to user %d", 
			assignment.ID, assignment.Priority, assignment.UserID)
	}
}

// StartReminderScheduler starts a background goroutine to process reminders
func (s *NotificationService) StartReminderScheduler() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		
		for range ticker.C {
			s.ProcessPendingReminders()
			s.ProcessUrgentReminders()
		}
	}()
	log.Println("Reminder scheduler started (one-time + urgent reminders)")
}

