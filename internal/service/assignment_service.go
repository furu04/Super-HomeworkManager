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

func (s *AssignmentService) Create(userID uint, title, description, subject, priority string, dueDate time.Time, reminderEnabled bool, reminderAt *time.Time, urgentReminderEnabled bool) (*models.Assignment, error) {
	if priority == "" {
		priority = "medium"
	}
	assignment := &models.Assignment{
		UserID:                userID,
		Title:                 title,
		Description:           description,
		Subject:               subject,
		Priority:              priority,
		DueDate:               dueDate,
		IsCompleted:           false,
		ReminderEnabled:       reminderEnabled,
		ReminderAt:            reminderAt,
		ReminderSent:          false,
		UrgentReminderEnabled: urgentReminderEnabled,
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

	assignments, totalCount, err := s.assignmentRepo.SearchWithPreload(userID, query, priority, filter, page, pageSize)
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

func (s *AssignmentService) Update(userID, assignmentID uint, title, description, subject, priority string, dueDate time.Time, reminderEnabled bool, reminderAt *time.Time, urgentReminderEnabled bool) (*models.Assignment, error) {
	assignment, err := s.GetByID(userID, assignmentID)
	if err != nil {
		return nil, err
	}

	assignment.Title = title
	assignment.Description = description
	assignment.Subject = subject
	assignment.Priority = priority
	assignment.DueDate = dueDate
	assignment.ReminderEnabled = reminderEnabled
	assignment.ReminderAt = reminderAt
	assignment.UrgentReminderEnabled = urgentReminderEnabled
	if reminderEnabled && reminderAt != nil {
		assignment.ReminderSent = false
	}

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
	TotalPending int64
	DueToday     int
	DueThisWeek  int
	Overdue      int
	Subjects     []string
}

func (s *AssignmentService) GetDashboardStats(userID uint) (*DashboardStats, error) {
	pending, _ := s.assignmentRepo.CountPendingByUserID(userID)
	dueToday, _ := s.assignmentRepo.FindDueTodayByUserID(userID)
	dueThisWeek, _ := s.assignmentRepo.FindDueThisWeekByUserID(userID)
	overdueCount, _ := s.assignmentRepo.CountOverdueByUserID(userID)
	subjects, _ := s.assignmentRepo.GetSubjectsByUserID(userID)

	return &DashboardStats{
		TotalPending: pending,
		DueToday:     len(dueToday),
		DueThisWeek:  len(dueThisWeek),
		Overdue:      int(overdueCount),
		Subjects:     subjects,
	}, nil
}

type StatisticsFilter struct {
	Subject         string
	From            *time.Time
	To              *time.Time
	IncludeArchived bool
}

type SubjectStats struct {
	Subject              string  `json:"subject"`
	Total                int64   `json:"total"`
	Completed            int64   `json:"completed"`
	Pending              int64   `json:"pending"`
	Overdue              int64   `json:"overdue"`
	OnTimeCompletionRate float64 `json:"on_time_completion_rate"`
	IsArchived           bool    `json:"is_archived,omitempty"`
}

type StatisticsSummary struct {
	TotalAssignments     int64          `json:"total_assignments"`
	CompletedAssignments int64          `json:"completed_assignments"`
	PendingAssignments   int64          `json:"pending_assignments"`
	OverdueAssignments   int64          `json:"overdue_assignments"`
	OnTimeCompletionRate float64        `json:"on_time_completion_rate"`
	Filter               *FilterInfo    `json:"filter,omitempty"`
	Subjects             []SubjectStats `json:"subjects,omitempty"`
}

type FilterInfo struct {
	Subject         *string `json:"subject"`
	From            *string `json:"from"`
	To              *string `json:"to"`
	IncludeArchived bool    `json:"include_archived"`
}

func (s *AssignmentService) GetStatistics(userID uint, filter StatisticsFilter) (*StatisticsSummary, error) {
	repoFilter := repository.StatisticsFilter{
		Subject:         filter.Subject,
		From:            filter.From,
		To:              filter.To,
		IncludeArchived: filter.IncludeArchived,
	}

	stats, err := s.assignmentRepo.GetStatistics(userID, repoFilter)
	if err != nil {
		return nil, err
	}

	summary := &StatisticsSummary{
		TotalAssignments:     stats.Total,
		CompletedAssignments: stats.Completed,
		PendingAssignments:   stats.Pending,
		OverdueAssignments:   stats.Overdue,
		OnTimeCompletionRate: stats.OnTimeCompletionRate,
	}

	filterInfo := &FilterInfo{}
	hasFilter := false
	if filter.Subject != "" {
		filterInfo.Subject = &filter.Subject
		hasFilter = true
	}
	if filter.From != nil {
		fromStr := filter.From.Format("2006-01-02")
		filterInfo.From = &fromStr
		hasFilter = true
	}
	if filter.To != nil {
		toStr := filter.To.Format("2006-01-02")
		filterInfo.To = &toStr
		hasFilter = true
	}
	filterInfo.IncludeArchived = filter.IncludeArchived
	if filter.IncludeArchived {
		hasFilter = true
	}
	if hasFilter {
		summary.Filter = filterInfo
	}

	if filter.Subject == "" {
		subjectStats, err := s.assignmentRepo.GetStatisticsBySubjects(userID, repoFilter)
		if err != nil {
			return nil, err
		}

		for _, ss := range subjectStats {
			summary.Subjects = append(summary.Subjects, SubjectStats{
				Subject:              ss.Subject,
				Total:                ss.Total,
				Completed:            ss.Completed,
				Pending:              ss.Pending,
				Overdue:              ss.Overdue,
				OnTimeCompletionRate: ss.OnTimeCompletionRate,
			})
		}
	}

	return summary, nil
}

func (s *AssignmentService) ArchiveSubject(userID uint, subject string) error {
	return s.assignmentRepo.ArchiveBySubject(userID, subject)
}
func (s *AssignmentService) UnarchiveSubject(userID uint, subject string) error {
	return s.assignmentRepo.UnarchiveBySubject(userID, subject)
}

func (s *AssignmentService) GetSubjectsWithArchived(userID uint, includeArchived bool) ([]string, error) {
	return s.assignmentRepo.GetSubjectsByUserIDWithArchived(userID, includeArchived)
}

func (s *AssignmentService) GetArchivedSubjects(userID uint) ([]string, error) {
	return s.assignmentRepo.GetArchivedSubjects(userID)
}
