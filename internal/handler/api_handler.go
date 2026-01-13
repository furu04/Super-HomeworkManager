package handler

import (
	"net/http"
	"strconv"
	"time"

	"homework-manager/internal/middleware"
	"homework-manager/internal/service"
	"homework-manager/internal/validation"

	"github.com/gin-gonic/gin"
)

type APIHandler struct {
	assignmentService *service.AssignmentService
	recurringService  *service.RecurringAssignmentService
}

func NewAPIHandler() *APIHandler {
	return &APIHandler{
		assignmentService: service.NewAssignmentService(),
		recurringService:  service.NewRecurringAssignmentService(),
	}
}

func (h *APIHandler) getUserID(c *gin.Context) uint {
	userID, _ := c.Get(middleware.UserIDKey)
	return userID.(uint)
}

func (h *APIHandler) ListAssignments(c *gin.Context) {
	userID := h.getUserID(c)
	filter := c.Query("filter") // pending, completed, overdue
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	switch filter {
	case "completed":
		result, err := h.assignmentService.GetCompletedByUserPaginated(userID, page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"assignments":  result.Assignments,
			"count":        len(result.Assignments),
			"total_count":  result.TotalCount,
			"total_pages":  result.TotalPages,
			"current_page": result.CurrentPage,
			"page_size":    result.PageSize,
		})
		return
	case "overdue":
		result, err := h.assignmentService.GetOverdueByUserPaginated(userID, page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"assignments":  result.Assignments,
			"count":        len(result.Assignments),
			"total_count":  result.TotalCount,
			"total_pages":  result.TotalPages,
			"current_page": result.CurrentPage,
			"page_size":    result.PageSize,
		})
		return
	case "pending":
		result, err := h.assignmentService.GetPendingByUserPaginated(userID, page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"assignments":  result.Assignments,
			"count":        len(result.Assignments),
			"total_count":  result.TotalCount,
			"total_pages":  result.TotalPages,
			"current_page": result.CurrentPage,
			"page_size":    result.PageSize,
		})
		return
	default:
		assignments, err := h.assignmentService.GetAllByUser(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
			return
		}

		totalCount := len(assignments)
		totalPages := (totalCount + pageSize - 1) / pageSize
		start := (page - 1) * pageSize
		end := start + pageSize
		if start > totalCount {
			start = totalCount
		}
		if end > totalCount {
			end = totalCount
		}

		c.JSON(http.StatusOK, gin.H{
			"assignments":  assignments[start:end],
			"count":        end - start,
			"total_count":  totalCount,
			"total_pages":  totalPages,
			"current_page": page,
			"page_size":    pageSize,
		})
	}
}

func (h *APIHandler) ListPendingAssignments(c *gin.Context) {
	userID := h.getUserID(c)
	page, pageSize := h.parsePagination(c)

	result, err := h.assignmentService.GetPendingByUserPaginated(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
		return
	}

	h.sendPaginatedResponse(c, result)
}

func (h *APIHandler) ListCompletedAssignments(c *gin.Context) {
	userID := h.getUserID(c)
	page, pageSize := h.parsePagination(c)

	result, err := h.assignmentService.GetCompletedByUserPaginated(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
		return
	}

	h.sendPaginatedResponse(c, result)
}

func (h *APIHandler) ListOverdueAssignments(c *gin.Context) {
	userID := h.getUserID(c)
	page, pageSize := h.parsePagination(c)

	result, err := h.assignmentService.GetOverdueByUserPaginated(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
		return
	}

	h.sendPaginatedResponse(c, result)
}

func (h *APIHandler) ListDueTodayAssignments(c *gin.Context) {
	userID := h.getUserID(c)

	assignments, err := h.assignmentService.GetDueTodayByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"assignments": assignments,
		"count":       len(assignments),
	})
}

func (h *APIHandler) ListDueThisWeekAssignments(c *gin.Context) {
	userID := h.getUserID(c)

	assignments, err := h.assignmentService.GetDueThisWeekByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"assignments": assignments,
		"count":       len(assignments),
	})
}

func (h *APIHandler) parsePagination(c *gin.Context) (page int, pageSize int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func (h *APIHandler) sendPaginatedResponse(c *gin.Context, result *service.PaginatedResult) {
	c.JSON(http.StatusOK, gin.H{
		"assignments":  result.Assignments,
		"count":        len(result.Assignments),
		"total_count":  result.TotalCount,
		"total_pages":  result.TotalPages,
		"current_page": result.CurrentPage,
		"page_size":    result.PageSize,
	})
}

func (h *APIHandler) GetAssignment(c *gin.Context) {
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	assignment, err := h.assignmentService.GetByID(userID, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
		return
	}

	c.JSON(http.StatusOK, assignment)
}

type CreateAssignmentInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Subject     string `json:"subject"`
	Priority    string `json:"priority"`
	DueDate     string `json:"due_date" binding:"required"`

	ReminderEnabled       bool   `json:"reminder_enabled"`
	ReminderAt            string `json:"reminder_at"`
	UrgentReminderEnabled *bool  `json:"urgent_reminder_enabled"`
	Recurrence            struct {
		Type     string      `json:"type"`
		Interval int         `json:"interval"`
		Weekday  interface{} `json:"weekday"`
		Day      interface{} `json:"day"`
		Until    struct {
			Type  string `json:"type"`
			Count int    `json:"count"`
			Date  string `json:"date"`
		} `json:"until"`
	} `json:"recurrence"`
}

func (h *APIHandler) CreateAssignment(c *gin.Context) {
	userID := h.getUserID(c)

	var input CreateAssignmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	if err := validation.ValidateAssignmentInput(input.Title, input.Description, input.Subject, input.Priority); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dueDate, err := parseDateString(input.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid due_date format. Use RFC3339 or 2006-01-02T15:04"})
		return
	}

	var reminderAt *time.Time
	if input.ReminderEnabled && input.ReminderAt != "" {
		reminderTime, err := parseDateString(input.ReminderAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reminder_at format"})
			return
		}
		reminderAt = &reminderTime
	}

	urgentReminder := true
	if input.UrgentReminderEnabled != nil {
		urgentReminder = *input.UrgentReminderEnabled
	}

	if input.Recurrence.Type != "" && input.Recurrence.Type != "none" {
		serviceInput := service.CreateRecurringAssignmentInput{
			Title:                 input.Title,
			Description:           input.Description,
			Subject:               input.Subject,
			Priority:              input.Priority,
			FirstDueDate:          dueDate,
			DueTime:               dueDate.Format("15:04"),
			RecurrenceType:        input.Recurrence.Type,
			RecurrenceInterval:    input.Recurrence.Interval,
			ReminderEnabled:       input.ReminderEnabled,
			ReminderOffset:        nil,
			UrgentReminderEnabled: urgentReminder,
		}

		if serviceInput.RecurrenceInterval < 1 {
			serviceInput.RecurrenceInterval = 1
		}

		if input.Recurrence.Weekday != nil {
			if wd, ok := input.Recurrence.Weekday.(float64); ok {
				wdInt := int(wd)
				serviceInput.RecurrenceWeekday = &wdInt
			}
		}

		if input.Recurrence.Day != nil {
			if d, ok := input.Recurrence.Day.(float64); ok {
				dInt := int(d)
				serviceInput.RecurrenceDay = &dInt
			}
		}

		serviceInput.EndType = input.Recurrence.Until.Type
		if serviceInput.EndType == "" {
			serviceInput.EndType = "never"
		}

		if serviceInput.EndType == "count" {
			count := input.Recurrence.Until.Count
			serviceInput.EndCount = &count
		} else if serviceInput.EndType == "date" && input.Recurrence.Until.Date != "" {
			endDate, err := parseDateString(input.Recurrence.Until.Date)
			if err == nil {
				serviceInput.EndDate = &endDate
			}
		}

		recurring, err := h.recurringService.Create(userID, serviceInput)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create recurring assignment: " + err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":              "Recurring assignment created",
			"recurring_assignment": recurring,
		})
		return
	}

	assignment, err := h.assignmentService.Create(userID, input.Title, input.Description, input.Subject, input.Priority, dueDate, input.ReminderEnabled, reminderAt, urgentReminder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create assignment"})
		return
	}

	c.JSON(http.StatusCreated, assignment)
}

type UpdateAssignmentInput struct {
	Title                 string `json:"title"`
	Description           string `json:"description"`
	Subject               string `json:"subject"`
	Priority              string `json:"priority"`
	DueDate               string `json:"due_date"`
	ReminderEnabled       *bool  `json:"reminder_enabled"`
	ReminderAt            string `json:"reminder_at"`
	UrgentReminderEnabled *bool  `json:"urgent_reminder_enabled"`
}

func (h *APIHandler) UpdateAssignment(c *gin.Context) {
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	existing, err := h.assignmentService.GetByID(userID, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
		return
	}

	var input UpdateAssignmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := validation.ValidateAssignmentInput(input.Title, input.Description, input.Subject, input.Priority); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	title := input.Title
	if title == "" {
		title = existing.Title
	}

	description := input.Description
	if description == "" {
		description = existing.Description
	}

	subject := input.Subject
	if subject == "" {
		subject = existing.Subject
	}

	priority := input.Priority
	if priority == "" {
		priority = existing.Priority
	}

	dueDate := existing.DueDate
	if input.DueDate != "" {
		parsedDate, err := parseDateString(input.DueDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid due_date format"})
			return
		}
		dueDate = parsedDate
	}

	reminderEnabled := existing.ReminderEnabled
	if input.ReminderEnabled != nil {
		reminderEnabled = *input.ReminderEnabled
	}

	reminderAt := existing.ReminderAt
	if input.ReminderAt != "" {
		parsedReminderAt, err := parseDateString(input.ReminderAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reminder_at format"})
			return
		}
		reminderAt = &parsedReminderAt
	} else if input.ReminderEnabled != nil && !*input.ReminderEnabled {
	}

	urgentReminderEnabled := existing.UrgentReminderEnabled
	if input.UrgentReminderEnabled != nil {
		urgentReminderEnabled = *input.UrgentReminderEnabled
	}

	assignment, err := h.assignmentService.Update(userID, uint(id), title, description, subject, priority, dueDate, reminderEnabled, reminderAt, urgentReminderEnabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update assignment"})
		return
	}

	c.JSON(http.StatusOK, assignment)
}

func (h *APIHandler) DeleteAssignment(c *gin.Context) {
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	deleteRecurring := c.Query("delete_recurring") == "true"

	if deleteRecurring {

		assignment, err := h.assignmentService.GetByID(userID, uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
			return
		}

		if assignment.RecurringAssignmentID != nil {
			if err := h.recurringService.Delete(userID, *assignment.RecurringAssignmentID, false); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recurring assignment"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Assignment and recurring settings deleted"})
			return
		}
	}

	if err := h.assignmentService.Delete(userID, uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assignment deleted"})
}

func (h *APIHandler) ToggleAssignment(c *gin.Context) {
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	assignment, err := h.assignmentService.ToggleComplete(userID, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
		return
	}

	c.JSON(http.StatusOK, assignment)
}

func (h *APIHandler) GetStatistics(c *gin.Context) {
	userID := h.getUserID(c)

	filter := service.StatisticsFilter{
		Subject:         c.Query("subject"),
		IncludeArchived: c.Query("include_archived") == "true",
	}

	if fromStr := c.Query("from"); fromStr != "" {
		fromDate, err := time.ParseInLocation("2006-01-02", fromStr, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'from' date format. Use YYYY-MM-DD"})
			return
		}
		filter.From = &fromDate
	}

	if toStr := c.Query("to"); toStr != "" {
		toDate, err := time.ParseInLocation("2006-01-02", toStr, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'to' date format. Use YYYY-MM-DD"})
			return
		}
		filter.To = &toDate
	}

	stats, err := h.assignmentService.GetStatistics(userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func parseDateString(dateStr string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006-01-02T15:04", dateStr, time.Local)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err == nil {
		return t.Add(23*time.Hour + 59*time.Minute), nil
	}
	return time.Time{}, err
}
