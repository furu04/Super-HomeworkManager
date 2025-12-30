package service

import (
	"errors"
	"time"

	"homework-manager/internal/models"
	"homework-manager/internal/repository"
)

var (
	ErrAssignmentNotFound = errors.New("assignment not found")
	ErrUnauthorized       = errors.New("unauthorized")
)

type PaginatedResult struct {
	Assignments []models.Assignment
	TotalCount  int64
	TotalPages  int
	CurrentPage int
	PageSize    int
}

type AssignmentService struct {
	assignmentRepo *repository.AssignmentRepository
}

func NewAssignmentService() *AssignmentService {
	return &AssignmentService{
		assignmentRepo: repository.NewAssignmentRepository(),
	}
}

func (s *AssignmentService) Create(userID uint, title, description, subject, priority string, dueDate time.Time) (*models.Assignment, error) {
	if priority == "" {
		priority = "medium"
	}
	assignment := &models.Assignment{
		UserID:      userID,
		Title:       title,
		Description: description,
		Subject:     subject,
		Priority:    priority,
		DueDate:     dueDate,
		IsCompleted: false,
	}

	if err := s.assignmentRepo.Create(assignment); err != nil {
		return nil, err
	}

	return assignment, nil
}

func (s *AssignmentService) GetByID(userID, assignmentID uint) (*models.Assignment, error) {
	assignment, err := s.assignmentRepo.FindByID(assignmentID)
	if err != nil {
		return nil, ErrAssignmentNotFound
	}

	if assignment.UserID != userID {
		return nil, ErrUnauthorized
	}

	return assignment, nil
}

func (s *AssignmentService) GetAllByUser(userID uint) ([]models.Assignment, error) {
	return s.assignmentRepo.FindByUserID(userID)
}

func (s *AssignmentService) GetPendingByUser(userID uint) ([]models.Assignment, error) {
	return s.assignmentRepo.FindPendingByUserID(userID, 0, 0)
}

func (s *AssignmentService) GetPendingByUserPaginated(userID uint, page, pageSize int) (*PaginatedResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	assignments, err := s.assignmentRepo.FindPendingByUserID(userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	totalCount, _ := s.assignmentRepo.CountPendingByUserID(userID)
	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))

	return &PaginatedResult{
		Assignments: assignments,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
	}, nil
}

func (s *AssignmentService) GetCompletedByUser(userID uint) ([]models.Assignment, error) {
	return s.assignmentRepo.FindCompletedByUserID(userID, 0, 0)
}

func (s *AssignmentService) GetCompletedByUserPaginated(userID uint, page, pageSize int) (*PaginatedResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	assignments, err := s.assignmentRepo.FindCompletedByUserID(userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	totalCount, _ := s.assignmentRepo.CountCompletedByUserID(userID)
	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))

	return &PaginatedResult{
		Assignments: assignments,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
	}, nil
}

func (s *AssignmentService) GetDueTodayByUser(userID uint) ([]models.Assignment, error) {
	return s.assignmentRepo.FindDueTodayByUserID(userID)
}

func (s *AssignmentService) GetDueThisWeekByUser(userID uint) ([]models.Assignment, error) {
	return s.assignmentRepo.FindDueThisWeekByUserID(userID)
}

func (s *AssignmentService) GetOverdueByUser(userID uint) ([]models.Assignment, error) {
	return s.assignmentRepo.FindOverdueByUserID(userID, 0, 0)
}

func (s *AssignmentService) GetOverdueByUserPaginated(userID uint, page, pageSize int) (*PaginatedResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	assignments, err := s.assignmentRepo.FindOverdueByUserID(userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	totalCount, _ := s.assignmentRepo.CountOverdueByUserID(userID)
	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))

	return &PaginatedResult{
		Assignments: assignments,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
	}, nil
}

func (s *AssignmentService) SearchAssignments(userID uint, query, priority, filter string, page, pageSize int) (*PaginatedResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	assignments, totalCount, err := s.assignmentRepo.Search(userID, query, priority, filter, page, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))

	return &PaginatedResult{
		Assignments: assignments,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
	}, nil
}

func (s *AssignmentService) Update(userID, assignmentID uint, title, description, subject, priority string, dueDate time.Time) (*models.Assignment, error) {
	assignment, err := s.GetByID(userID, assignmentID)
	if err != nil {
		return nil, err
	}

	assignment.Title = title
	assignment.Description = description
	assignment.Subject = subject
	assignment.Priority = priority
	assignment.DueDate = dueDate

	if err := s.assignmentRepo.Update(assignment); err != nil {
		return nil, err
	}

	return assignment, nil
}

func (s *AssignmentService) ToggleComplete(userID, assignmentID uint) (*models.Assignment, error) {
	assignment, err := s.GetByID(userID, assignmentID)
	if err != nil {
		return nil, err
	}

	assignment.IsCompleted = !assignment.IsCompleted
	if assignment.IsCompleted {
		now := time.Now()
		assignment.CompletedAt = &now
	} else {
		assignment.CompletedAt = nil
	}

	if err := s.assignmentRepo.Update(assignment); err != nil {
		return nil, err
	}

	return assignment, nil
}

func (s *AssignmentService) Delete(userID, assignmentID uint) error {
	assignment, err := s.GetByID(userID, assignmentID)
	if err != nil {
		return err
	}

	return s.assignmentRepo.Delete(assignment.ID)
}

func (s *AssignmentService) GetSubjectsByUser(userID uint) ([]string, error) {
	return s.assignmentRepo.GetSubjectsByUserID(userID)
}

type DashboardStats struct {
	TotalPending  int64
	DueToday      int
	DueThisWeek   int
	Overdue       int
	Subjects      []string
}

func (s *AssignmentService) GetDashboardStats(userID uint) (*DashboardStats, error) {
	pending, _ := s.assignmentRepo.CountPendingByUserID(userID)
	dueToday, _ := s.assignmentRepo.FindDueTodayByUserID(userID)
	dueThisWeek, _ := s.assignmentRepo.FindDueThisWeekByUserID(userID)
	overdueCount, _ := s.assignmentRepo.CountOverdueByUserID(userID)
	subjects, _ := s.assignmentRepo.GetSubjectsByUserID(userID)

	return &DashboardStats{
		TotalPending:  pending,
		DueToday:      len(dueToday),
		DueThisWeek:   len(dueThisWeek),
		Overdue:       int(overdueCount),
		Subjects:      subjects,
	}, nil
}
