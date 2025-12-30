package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"homework-manager/internal/database"
	"homework-manager/internal/models"
)

type APIKeyService struct{}

func NewAPIKeyService() *APIKeyService {
	return &APIKeyService{}
}


func (s *APIKeyService) generateRandomKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "hm_" + hex.EncodeToString(bytes), nil
}

func (s *APIKeyService) hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

func (s *APIKeyService) CreateAPIKey(userID uint, name string) (string, *models.APIKey, error) {
	if name == "" {
		return "", nil, errors.New("キー名を入力してください")
	}

	plainKey, err := s.generateRandomKey()
	if err != nil {
		return "", nil, errors.New("キーの生成に失敗しました")
	}

	apiKey := &models.APIKey{
		UserID:  userID,
		Name:    name,
		KeyHash: s.hashKey(plainKey),
	}

	if err := database.GetDB().Create(apiKey).Error; err != nil {
		return "", nil, errors.New("キーの保存に失敗しました")
	}

	return plainKey, apiKey, nil
}

func (s *APIKeyService) ValidateAPIKey(plainKey string) (uint, error) {
	hash := s.hashKey(plainKey)

	var apiKey models.APIKey
	if err := database.GetDB().Where("key_hash = ?", hash).First(&apiKey).Error; err != nil {
		return 0, errors.New("無効なAPIキーです")
	}

	now := time.Now()
	database.GetDB().Model(&apiKey).Update("last_used", now)

	return apiKey.UserID, nil
}

func (s *APIKeyService) GetAllAPIKeys() ([]models.APIKey, error) {
	var keys []models.APIKey
	err := database.GetDB().Preload("User").Order("created_at desc").Find(&keys).Error
	return keys, err
}

func (s *APIKeyService) GetAPIKeysByUser(userID uint) ([]models.APIKey, error) {
	var keys []models.APIKey
	err := database.GetDB().Where("user_id = ?", userID).Order("created_at desc").Find(&keys).Error
	return keys, err
}

func (s *APIKeyService) DeleteAPIKey(id uint) error {
	result := database.GetDB().Delete(&models.APIKey{}, id)
	if result.RowsAffected == 0 {
		return errors.New("APIキーが見つかりません")
	}
	return result.Error
}
