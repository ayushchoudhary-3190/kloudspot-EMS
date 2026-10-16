package service

import (
	"errors"
	"fmt"
	"strings"

	"EMS/internal/auth"
	"EMS/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ErrInvalidCredentials is returned when login fails.
var ErrInvalidCredentials = errors.New("invalid username or password")

// AdminRepository is the persistence contract required by AuthService.
type AdminRepository interface {
	FindByUsername(username string) (*models.Admin, error)
}

// AuthService handles administrator authentication.
type AuthService struct {
	repo      AdminRepository
	jwtSecret string
	expiryHrs int
}

// NewAuthService constructs a service.
func NewAuthService(repo AdminRepository, jwtSecret string, expiryHrs int) *AuthService {
	return &AuthService{repo: repo, jwtSecret: jwtSecret, expiryHrs: expiryHrs}
}

// Login verifies credentials and returns a signed JWT.
func (s *AuthService) Login(in models.LoginInput) (string, error) {
	username := strings.TrimSpace(in.Username)
	if username == "" || in.Password == "" {
		return "", ErrInvalidCredentials
	}

	admin, err := s.repo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("failed to load admin: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(in.Password)); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := auth.GenerateToken(admin.Username, s.jwtSecret, s.expiryHrs)
	if err != nil {
		return "", fmt.Errorf("failed to issue token: %w", err)
	}
	return token, nil
}
