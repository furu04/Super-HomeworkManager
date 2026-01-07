package repository

import (
	"time"

	"homework-manager/internal/database"
	"homework-manager/internal/models"

	"gorm.io/gorm"
)

type RecurringAssignmentRepository struct {
	db *gorm.DB
}

func NewRecurringAssignmentRepository() *RecurringAssignmentRepository {
	return &RecurringAssignmentRepository{db: database.GetDB()}
}

func (r *RecurringAssignmentRepository) Create(recurring *models.RecurringAssignment) error {
	return r.db.Create(recurring).Error
}

func (r *RecurringAssignmentRepository) FindByID(id uint) (*models.RecurringAssignment, error) {
	var recurring models.RecurringAssignment
	err := r.db.First(&recurring, id).Error
	if err != nil {
		return nil, err
	}
	return &recurring, nil
}

func (r *RecurringAssignmentRepository) FindByUserID(userID uint) ([]models.RecurringAssignment, error) {
	var recurrings []models.RecurringAssignment
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&recurrings).Error
	return recurrings, err
}

func (r *RecurringAssignmentRepository) FindActiveByUserID(userID uint) ([]models.RecurringAssignment, error) {
	var recurrings []models.RecurringAssignment
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).Order("created_at DESC").Find(&recurrings).Error
	return recurrings, err
}

func (r *RecurringAssignmentRepository) Update(recurring *models.RecurringAssignment) error {
	return r.db.Save(recurring).Error
}

func (r *RecurringAssignmentRepository) Delete(id uint) error {
	return r.db.Delete(&models.RecurringAssignment{}, id).Error
}

func (r *RecurringAssignmentRepository) FindDueForGeneration() ([]models.RecurringAssignment, error) {
	var recurrings []models.RecurringAssignment

	err := r.db.Where("is_active = ? AND recurrence_type != ?", true, models.RecurrenceNone).
		Find(&recurrings).Error

	if err != nil {
		return nil, err
	}

	var result []models.RecurringAssignment
	now := time.Now()
	for _, rec := range recurrings {
		shouldGenerate := true

		switch rec.EndType {
		case models.EndTypeCount:
			if rec.EndCount != nil && rec.GeneratedCount >= *rec.EndCount {
				shouldGenerate = false
			}
		case models.EndTypeDate:
			if rec.EndDate != nil && now.After(*rec.EndDate) {
				shouldGenerate = false
			}
		}

		if shouldGenerate {
			result = append(result, rec)
		}
	}

	return result, nil
}

func (r *RecurringAssignmentRepository) GetLatestAssignmentByRecurringID(recurringID uint) (*models.Assignment, error) {
	var assignment models.Assignment
	err := r.db.Where("recurring_assignment_id = ?", recurringID).
		Order("due_date DESC").
		First(&assignment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &assignment, nil
}

func (r *RecurringAssignmentRepository) GetAssignmentsByRecurringID(recurringID uint) ([]models.Assignment, error) {
	var assignments []models.Assignment
	err := r.db.Where("recurring_assignment_id = ?", recurringID).
		Order("due_date ASC").
		Find(&assignments).Error
	return assignments, err
}

func (r *RecurringAssignmentRepository) GetFutureAssignmentsByRecurringID(recurringID uint, fromDate time.Time) ([]models.Assignment, error) {
	var assignments []models.Assignment
	err := r.db.Where("recurring_assignment_id = ? AND due_date >= ?", recurringID, fromDate).
		Order("due_date ASC").
		Find(&assignments).Error
	return assignments, err
}

func (r *RecurringAssignmentRepository) CountPendingByRecurringID(recurringID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Assignment{}).
		Where("recurring_assignment_id = ? AND is_completed = ?", recurringID, false).
		Count(&count).Error
	return count, err
}
