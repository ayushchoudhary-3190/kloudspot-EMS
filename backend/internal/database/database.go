package database

import (
	"log"

	"EMS/internal/config"
	"EMS/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens the SQLite database and runs migrations.
func Connect(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.Employee{}, &models.Admin{}); err != nil {
		return nil, err
	}
	return db, nil
}

// SeedAdmin creates the default administrator account if it does not exist.
func SeedAdmin(db *gorm.DB, cfg config.Config) error {
	var count int64
	if err := db.Model(&models.Admin{}).Where("username = ?", cfg.AdminUsername).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := models.Admin{Username: cfg.AdminUsername, PasswordHash: string(hash)}
	if err := db.Create(&admin).Error; err != nil {
		return err
	}
	log.Printf("seeded admin user %q", cfg.AdminUsername)
	return nil
}
