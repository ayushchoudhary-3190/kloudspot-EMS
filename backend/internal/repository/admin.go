package repository

import (
	"EMS/internal/models"

	"gorm.io/gorm"
)

// AdminRepository persists administrator accounts.
type AdminRepository struct {
	DB *gorm.DB
}

// NewAdminRepository constructs a repository.
func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{DB: db}
}

// FindByUsername returns an admin by username or gorm.ErrRecordNotFound.
func (r *AdminRepository) FindByUsername(username string) (*models.Admin, error) {
	var a models.Admin
	if err := r.DB.Where("username = ?", username).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}
