package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"homework-manager/internal/middleware"
	"homework-manager/internal/models"
	"homework-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type AssignmentHandler struct {
	assignmentService *service.AssignmentService
}

func NewAssignmentHandler() *AssignmentHandler {
	return &AssignmentHandler{
		assignmentService: service.NewAssignmentService(),
	}
}

func (h *AssignmentHandler) getUserID(c *gin.Context) uint {
	userID, _ := c.Get(middleware.UserIDKey)
	return userID.(uint)
}

func (h *AssignmentHandler) Dashboard(c *gin.Context) {
	userID := h.getUserID(c)
	stats, _ := h.assignmentService.GetDashboardStats(userID)
	dueToday, _ := h.assignmentService.GetDueTodayByUser(userID)
	overdue, _ := h.assignmentService.GetOverdueByUser(userID)
	upcoming, _ := h.assignmentService.GetDueThisWeekByUser(userID)

	role, _ := c.Get(middleware.UserRoleKey)
	name, _ := c.Get(middleware.UserNameKey)

	RenderHTML(c, http.StatusOK, "dashboard.html", gin.H{
		"title":    "ダッシュボード",
		"stats":    stats,
		"dueToday": dueToday,
		"overdue":  overdue,
		"upcoming": upcoming,
		"isAdmin":  role == "admin",
		"userName": name,
	})
}

func (h *AssignmentHandler) Index(c *gin.Context) {
	userID := h.getUserID(c)
	filter := c.Query("filter")
	filter = strings.TrimSpace(filter)
	if filter == "" {
		filter = "pending"
	}
	query := c.Query("q")
	priority := c.Query("priority")
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	const pageSize = 10

	result, err := h.assignmentService.SearchAssignments(userID, query, priority, filter, page, pageSize)

	var assignments []models.Assignment
	var totalPages, currentPage int
	if err != nil || result == nil {
		assignments = []models.Assignment{}
		totalPages = 1
		currentPage = 1
	} else {
		assignments = result.Assignments
		totalPages = result.TotalPages
		currentPage = result.CurrentPage
	}

	role, _ := c.Get(middleware.UserRoleKey)
	name, _ := c.Get(middleware.UserNameKey)

	RenderHTML(c, http.StatusOK, "assignments/index.html", gin.H{
		"title":       "課題一覧",
		"assignments": assignments,
		"filter":      filter,
		"query":       query,
		"priority":    priority,
		"isAdmin":     role == "admin",
		"userName":    name,
		"currentPage": currentPage,
		"totalPages":  totalPages,
		"hasPrev":     currentPage > 1,
		"hasNext":     currentPage < totalPages,
		"prevPage":    currentPage - 1,
		"nextPage":    currentPage + 1,
	})
}

func (h *AssignmentHandler) New(c *gin.Context) {
	role, _ := c.Get(middleware.UserRoleKey)
	name, _ := c.Get(middleware.UserNameKey)

	RenderHTML(c, http.StatusOK, "assignments/new.html", gin.H{
		"title":    "課題登録",
		"isAdmin":  role == "admin",
		"userName": name,
	})
}

func (h *AssignmentHandler) Create(c *gin.Context) {
	userID := h.getUserID(c)

	title := c.PostForm("title")
	description := c.PostForm("description")
	subject := c.PostForm("subject")
	priority := c.PostForm("priority")
	dueDateStr := c.PostForm("due_date")

	// Parse reminder settings
	reminderEnabled := c.PostForm("reminder_enabled") == "on"
	reminderAtStr := c.PostForm("reminder_at")
	var reminderAt *time.Time
	if reminderEnabled && reminderAtStr != "" {
		if parsed, err := time.ParseInLocation("2006-01-02T15:04", reminderAtStr, time.Local); err == nil {
			reminderAt = &parsed
		}
	}
	urgentReminderEnabled := c.PostForm("urgent_reminder_enabled") == "on"

	dueDate, err := time.ParseInLocation("2006-01-02T15:04", dueDateStr, time.Local)
	if err != nil {
		dueDate, err = time.ParseInLocation("2006-01-02", dueDateStr, time.Local)
		if err != nil {
			role, _ := c.Get(middleware.UserRoleKey)
			name, _ := c.Get(middleware.UserNameKey)
			RenderHTML(c, http.StatusOK, "assignments/new.html", gin.H{
				"title":       "課題登録",
				"error":       "提出期限の形式が正しくありません",
				"formTitle":   title,
				"description": description,
				"subject":     subject,
				"priority":    priority,
				"isAdmin":     role == "admin",
				"userName":    name,
			})
			return
		}
		dueDate = dueDate.Add(23*time.Hour + 59*time.Minute)
	}

	_, err = h.assignmentService.Create(userID, title, description, subject, priority, dueDate, reminderEnabled, reminderAt, urgentReminderEnabled)
	if err != nil {
		role, _ := c.Get(middleware.UserRoleKey)
		name, _ := c.Get(middleware.UserNameKey)
		RenderHTML(c, http.StatusOK, "assignments/new.html", gin.H{
			"title":       "課題登録",
			"error":       "課題の登録に失敗しました",
			"formTitle":   title,
			"description": description,
			"subject":     subject,
			"priority":    priority,
			"isAdmin":     role == "admin",
			"userName":    name,
		})
		return
	}

	c.Redirect(http.StatusFound, "/assignments")
}

func (h *AssignmentHandler) Edit(c *gin.Context) {
	userID := h.getUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	assignment, err := h.assignmentService.GetByID(userID, uint(id))
	if err != nil {
		c.Redirect(http.StatusFound, "/assignments")
		return
	}

	role, _ := c.Get(middleware.UserRoleKey)
	name, _ := c.Get(middleware.UserNameKey)

	RenderHTML(c, http.StatusOK, "assignments/edit.html", gin.H{
		"title":      "課題編集",
		"assignment": assignment,
		"isAdmin":    role == "admin",
		"userName":   name,
	})
}

func (h *AssignmentHandler) Update(c *gin.Context) {
	userID := h.getUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	title := c.PostForm("title")
	description := c.PostForm("description")
	subject := c.PostForm("subject")
	priority := c.PostForm("priority")
	dueDateStr := c.PostForm("due_date")

	// Parse reminder settings
	reminderEnabled := c.PostForm("reminder_enabled") == "on"
	reminderAtStr := c.PostForm("reminder_at")
	var reminderAt *time.Time
	if reminderEnabled && reminderAtStr != "" {
		if parsed, err := time.ParseInLocation("2006-01-02T15:04", reminderAtStr, time.Local); err == nil {
			reminderAt = &parsed
		}
	}
	urgentReminderEnabled := c.PostForm("urgent_reminder_enabled") == "on"

	dueDate, err := time.ParseInLocation("2006-01-02T15:04", dueDateStr, time.Local)
	if err != nil {
		dueDate, err = time.ParseInLocation("2006-01-02", dueDateStr, time.Local)
		if err != nil {
			c.Redirect(http.StatusFound, "/assignments")
			return
		}
		dueDate = dueDate.Add(23*time.Hour + 59*time.Minute)
	}

	_, err = h.assignmentService.Update(userID, uint(id), title, description, subject, priority, dueDate, reminderEnabled, reminderAt, urgentReminderEnabled)
	if err != nil {
		c.Redirect(http.StatusFound, "/assignments")
		return
	}

	c.Redirect(http.StatusFound, "/assignments")
}

func (h *AssignmentHandler) Toggle(c *gin.Context) {
	userID := h.getUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	h.assignmentService.ToggleComplete(userID, uint(id))

	referer := c.Request.Referer()
	if referer == "" {
		referer = "/assignments"
	}
	c.Redirect(http.StatusFound, referer)
}

func (h *AssignmentHandler) Delete(c *gin.Context) {
	userID := h.getUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	h.assignmentService.Delete(userID, uint(id))

	c.Redirect(http.StatusFound, "/assignments")
}

func (h *AssignmentHandler) Statistics(c *gin.Context) {
	userID := h.getUserID(c)
	role, _ := c.Get(middleware.UserRoleKey)
	name, _ := c.Get(middleware.UserNameKey)

	// Parse filter parameters
	filter := service.StatisticsFilter{
		Subject:         c.Query("subject"),
		IncludeArchived: c.Query("include_archived") == "true",
	}

	fromStr := c.Query("from")
	toStr := c.Query("to")

	// Parse from date
	if fromStr != "" {
		fromDate, err := time.ParseInLocation("2006-01-02", fromStr, time.Local)
		if err == nil {
			filter.From = &fromDate
		}
	}

	// Parse to date
	if toStr != "" {
		toDate, err := time.ParseInLocation("2006-01-02", toStr, time.Local)
		if err == nil {
			filter.To = &toDate
		}
	}

	stats, err := h.assignmentService.GetStatistics(userID, filter)
	if err != nil {
		RenderHTML(c, http.StatusInternalServerError, "error.html", gin.H{
			"title":   "エラー",
			"message": "統計情報の取得に失敗しました",
		})
		return
	}

	// Get available subjects for filter dropdown (exclude archived)
	subjects, _ := h.assignmentService.GetSubjectsWithArchived(userID, false)
	archivedSubjects, _ := h.assignmentService.GetArchivedSubjects(userID)

	// Create a map for quick lookup of archived subjects
	archivedMap := make(map[string]bool)
	for _, s := range archivedSubjects {
		archivedMap[s] = true
	}

	RenderHTML(c, http.StatusOK, "assignments/statistics.html", gin.H{
		"title":            "統計",
		"stats":            stats,
		"subjects":         subjects,
		"archivedSubjects": archivedMap,
		"selectedSubject":  filter.Subject,
		"fromDate":         fromStr,
		"toDate":           toStr,
		"includeArchived":  filter.IncludeArchived,
		"isAdmin":          role == "admin",
		"userName":         name,
	})
}

func (h *AssignmentHandler) ArchiveSubject(c *gin.Context) {
	userID := h.getUserID(c)
	subject := c.PostForm("subject")

	if subject != "" {
		h.assignmentService.ArchiveSubject(userID, subject)
	}

	c.Redirect(http.StatusFound, "/statistics")
}

func (h *AssignmentHandler) UnarchiveSubject(c *gin.Context) {
	userID := h.getUserID(c)
	subject := c.PostForm("subject")

	if subject != "" {
		h.assignmentService.UnarchiveSubject(userID, subject)
	}

	c.Redirect(http.StatusFound, "/statistics?include_archived=true")
}


