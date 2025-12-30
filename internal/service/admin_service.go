package service

import (
	"errors"

	"homework-manager/internal/models"
	"homework-manager/internal/repository"
)

var (
	ErrCannotDeleteSelf = errors.New("cannot delete yourself")
	ErrCannotChangeSelfRole = errors.New("cannot change your own role")
)

type AdminService struct {
	userRepo *repository.UserRepository
}

func NewAdminService() *AdminService {
	return &AdminService{
		userRepo: repository.NewUserRepository(),
	}
}

func (s *AdminService) GetAllUsers() ([]models.User, error) {
	return s.userRepo.FindAll()
}

func (s *AdminService) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *AdminService) DeleteUser(adminID, targetID uint) error {
	if adminID == targetID {
		return ErrCannotDeleteSelf
	}

	_, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return ErrUserNotFound
	}

	return s.userRepo.Delete(targetID)
}

func (s *AdminService) ChangeRole(adminID, targetID uint, newRole string) error {
	if adminID == targetID {
		return ErrCannotChangeSelfRole
	}

	if newRole != "admin" && newRole != "user" {
		return errors.New("invalid role")
	}

	user, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return ErrUserNotFound
	}

	user.Role = newRole
	return s.userRepo.Update(user)
}
