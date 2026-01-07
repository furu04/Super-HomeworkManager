package repository

import (
	"time"

	"homework-manager/internal/database"
	"homework-manager/internal/models"

	"gorm.io/gorm"
)

type AssignmentRepository struct {
	db *gorm.DB
}

func NewAssignmentRepository() *AssignmentRepository {
	return &AssignmentRepository{db: database.GetDB()}
}

func (r *AssignmentRepository) Create(assignment *models.Assignment) error {
	return r.db.Create(assignment).Error
}

func (r *AssignmentRepository) FindByID(id uint) (*models.Assignment, error) {
	var assignment models.Assignment
	err := r.db.First(&assignment, id).Error
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

func (r *AssignmentRepository) FindByUserID(userID uint) ([]models.Assignment, error) {
	var assignments []models.Assignment
	err := r.db.Where("user_id = ?", userID).Order("due_date ASC").Find(&assignments).Error
	return assignments, err
}

func (r *AssignmentRepository) FindPendingByUserID(userID uint, limit, offset int) ([]models.Assignment, error) {
	var assignments []models.Assignment
	query := r.db.Where("user_id = ? AND is_completed = ?", userID, false).
		Order("due_date ASC")
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}
	err := query.Find(&assignments).Error
	return assignments, err
}

func (r *AssignmentRepository) FindCompletedByUserID(userID uint, limit, offset int) ([]models.Assignment, error) {
	var assignments []models.Assignment
	query := r.db.Where("user_id = ? AND is_completed = ?", userID, true).
		Order("completed_at DESC")
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}
	err := query.Find(&assignments).Error
	return assignments, err
}

func (r *AssignmentRepository) FindDueTodayByUserID(userID uint) ([]models.Assignment, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)

	var assignments []models.Assignment
	err := r.db.Where("user_id = ? AND is_completed = ? AND due_date >= ? AND due_date < ?",
		userID, false, startOfDay, endOfDay).
		Order("due_date ASC").Find(&assignments).Error
	return assignments, err
}

func (r *AssignmentRepository) FindDueThisWeekByUserID(userID uint) ([]models.Assignment, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekLater := startOfDay.AddDate(0, 0, 7)

	var assignments []models.Assignment
	err := r.db.Where("user_id = ? AND is_completed = ? AND due_date >= ? AND due_date < ?",
		userID, false, startOfDay, weekLater).
		Order("due_date ASC").Find(&assignments).Error
	return assignments, err
}

func (r *AssignmentRepository) FindOverdueByUserID(userID uint, limit, offset int) ([]models.Assignment, error) {
	now := time.Now()

	var assignments []models.Assignment
	query := r.db.Where("user_id = ? AND is_completed = ? AND due_date < ?",
		userID, false, now).
		Order("due_date ASC")
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}
	err := query.Find(&assignments).Error
	return assignments, err
}

func (r *AssignmentRepository) Update(assignment *models.Assignment) error {
	return r.db.Save(assignment).Error
}

func (r *AssignmentRepository) Delete(id uint) error {
	return r.db.Delete(&models.Assignment{}, id).Error
}

func (r *AssignmentRepository) CountByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Assignment{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *AssignmentRepository) CountPendingByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Assignment{}).
		Where("user_id = ? AND is_completed = ?", userID, false).Count(&count).Error
	return count, err
}

func (r *AssignmentRepository) GetSubjectsByUserID(userID uint) ([]string, error) {
	var subjects []string
	err := r.db.Model(&models.Assignment{}).
		Where("user_id = ? AND subject != ''", userID).
		Distinct("subject").
		Pluck("subject", &subjects).Error
	return subjects, err
}

func (r *AssignmentRepository) CountCompletedByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Assignment{}).
		Where("user_id = ? AND is_completed = ?", userID, true).Count(&count).Error
	return count, err
}

func (r *AssignmentRepository) Search(userID uint, queryStr, priority, filter string, page, pageSize int) ([]models.Assignment, int64, error) {
	var assignments []models.Assignment
	var totalCount int64

	dbQuery := r.db.Model(&models.Assignment{}).Where("user_id = ?", userID)

	if queryStr != "" {
		dbQuery = dbQuery.Where("title LIKE ? OR description LIKE ?", "%"+queryStr+"%", "%"+queryStr+"%")
	}

	if priority != "" {
		dbQuery = dbQuery.Where("priority = ?", priority)
	}

	now := time.Now()
	switch filter {
	case "completed":
		dbQuery = dbQuery.Where("is_completed = ?", true)
	case "overdue":
		dbQuery = dbQuery.Where("is_completed = ? AND due_date < ?", false, now)
	default: // pending
		dbQuery = dbQuery.Where("is_completed = ?", false)
	}

	if err := dbQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if filter == "completed" {
		dbQuery = dbQuery.Order("completed_at DESC")
	} else {
		dbQuery = dbQuery.Order("due_date ASC")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	err := dbQuery.Limit(pageSize).Offset(offset).Find(&assignments).Error
	return assignments, totalCount, err
}

func (r *AssignmentRepository) CountOverdueByUserID(userID uint) (int64, error) {
	var count int64
	now := time.Now()
	err := r.db.Model(&models.Assignment{}).
		Where("user_id = ? AND is_completed = ? AND due_date < ?", userID, false, now).Count(&count).Error
	return count, err
}

type StatisticsFilter struct {
	Subject         string
	From            *time.Time
	To              *time.Time
	IncludeArchived bool
}

type AssignmentStatistics struct {
	Total                  int64
	Completed              int64
	Pending                int64
	Overdue                int64
	CompletedOnTime        int64
	OnTimeCompletionRate   float64
}

type SubjectStatistics struct {
	Subject              string
	Total                int64
	Completed            int64
	Pending              int64
	Overdue              int64
	CompletedOnTime      int64
	OnTimeCompletionRate float64
}

func (r *AssignmentRepository) GetStatistics(userID uint, filter StatisticsFilter) (*AssignmentStatistics, error) {
	now := time.Now()
	stats := &AssignmentStatistics{}
	baseQuery := r.db.Model(&models.Assignment{}).Where("user_id = ?", userID)
	
	if filter.Subject != "" {
		baseQuery = baseQuery.Where("subject = ?", filter.Subject)
	}
	
	if filter.From != nil {
		baseQuery = baseQuery.Where("created_at >= ?", *filter.From)
	}
	if filter.To != nil {
		toEnd := filter.To.AddDate(0, 0, 1)
		baseQuery = baseQuery.Where("created_at < ?", toEnd)
	}
	if !filter.IncludeArchived {
		baseQuery = baseQuery.Where("is_archived = ?", false)
	}

	if err := baseQuery.Count(&stats.Total).Error; err != nil {
		return nil, err
	}
	completedQuery := baseQuery.Session(&gorm.Session{})
	if err := completedQuery.Where("is_completed = ?", true).Count(&stats.Completed).Error; err != nil {
		return nil, err
	}

	pendingQuery := baseQuery.Session(&gorm.Session{})
	if err := pendingQuery.Where("is_completed = ?", false).Count(&stats.Pending).Error; err != nil {
		return nil, err
	}
	overdueQuery := baseQuery.Session(&gorm.Session{})
	if err := overdueQuery.Where("is_completed = ? AND due_date < ?", false, now).Count(&stats.Overdue).Error; err != nil {
		return nil, err
	}

	onTimeQuery := baseQuery.Session(&gorm.Session{})
	if err := onTimeQuery.Where("is_completed = ? AND completed_at <= due_date", true).Count(&stats.CompletedOnTime).Error; err != nil {
		return nil, err
	}

	if stats.Completed > 0 {
		stats.OnTimeCompletionRate = float64(stats.CompletedOnTime) / float64(stats.Completed) * 100
	}

	return stats, nil
}

func (r *AssignmentRepository) GetStatisticsBySubjects(userID uint, filter StatisticsFilter) ([]SubjectStatistics, error) {
	now := time.Now()
	subjects, err := r.GetSubjectsByUserID(userID)
	if err != nil {
		return nil, err
	}

	var results []SubjectStatistics
	for _, subject := range subjects {
		subjectFilter := StatisticsFilter{
			Subject: subject,
			From:    filter.From,
			To:      filter.To,
		}
		stats, err := r.GetStatistics(userID, subjectFilter)
		if err != nil {
			return nil, err
		}

		overdueQuery := r.db.Model(&models.Assignment{}).
			Where("user_id = ? AND subject = ? AND is_completed = ? AND due_date < ?", userID, subject, false, now)
		if filter.From != nil {
			overdueQuery = overdueQuery.Where("created_at >= ?", *filter.From)
		}
		if filter.To != nil {
			toEnd := filter.To.AddDate(0, 0, 1)
			overdueQuery = overdueQuery.Where("created_at < ?", toEnd)
		}
		var overdueCount int64
		overdueQuery.Count(&overdueCount)

		results = append(results, SubjectStatistics{
			Subject:              subject,
			Total:                stats.Total,
			Completed:            stats.Completed,
			Pending:              stats.Pending,
			Overdue:              overdueCount,
			CompletedOnTime:      stats.CompletedOnTime,
			OnTimeCompletionRate: stats.OnTimeCompletionRate,
		})
	}

	return results, nil
}

func (r *AssignmentRepository) ArchiveBySubject(userID uint, subject string) error {
	return r.db.Model(&models.Assignment{}).
		Where("user_id = ? AND subject = ?", userID, subject).
		Update("is_archived", true).Error
}

func (r *AssignmentRepository) UnarchiveBySubject(userID uint, subject string) error {
	return r.db.Model(&models.Assignment{}).
		Where("user_id = ? AND subject = ?", userID, subject).
		Update("is_archived", false).Error
}

func (r *AssignmentRepository) GetArchivedSubjects(userID uint) ([]string, error) {
	var subjects []string
	err := r.db.Model(&models.Assignment{}).
		Where("user_id = ? AND is_archived = ? AND subject != ''", userID, true).
		Distinct("subject").
		Pluck("subject", &subjects).Error
	return subjects, err
}

func (r *AssignmentRepository) GetSubjectsByUserIDWithArchived(userID uint, includeArchived bool) ([]string, error) {
	var subjects []string
	query := r.db.Model(&models.Assignment{}).
		Where("user_id = ? AND subject != ''", userID)
	if !includeArchived {
		query = query.Where("is_archived = ?", false)
	}
	err := query.Distinct("subject").Pluck("subject", &subjects).Error
	return subjects, err
}

