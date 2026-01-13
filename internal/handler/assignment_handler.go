package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"homework-manager/internal/middleware"
	"homework-manager/internal/models"
	"homework-manager/internal/service"
	"homework-manager/internal/validation"

	"github.com/gin-gonic/gin"
)

type AssignmentHandler struct {
	assignmentService   *service.AssignmentService
	notificationService *service.NotificationService
	recurringService    *service.RecurringAssignmentService
}

func NewAssignmentHandler(notificationService *service.NotificationService) *AssignmentHandler {
	return &AssignmentHandler{
		assignmentService:   service.NewAssignmentService(),
		notificationService: notificationService,
		recurringService:    service.NewRecurringAssignmentService(),
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
	now := time.Now()

	RenderHTML(c, http.StatusOK, "assignments/new.html", gin.H{
		"title":          "課題登録",
		"isAdmin":        role == "admin",
		"userName":       name,
		"currentWeekday": int(now.Weekday()),
		"currentDay":     now.Day(),
	})
}

func (h *AssignmentHandler) Create(c *gin.Context) {
	userID := h.getUserID(c)

	title := c.PostForm("title")
	description := c.PostForm("description")
	subject := c.PostForm("subject")
	priority := c.PostForm("priority")
	dueDateStr := c.PostForm("due_date")

	if err := validation.ValidateAssignmentInput(title, description, subject, priority); err != nil {
		role, _ := c.Get(middleware.UserRoleKey)
		name, _ := c.Get(middleware.UserNameKey)
		RenderHTML(c, http.StatusOK, "assignments/new.html", gin.H{
			"title":       "課題登録",
			"error":       err.Error(),
			"formTitle":   title,
			"description": description,
			"subject":     subject,
			"priority":    priority,
			"isAdmin":     role == "admin",
			"userName":    name,
		})
		return
	}

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

	recurrenceType := c.PostForm("recurrence_type")
	if recurrenceType != "" && recurrenceType != "none" {

		recurrenceInterval := 1
		if v, err := strconv.Atoi(c.PostForm("recurrence_interval")); err == nil && v > 0 {
			recurrenceInterval = v
		}

		var recurrenceWeekday *int
		if wd := c.PostForm("recurrence_weekday"); wd != "" {
			if v, err := strconv.Atoi(wd); err == nil && v >= 0 && v <= 6 {
				recurrenceWeekday = &v
			}
		}

		var recurrenceDay *int
		if d := c.PostForm("recurrence_day"); d != "" {
			if v, err := strconv.Atoi(d); err == nil && v >= 1 && v <= 31 {
				recurrenceDay = &v
			}
		}

		endType := c.PostForm("end_type")
		if endType == "" {
			endType = models.EndTypeNever
		}

		var endCount *int
		if ec := c.PostForm("end_count"); ec != "" {
			if v, err := strconv.Atoi(ec); err == nil && v > 0 {
				endCount = &v
			}
		}

		var endDate *time.Time
		if ed := c.PostForm("end_date"); ed != "" {
			if v, err := time.ParseInLocation("2006-01-02", ed, time.Local); err == nil {
				endDate = &v
			}
		}

		dueTime := dueDate.Format("15:04")

		recurringService := service.NewRecurringAssignmentService()
		input := service.CreateRecurringAssignmentInput{
			Title:                 title,
			Description:           description,
			Subject:               subject,
			Priority:              priority,
			RecurrenceType:        recurrenceType,
			RecurrenceInterval:    recurrenceInterval,
			RecurrenceWeekday:     recurrenceWeekday,
			RecurrenceDay:         recurrenceDay,
			DueTime:               dueTime,
			EndType:               endType,
			EndCount:              endCount,
			EndDate:               endDate,
			ReminderEnabled:       reminderEnabled,
			UrgentReminderEnabled: urgentReminderEnabled,
			FirstDueDate:          dueDate,
		}

		_, err = recurringService.Create(userID, input)
		if err != nil {
			role, _ := c.Get(middleware.UserRoleKey)
			name, _ := c.Get(middleware.UserNameKey)
			RenderHTML(c, http.StatusOK, "assignments/new.html", gin.H{
				"title":       "課題登録",
				"error":       "繰り返し課題の登録に失敗しました: " + err.Error(),
				"formTitle":   title,
				"description": description,
				"subject":     subject,
				"priority":    priority,
				"isAdmin":     role == "admin",
				"userName":    name,
			})
			return
		}
	} else {
		assignment, err := h.assignmentService.Create(userID, title, description, subject, priority, dueDate, reminderEnabled, reminderAt, urgentReminderEnabled)
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

		if h.notificationService != nil {
			go h.notificationService.SendAssignmentCreatedNotification(userID, assignment)
		}
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

	var recurring *models.RecurringAssignment
	if assignment.RecurringAssignmentID != nil {
		recurring, _ = h.recurringService.GetByID(userID, *assignment.RecurringAssignmentID)
	}

	role, _ := c.Get(middleware.UserRoleKey)
	name, _ := c.Get(middleware.UserNameKey)

	RenderHTML(c, http.StatusOK, "assignments/edit.html", gin.H{
		"title":      "課題編集",
		"assignment": assignment,
		"recurring":  recurring,
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

	if err := validation.ValidateAssignmentInput(title, description, subject, priority); err != nil {
		c.Redirect(http.StatusFound, "/assignments")
		return
	}

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

	deleteRecurringStr := c.Query("stop_recurring")
	if deleteRecurringStr != "" {
		recurringID, err := strconv.ParseUint(deleteRecurringStr, 10, 32)
		if err == nil {
			h.recurringService.Delete(userID, uint(recurringID), false)
		}
	}

	h.assignmentService.Delete(userID, uint(id))

	c.Redirect(http.StatusFound, "/assignments")
}

func (h *AssignmentHandler) Statistics(c *gin.Context) {
	userID := h.getUserID(c)
	role, _ := c.Get(middleware.UserRoleKey)
	name, _ := c.Get(middleware.UserNameKey)

	filter := service.StatisticsFilter{
		Subject:         c.Query("subject"),
		IncludeArchived: c.Query("include_archived") == "true",
	}

	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr != "" {
		fromDate, err := time.ParseInLocation("2006-01-02", fromStr, time.Local)
		if err == nil {
			filter.From = &fromDate
		}
	}

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

	subjects, _ := h.assignmentService.GetSubjectsWithArchived(userID, false)
	archivedSubjects, _ := h.assignmentService.GetArchivedSubjects(userID)

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

func (h *AssignmentHandler) StopRecurring(c *gin.Context) {
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/assignments")
		return
	}

	h.recurringService.SetActive(userID, uint(id), false)
	c.Redirect(http.StatusFound, "/assignments")
}

func (h *AssignmentHandler) ResumeRecurring(c *gin.Context) {
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/assignments")
		return
	}

	h.recurringService.SetActive(userID, uint(id), true)
	referer := c.Request.Referer()
	if referer == "" {
		referer = "/assignments"
	}
	c.Redirect(http.StatusFound, referer)
}

func (h *AssignmentHandler) ListRecurring(c *gin.Context) {
	userID := h.getUserID(c)
	role, _ := c.Get(middleware.UserRoleKey)
	name, _ := c.Get(middleware.UserNameKey)

	recurrings, err := h.recurringService.GetAllByUser(userID)
	if err != nil {
		recurrings = []models.RecurringAssignment{}
	}

	RenderHTML(c, http.StatusOK, "recurring/index.html", gin.H{
		"title":      "繰り返し設定一覧",
		"recurrings": recurrings,
		"isAdmin":    role == "admin",
		"userName":   name,
	})
}

func (h *AssignmentHandler) EditRecurring(c *gin.Context) {
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/assignments")
		return
	}

	recurring, err := h.recurringService.GetByID(userID, uint(id))
	if err != nil {
		c.Redirect(http.StatusFound, "/assignments")
		return
	}

	role, _ := c.Get(middleware.UserRoleKey)
	name, _ := c.Get(middleware.UserNameKey)

	RenderHTML(c, http.StatusOK, "recurring/edit.html", gin.H{
		"title":     "繰り返し課題の編集",
		"recurring": recurring,
		"isAdmin":   role == "admin",
		"userName":  name,
	})
}

func (h *AssignmentHandler) UpdateRecurring(c *gin.Context) {
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/assignments")
		return
	}

	title := c.PostForm("title")
	description := c.PostForm("description")
	subject := c.PostForm("subject")
	priority := c.PostForm("priority")
	recurrenceType := c.PostForm("recurrence_type")
	dueTime := c.PostForm("due_time")
	editBehavior := c.PostForm("edit_behavior")

	recurrenceInterval := 1
	if v, err := strconv.Atoi(c.PostForm("recurrence_interval")); err == nil && v > 0 {
		recurrenceInterval = v
	}

	var recurrenceWeekday *int
	if wd := c.PostForm("recurrence_weekday"); wd != "" {
		if v, err := strconv.Atoi(wd); err == nil && v >= 0 && v <= 6 {
			recurrenceWeekday = &v
		}
	}

	var recurrenceDay *int
	if d := c.PostForm("recurrence_day"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v >= 1 && v <= 31 {
			recurrenceDay = &v
		}
	}

	endType := c.PostForm("end_type")
	var endCount *int
	if ec := c.PostForm("end_count"); ec != "" {
		if v, err := strconv.Atoi(ec); err == nil && v > 0 {
			endCount = &v
		}
	}

	var endDate *time.Time
	if ed := c.PostForm("end_date"); ed != "" {
		if v, err := time.ParseInLocation("2006-01-02", ed, time.Local); err == nil {
			endDate = &v
		}
	}

	input := service.UpdateRecurringInput{
		Title:              &title,
		Description:        &description,
		Subject:            &subject,
		Priority:           &priority,
		RecurrenceType:     &recurrenceType,
		RecurrenceInterval: &recurrenceInterval,
		RecurrenceWeekday:  recurrenceWeekday,
		RecurrenceDay:      recurrenceDay,
		DueTime:            &dueTime,
		EndType:            &endType,
		EndCount:           endCount,
		EndDate:            endDate,
		EditBehavior:       editBehavior,
	}

	_, err = h.recurringService.Update(userID, uint(id), input)
	if err != nil {
		c.Redirect(http.StatusFound, "/recurring/"+c.Param("id")+"/edit")
		return
	}

	c.Redirect(http.StatusFound, "/assignments")
}

func (h *AssignmentHandler) DeleteRecurring(c *gin.Context) {
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/assignments")
		return
	}

	h.recurringService.Delete(userID, uint(id), false)

	c.Redirect(http.StatusFound, "/assignments")
}
