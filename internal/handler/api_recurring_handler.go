package handler

import (
	"net/http"
	"strconv"

	"homework-manager/internal/middleware"
	"homework-manager/internal/models"
	"homework-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type APIRecurringHandler struct {
	recurringService *service.RecurringAssignmentService
}

func NewAPIRecurringHandler() *APIRecurringHandler {
	return &APIRecurringHandler{
		recurringService: service.NewRecurringAssignmentService(),
	}
}

func (h *APIRecurringHandler) getUserID(c *gin.Context) uint {
	userID, _ := c.Get(middleware.UserIDKey)
	return userID.(uint)
}

func (h *APIRecurringHandler) ListRecurring(c *gin.Context) {
	userID := h.getUserID(c)

	recurringList, err := h.recurringService.GetAllByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recurring assignments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recurring_assignments": recurringList,
		"count":                 len(recurringList),
	})
}

func (h *APIRecurringHandler) GetRecurring(c *gin.Context) {
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	recurring, err := h.recurringService.GetByID(userID, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recurring assignment not found"})
		return
	}

	c.JSON(http.StatusOK, recurring)
}

type UpdateRecurringAPIInput struct {
	Title                 *string `json:"title"`
	Description           *string `json:"description"`
	Subject               *string `json:"subject"`
	Priority              *string `json:"priority"`
	RecurrenceType        *string `json:"recurrence_type"`
	RecurrenceInterval    *int    `json:"recurrence_interval"`
	RecurrenceWeekday     *int    `json:"recurrence_weekday"`
	RecurrenceDay         *int    `json:"recurrence_day"`
	DueTime               *string `json:"due_time"`
	EndType               *string `json:"end_type"`
	EndCount              *int    `json:"end_count"`
	EndDate               *string `json:"end_date"`  // YYYY-MM-DD
	IsActive              *bool   `json:"is_active"` // To stop/resume
	ReminderEnabled       *bool   `json:"reminder_enabled"`
	ReminderOffset        *int    `json:"reminder_offset"`
	UrgentReminderEnabled *bool   `json:"urgent_reminder_enabled"`
	EditBehavior          string  `json:"edit_behavior"` // this_only, this_and_future, all (default: this_only)
}

func (h *APIRecurringHandler) UpdateRecurring(c *gin.Context) {
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var input UpdateRecurringAPIInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	existing, err := h.recurringService.GetByID(userID, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recurring assignment not found"})
		return
	}

	if input.IsActive != nil {
		if err := h.recurringService.SetActive(userID, uint(id), *input.IsActive); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update active status"})
			return
		}
		existing.IsActive = *input.IsActive
	}

	serviceInput := service.UpdateRecurringInput{
		Title:                 input.Title,
		Description:           input.Description,
		Subject:               input.Subject,
		Priority:              input.Priority,
		RecurrenceType:        input.RecurrenceType,
		RecurrenceInterval:    input.RecurrenceInterval,
		RecurrenceWeekday:     input.RecurrenceWeekday,
		RecurrenceDay:         input.RecurrenceDay,
		DueTime:               input.DueTime,
		EndType:               input.EndType,
		EndCount:              input.EndCount,
		EditBehavior:          input.EditBehavior,
		ReminderEnabled:       input.ReminderEnabled,
		ReminderOffset:        input.ReminderOffset,
		UrgentReminderEnabled: input.UrgentReminderEnabled,
	}

	if input.EndDate != nil && *input.EndDate != "" {
		endDate, err := parseDateString(*input.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
			return
		}
		serviceInput.EndDate = &endDate
	}

	if serviceInput.EditBehavior == "" {
		serviceInput.EditBehavior = models.EditBehaviorThisOnly
	}

	updated, err := h.recurringService.Update(userID, uint(id), serviceInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recurring assignment"})
		return
	}

	updated.IsActive = existing.IsActive

	c.JSON(http.StatusOK, updated)
}

func (h *APIRecurringHandler) DeleteRecurring(c *gin.Context) {
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = h.recurringService.Delete(userID, uint(id), false)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recurring assignment not found or failed to delete"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Recurring assignment deleted"})
}
