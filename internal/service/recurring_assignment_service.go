package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"homework-manager/internal/models"
	"homework-manager/internal/repository"
)

var (
	ErrRecurringAssignmentNotFound = errors.New("recurring assignment not found")
	ErrRecurringUnauthorized       = errors.New("unauthorized")
	ErrInvalidRecurrenceType       = errors.New("invalid recurrence type")
	ErrInvalidEndType              = errors.New("invalid end type")
)

type RecurringAssignmentService struct {
	recurringRepo  *repository.RecurringAssignmentRepository
	assignmentRepo *repository.AssignmentRepository
}

func NewRecurringAssignmentService() *RecurringAssignmentService {
	return &RecurringAssignmentService{
		recurringRepo:  repository.NewRecurringAssignmentRepository(),
		assignmentRepo: repository.NewAssignmentRepository(),
	}
}

type CreateRecurringInput struct {
	Title                 string
	Description           string
	Subject               string
	Priority              string
	RecurrenceType        string
	RecurrenceInterval    int
	RecurrenceWeekday     *int
	RecurrenceDay         *int
	DueTime               string
	EndType               string
	EndCount              *int
	EndDate               *time.Time
	EditBehavior          string
	ReminderEnabled       bool
	ReminderOffset        *int
	UrgentReminderEnabled bool
	FirstDueDate          time.Time
}

func (s *RecurringAssignmentService) Create(userID uint, input CreateRecurringInput) (*models.RecurringAssignment, error) {
	if !isValidRecurrenceType(input.RecurrenceType) {
		return nil, ErrInvalidRecurrenceType
	}

	if !isValidEndType(input.EndType) {
		return nil, ErrInvalidEndType
	}

	if input.RecurrenceInterval < 1 {
		input.RecurrenceInterval = 1
	}
	if input.EditBehavior == "" {
		input.EditBehavior = models.EditBehaviorThisOnly
	}

	recurring := &models.RecurringAssignment{
		UserID:                userID,
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
		EndDate:               input.EndDate,
		EditBehavior:          input.EditBehavior,
		ReminderEnabled:       input.ReminderEnabled,
		ReminderOffset:        input.ReminderOffset,
		UrgentReminderEnabled: input.UrgentReminderEnabled,
		IsActive:              true,
		GeneratedCount:        0,
	}

	if err := s.recurringRepo.Create(recurring); err != nil {
		return nil, err
	}

	if err := s.generateAssignment(recurring, input.FirstDueDate); err != nil {
		return nil, err
	}

	return recurring, nil
}

func (s *RecurringAssignmentService) GetByID(userID, recurringID uint) (*models.RecurringAssignment, error) {
	recurring, err := s.recurringRepo.FindByID(recurringID)
	if err != nil {
		return nil, ErrRecurringAssignmentNotFound
	}

	if recurring.UserID != userID {
		return nil, ErrRecurringUnauthorized
	}

	return recurring, nil
}

func (s *RecurringAssignmentService) GetAllByUser(userID uint) ([]models.RecurringAssignment, error) {
	return s.recurringRepo.FindByUserID(userID)
}

func (s *RecurringAssignmentService) GetActiveByUser(userID uint) ([]models.RecurringAssignment, error) {
	return s.recurringRepo.FindActiveByUserID(userID)
}

type UpdateRecurringInput struct {
	Title                 string
	Description           string
	Subject               string
	Priority              string
	DueTime               string
	EditBehavior          string
	ReminderEnabled       bool
	ReminderOffset        *int
	UrgentReminderEnabled bool
}

func (s *RecurringAssignmentService) Update(userID, recurringID uint, input UpdateRecurringInput) (*models.RecurringAssignment, error) {
	recurring, err := s.GetByID(userID, recurringID)
	if err != nil {
		return nil, err
	}

	recurring.Title = input.Title
	recurring.Description = input.Description
	recurring.Subject = input.Subject
	recurring.Priority = input.Priority
	if input.DueTime != "" {
		recurring.DueTime = input.DueTime
	}
	if input.EditBehavior != "" {
		recurring.EditBehavior = input.EditBehavior
	}
	recurring.ReminderEnabled = input.ReminderEnabled
	recurring.ReminderOffset = input.ReminderOffset
	recurring.UrgentReminderEnabled = input.UrgentReminderEnabled

	if err := s.recurringRepo.Update(recurring); err != nil {
		return nil, err
	}

	return recurring, nil
}

func (s *RecurringAssignmentService) UpdateAssignmentWithBehavior(
	userID uint,
	assignment *models.Assignment,
	title, description, subject, priority string,
	dueDate time.Time,
	reminderEnabled bool,
	reminderAt *time.Time,
	urgentReminderEnabled bool,
	editBehavior string,
) error {
	if assignment.RecurringAssignmentID == nil {
		return s.updateSingleAssignment(assignment, title, description, subject, priority, dueDate, reminderEnabled, reminderAt, urgentReminderEnabled)
	}

	recurring, err := s.GetByID(userID, *assignment.RecurringAssignmentID)
	if err != nil {
		return s.updateSingleAssignment(assignment, title, description, subject, priority, dueDate, reminderEnabled, reminderAt, urgentReminderEnabled)
	}

	switch editBehavior {
	case models.EditBehaviorThisOnly:
		return s.updateSingleAssignment(assignment, title, description, subject, priority, dueDate, reminderEnabled, reminderAt, urgentReminderEnabled)

	case models.EditBehaviorThisAndFuture:
		if err := s.updateSingleAssignment(assignment, title, description, subject, priority, dueDate, reminderEnabled, reminderAt, urgentReminderEnabled); err != nil {
			return err
		}
		recurring.Title = title
		recurring.Description = description
		recurring.Subject = subject
		recurring.Priority = priority
		recurring.UrgentReminderEnabled = urgentReminderEnabled
		if err := s.recurringRepo.Update(recurring); err != nil {
			return err
		}
		return s.updateFutureAssignments(recurring.ID, assignment.DueDate, title, description, subject, priority, urgentReminderEnabled)

	case models.EditBehaviorAll:
		recurring.Title = title
		recurring.Description = description
		recurring.Subject = subject
		recurring.Priority = priority
		recurring.UrgentReminderEnabled = urgentReminderEnabled
		if err := s.recurringRepo.Update(recurring); err != nil {
			return err
		}
		return s.updateAllPendingAssignments(recurring.ID, title, description, subject, priority, urgentReminderEnabled)

	default:
		return s.updateSingleAssignment(assignment, title, description, subject, priority, dueDate, reminderEnabled, reminderAt, urgentReminderEnabled)
	}
}

func (s *RecurringAssignmentService) updateSingleAssignment(
	assignment *models.Assignment,
	title, description, subject, priority string,
	dueDate time.Time,
	reminderEnabled bool,
	reminderAt *time.Time,
	urgentReminderEnabled bool,
) error {
	assignment.Title = title
	assignment.Description = description
	assignment.Subject = subject
	assignment.Priority = priority
	assignment.DueDate = dueDate
	assignment.ReminderEnabled = reminderEnabled
	assignment.ReminderAt = reminderAt
	assignment.UrgentReminderEnabled = urgentReminderEnabled
	return s.assignmentRepo.Update(assignment)
}

func (s *RecurringAssignmentService) updateFutureAssignments(
	recurringID uint,
	fromDate time.Time,
	title, description, subject, priority string,
	urgentReminderEnabled bool,
) error {
	assignments, err := s.recurringRepo.GetFutureAssignmentsByRecurringID(recurringID, fromDate)
	if err != nil {
		return err
	}

	for _, a := range assignments {
		if a.IsCompleted {
			continue
		}
		a.Title = title
		a.Description = description
		a.Subject = subject
		a.Priority = priority
		a.UrgentReminderEnabled = urgentReminderEnabled
		if err := s.assignmentRepo.Update(&a); err != nil {
			return err
		}
	}
	return nil
}

func (s *RecurringAssignmentService) updateAllPendingAssignments(
	recurringID uint,
	title, description, subject, priority string,
	urgentReminderEnabled bool,
) error {
	assignments, err := s.recurringRepo.GetAssignmentsByRecurringID(recurringID)
	if err != nil {
		return err
	}

	for _, a := range assignments {
		if a.IsCompleted {
			continue
		}
		a.Title = title
		a.Description = description
		a.Subject = subject
		a.Priority = priority
		a.UrgentReminderEnabled = urgentReminderEnabled
		if err := s.assignmentRepo.Update(&a); err != nil {
			return err
		}
	}
	return nil
}

func (s *RecurringAssignmentService) Delete(userID, recurringID uint, deleteFutureAssignments bool) error {
	recurring, err := s.GetByID(userID, recurringID)
	if err != nil {
		return err
	}

	if deleteFutureAssignments {
		assignments, err := s.recurringRepo.GetFutureAssignmentsByRecurringID(recurringID, time.Now())
		if err != nil {
			return err
		}
		for _, a := range assignments {
			if !a.IsCompleted {
				s.assignmentRepo.Delete(a.ID)
			}
		}
	}

	return s.recurringRepo.Delete(recurring.ID)
}

func (s *RecurringAssignmentService) GenerateNextAssignments() error {
	recurrings, err := s.recurringRepo.FindDueForGeneration()
	if err != nil {
		return err
	}

	for _, recurring := range recurrings {
		pendingCount, err := s.recurringRepo.CountPendingByRecurringID(recurring.ID)
		if err != nil {
			continue
		}

		if pendingCount == 0 {
			latest, err := s.recurringRepo.GetLatestAssignmentByRecurringID(recurring.ID)
			if err != nil {
				continue
			}

			var nextDueDate time.Time
			if latest != nil {
				nextDueDate = recurring.CalculateNextDueDate(latest.DueDate)
			} else {
				nextDueDate = time.Now()
			}

			if nextDueDate.After(time.Now()) {
				s.generateAssignment(&recurring, nextDueDate)
			}
		}
	}

	return nil
}

func (s *RecurringAssignmentService) generateAssignment(recurring *models.RecurringAssignment, dueDate time.Time) error {
	if recurring.DueTime != "" {
		parts := strings.Split(recurring.DueTime, ":")
		if len(parts) == 2 {
			hour, _ := strconv.Atoi(parts[0])
			minute, _ := strconv.Atoi(parts[1])
			dueDate = time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), hour, minute, 0, 0, dueDate.Location())
		}
	}

	var reminderAt *time.Time
	if recurring.ReminderEnabled && recurring.ReminderOffset != nil {
		t := dueDate.Add(-time.Duration(*recurring.ReminderOffset) * time.Minute)
		reminderAt = &t
	}

	assignment := &models.Assignment{
		UserID:                userID(recurring.UserID),
		Title:                 recurring.Title,
		Description:           recurring.Description,
		Subject:               recurring.Subject,
		Priority:              recurring.Priority,
		DueDate:               dueDate,
		ReminderEnabled:       recurring.ReminderEnabled,
		ReminderAt:            reminderAt,
		UrgentReminderEnabled: recurring.UrgentReminderEnabled,
		RecurringAssignmentID: &recurring.ID,
	}

	if err := s.assignmentRepo.Create(assignment); err != nil {
		return err
	}

	recurring.GeneratedCount++
	return s.recurringRepo.Update(recurring)
}

func userID(id uint) uint {
	return id
}

func isValidRecurrenceType(t string) bool {
	switch t {
	case models.RecurrenceNone, models.RecurrenceDaily, models.RecurrenceWeekly, models.RecurrenceMonthly:
		return true
	}
	return false
}

func isValidEndType(t string) bool {
	switch t {
	case models.EndTypeNever, models.EndTypeCount, models.EndTypeDate:
		return true
	}
	return false
}

func GetRecurrenceTypeLabel(t string) string {
	switch t {
	case models.RecurrenceDaily:
		return "毎日"
	case models.RecurrenceWeekly:
		return "毎週"
	case models.RecurrenceMonthly:
		return "毎月"
	default:
		return "なし"
	}
}

func GetEndTypeLabel(t string) string {
	switch t {
	case models.EndTypeCount:
		return "回数指定"
	case models.EndTypeDate:
		return "終了日指定"
	default:
		return "無期限"
	}
}

func FormatRecurringSummary(recurring *models.RecurringAssignment) string {
	if recurring.RecurrenceType == models.RecurrenceNone {
		return ""
	}

	var parts []string

	typeLabel := GetRecurrenceTypeLabel(recurring.RecurrenceType)
	if recurring.RecurrenceInterval > 1 {
		switch recurring.RecurrenceType {
		case models.RecurrenceDaily:
			parts = append(parts, fmt.Sprintf("%d日ごと", recurring.RecurrenceInterval))
		case models.RecurrenceWeekly:
			parts = append(parts, fmt.Sprintf("%d週間ごと", recurring.RecurrenceInterval))
		case models.RecurrenceMonthly:
			parts = append(parts, fmt.Sprintf("%dヶ月ごと", recurring.RecurrenceInterval))
		}
	} else {
		parts = append(parts, typeLabel)
	}

	if recurring.RecurrenceType == models.RecurrenceWeekly && recurring.RecurrenceWeekday != nil {
		weekdays := []string{"日", "月", "火", "水", "木", "金", "土"}
		if *recurring.RecurrenceWeekday >= 0 && *recurring.RecurrenceWeekday < 7 {
			parts = append(parts, fmt.Sprintf("(%s曜日)", weekdays[*recurring.RecurrenceWeekday]))
		}
	}

	if recurring.RecurrenceType == models.RecurrenceMonthly && recurring.RecurrenceDay != nil {
		parts = append(parts, fmt.Sprintf("(%d日)", *recurring.RecurrenceDay))
	}

	switch recurring.EndType {
	case models.EndTypeCount:
		if recurring.EndCount != nil {
			parts = append(parts, fmt.Sprintf("/ %d回まで", *recurring.EndCount))
		}
	case models.EndTypeDate:
		if recurring.EndDate != nil {
			parts = append(parts, fmt.Sprintf("/ %sまで", recurring.EndDate.Format("2006/01/02")))
		}
	}

	return strings.Join(parts, " ")
}
