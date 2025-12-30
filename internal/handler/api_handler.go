package handler

import (
	"net/http"
	"strconv"
	"time"

	"homework-manager/internal/middleware"
	"homework-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type APIHandler struct {
	assignmentService *service.AssignmentService
}

func NewAPIHandler() *APIHandler {
	return &APIHandler{
		assignmentService: service.NewAssignmentService(),
	}
}

func (h *APIHandler) getUserID(c *gin.Context) uint {
	userID, _ := c.Get(middleware.UserIDKey)
	return userID.(uint)
}

// ListAssignments returns all assignments for the authenticated user with pagination
// GET /api/v1/assignments?filter=pending&page=1&page_size=20
func (h *APIHandler) ListAssignments(c *gin.Context) {
	userID := h.getUserID(c)
	filter := c.Query("filter") // pending, completed, overdue

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100 // Maximum page size to prevent abuse
	}

	// Use paginated methods for filtered queries
	switch filter {
	case "completed":
		result, err := h.assignmentService.GetCompletedByUserPaginated(userID, page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"assignments":   result.Assignments,
			"count":         len(result.Assignments),
			"total_count":   result.TotalCount,
			"total_pages":   result.TotalPages,
			"current_page":  result.CurrentPage,
			"page_size":     result.PageSize,
		})
		return
	case "overdue":
		result, err := h.assignmentService.GetOverdueByUserPaginated(userID, page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"assignments":   result.Assignments,
			"count":         len(result.Assignments),
			"total_count":   result.TotalCount,
			"total_pages":   result.TotalPages,
			"current_page":  result.CurrentPage,
			"page_size":     result.PageSize,
		})
		return
	case "pending":
		result, err := h.assignmentService.GetPendingByUserPaginated(userID, page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"assignments":   result.Assignments,
			"count":         len(result.Assignments),
			"total_count":   result.TotalCount,
			"total_pages":   result.TotalPages,
			"current_page":  result.CurrentPage,
			"page_size":     result.PageSize,
		})
		return
	default:
		// For "all" filter, use simple pagination without a dedicated method
		assignments, err := h.assignmentService.GetAllByUser(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
			return
		}

		// Manual pagination for all assignments
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
			"assignments":   assignments[start:end],
			"count":         end - start,
			"total_count":   totalCount,
			"total_pages":   totalPages,
			"current_page":  page,
			"page_size":     pageSize,
		})
	}
}

// ListPendingAssignments returns pending assignments with pagination
// GET /api/v1/assignments/pending?page=1&page_size=20
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

// ListCompletedAssignments returns completed assignments with pagination
// GET /api/v1/assignments/completed?page=1&page_size=20
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

// ListOverdueAssignments returns overdue assignments with pagination
// GET /api/v1/assignments/overdue?page=1&page_size=20
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

// ListDueTodayAssignments returns assignments due today
// GET /api/v1/assignments/due-today
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

// ListDueThisWeekAssignments returns assignments due within this week
// GET /api/v1/assignments/due-this-week
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

// parsePagination extracts and validates pagination parameters
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

// sendPaginatedResponse sends a standard paginated JSON response
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

// GetAssignment returns a single assignment by ID
// GET /api/v1/assignments/:id
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

// CreateAssignmentInput represents the JSON input for creating an assignment
type CreateAssignmentInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Subject     string `json:"subject"`
	Priority    string `json:"priority"` // low, medium, high (default: medium)
	DueDate     string `json:"due_date" binding:"required"` // RFC3339 or 2006-01-02T15:04
}

// CreateAssignment creates a new assignment
// POST /api/v1/assignments
func (h *APIHandler) CreateAssignment(c *gin.Context) {
	userID := h.getUserID(c)

	var input CreateAssignmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: title and due_date are required"})
		return
	}

	dueDate, err := time.Parse(time.RFC3339, input.DueDate)
	if err != nil {
		dueDate, err = time.ParseInLocation("2006-01-02T15:04", input.DueDate, time.Local)
		if err != nil {
			dueDate, err = time.ParseInLocation("2006-01-02", input.DueDate, time.Local)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid due_date format. Use RFC3339 or 2006-01-02T15:04"})
				return
			}
			dueDate = dueDate.Add(23*time.Hour + 59*time.Minute)
		}
	}

	assignment, err := h.assignmentService.Create(userID, input.Title, input.Description, input.Subject, input.Priority, dueDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create assignment"})
		return
	}

	c.JSON(http.StatusCreated, assignment)
}

// UpdateAssignmentInput represents the JSON input for updating an assignment
type UpdateAssignmentInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Subject     string `json:"subject"`
	Priority    string `json:"priority"`
	DueDate     string `json:"due_date"`
}

// UpdateAssignment updates an existing assignment
// PUT /api/v1/assignments/:id
func (h *APIHandler) UpdateAssignment(c *gin.Context) {
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	// Get existing assignment
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

	// Use existing values if not provided
	title := input.Title
	if title == "" {
		title = existing.Title
	}

	description := input.Description
	subject := input.Subject
	priority := input.Priority
	if priority == "" {
		priority = existing.Priority
	}

	dueDate := existing.DueDate
	if input.DueDate != "" {
		dueDate, err = time.Parse(time.RFC3339, input.DueDate)
		if err != nil {
			dueDate, err = time.ParseInLocation("2006-01-02T15:04", input.DueDate, time.Local)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid due_date format"})
				return
			}
		}
	}

	assignment, err := h.assignmentService.Update(userID, uint(id), title, description, subject, priority, dueDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update assignment"})
		return
	}

	c.JSON(http.StatusOK, assignment)
}

// DeleteAssignment deletes an assignment
// DELETE /api/v1/assignments/:id
func (h *APIHandler) DeleteAssignment(c *gin.Context) {
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	if err := h.assignmentService.Delete(userID, uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assignment deleted"})
}

// ToggleAssignment toggles the completion status of an assignment
// PATCH /api/v1/assignments/:id/toggle
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
