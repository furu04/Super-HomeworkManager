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
	ErrLeadDaysTooLarge            = errors.New("generation_lead_days must be less than the recurrence interval")
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

type CreateRecurringAssignmentInput struct {
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
	GenerationLeadDays    int
	GenerationLeadTime    string
	FirstDueDate          time.Time
}

func (s *RecurringAssignmentService) Create(userID uint, input CreateRecurringAssignmentInput) (*models.RecurringAssignment, error) {
	if !isValidRecurrenceType(input.RecurrenceType) {
		return nil, ErrInvalidRecurrenceType
	}

	if !isValidEndType(input.EndType) {
		return nil, ErrInvalidEndType
	}

	if input.RecurrenceInterval < 1 {
		input.RecurrenceInterval = 1
	}

	if input.GenerationLeadDays > 0 {
		maxLead := maxGenerationLeadDays(input.RecurrenceType, input.RecurrenceInterval)
		if input.GenerationLeadDays > maxLead {
			return nil, ErrLeadDaysTooLarge
		}
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
		GenerationLeadDays:    input.GenerationLeadDays,
		GenerationLeadTime:    input.GenerationLeadTime,
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
	Title                 *string
	Description           *string
	Subject               *string
	Priority              *string
	RecurrenceType        *string
	RecurrenceInterval    *int
	RecurrenceWeekday     *int
	RecurrenceDay         *int
	DueTime               *string
	EndType               *string
	EndCount              *int
	EndDate               *time.Time
	EditBehavior          string
	ReminderEnabled       *bool
	ReminderOffset        *int
	UrgentReminderEnabled *bool
	GenerationLeadDays    *int
	GenerationLeadTime    *string
}

func (s *RecurringAssignmentService) Update(userID, recurringID uint, input UpdateRecurringInput) (*models.RecurringAssignment, error) {
	recurring, err := s.GetByID(userID, recurringID)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		recurring.Title = *input.Title
	}
	if input.Description != nil {
		recurring.Description = *input.Description
	}
	if input.Subject != nil {
		recurring.Subject = *input.Subject
	}
	if input.Priority != nil {
		recurring.Priority = *input.Priority
	}
	if input.DueTime != nil {
		recurring.DueTime = *input.DueTime
	}
	if input.EditBehavior != "" {
		recurring.EditBehavior = input.EditBehavior
	}
	if input.ReminderEnabled != nil {
		recurring.ReminderEnabled = *input.ReminderEnabled
	}
	if input.ReminderOffset != nil {
		recurring.ReminderOffset = input.ReminderOffset
	}
	if input.UrgentReminderEnabled != nil {
		recurring.UrgentReminderEnabled = *input.UrgentReminderEnabled
	}

	if input.RecurrenceType != nil && *input.RecurrenceType != "" && isValidRecurrenceType(*input.RecurrenceType) {
		recurring.RecurrenceType = *input.RecurrenceType
	}
	if input.RecurrenceInterval != nil && *input.RecurrenceInterval > 0 {
		recurring.RecurrenceInterval = *input.RecurrenceInterval
	}
	if input.RecurrenceWeekday != nil {
		recurring.RecurrenceWeekday = input.RecurrenceWeekday
	}
	if input.RecurrenceDay != nil {
		recurring.RecurrenceDay = input.RecurrenceDay
	}

	if input.EndType != nil && isValidEndType(*input.EndType) {
		recurring.EndType = *input.EndType
	}
	if input.EndCount != nil {
		recurring.EndCount = input.EndCount
	}
	if input.EndDate != nil {
		recurring.EndDate = input.EndDate
	}
	if input.GenerationLeadDays != nil && *input.GenerationLeadDays >= 0 {
		if *input.GenerationLeadDays > 0 {
			maxLead := maxGenerationLeadDays(recurring.RecurrenceType, recurring.RecurrenceInterval)
			if *input.GenerationLeadDays > maxLead {
				return nil, ErrLeadDaysTooLarge
			}
		}
		recurring.GenerationLeadDays = *input.GenerationLeadDays
	}
	if input.GenerationLeadTime != nil {
		recurring.GenerationLeadTime = *input.GenerationLeadTime
	}

	if err := s.recurringRepo.Update(recurring); err != nil {
		return nil, err
	}

	return recurring, nil
}

func (s *RecurringAssignmentService) SetActive(userID, recurringID uint, isActive bool) error {
	recurring, err := s.GetByID(userID, recurringID)
	if err != nil {
		return err
	}

	recurring.IsActive = isActive
	return s.recurringRepo.Update(recurring)
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
		if err := s.generateNextIfPending(&recurring); err != nil {
			continue
		}
	}

	return nil
}

func (s *RecurringAssignmentService) TriggerForRecurring(recurringID uint) error {
	recurring, err := s.recurringRepo.FindByID(recurringID)
	if err != nil {
		return nil
	}
	return s.generateNextIfPending(recurring)
}

func (s *RecurringAssignmentService) generateNextIfPending(recurring *models.RecurringAssignment) error {
	if !recurring.ShouldGenerateNext() {
		return nil
	}

	latest, err := s.recurringRepo.GetLatestAssignmentByRecurringID(recurring.ID)
	if err != nil {
		return err
	}
	if latest == nil {
		return nil
	}

	if !latest.IsCompleted && !latest.IsOverdue() {
		return nil
	}

	nextDueDate := recurring.CalculateNextDueDate(latest.DueDate)

	if recurring.GenerationLeadDays > 0 {
		generationAt := nextDueDate.AddDate(0, 0, -recurring.GenerationLeadDays)
		if recurring.GenerationLeadTime != "" {
			parts := strings.Split(recurring.GenerationLeadTime, ":")
			if len(parts) == 2 {
				hour, _ := strconv.Atoi(parts[0])
				min, _ := strconv.Atoi(parts[1])
				generationAt = time.Date(generationAt.Year(), generationAt.Month(), generationAt.Day(), hour, min, 0, 0, generationAt.Location())
			}
		} else {
			generationAt = time.Date(generationAt.Year(), generationAt.Month(), generationAt.Day(), 0, 0, 0, 0, generationAt.Location())
		}
		if time.Now().Before(generationAt) {
			return nil
		}
	}

	return s.generateAssignment(recurring, nextDueDate)
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

	existing, err := s.assignmentRepo.FindByRecurringAndDue(recurring.ID, dueDate)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	var reminderAt *time.Time
	if recurring.ReminderEnabled && recurring.ReminderOffset != nil {
		t := dueDate.Add(-time.Duration(*recurring.ReminderOffset) * time.Minute)
		reminderAt = &t
	}

	assignment := &models.Assignment{
		UserID:                recurring.UserID,
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


func maxGenerationLeadDays(recurrenceType string, interval int) int {
	switch recurrenceType {
	case models.RecurrenceDaily:
		return interval
	case models.RecurrenceWeekly:
		return interval * 7
	case models.RecurrenceMonthly:
		return interval * 28
	}
	return 0
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
