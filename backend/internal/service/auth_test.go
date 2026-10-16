package service

import (
	"errors"
	"testing"

	"EMS/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type mockAdminRepository struct {
	mock.Mock
}

func (m *mockAdminRepository) FindByUsername(username string) (*models.Admin, error) {
	args := m.Called(username)
	if a := args.Get(0); a != nil {
		return a.(*models.Admin), args.Error(1)
	}
	return nil, args.Error(1)
}

func hashedAdmin(t *testing.T, username, password string) *models.Admin {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	return &models.Admin{ID: 1, Username: username, PasswordHash: string(hash)}
}

func TestLogin_Success(t *testing.T) {
	repo := new(mockAdminRepository)
	repo.On("FindByUsername", "admin").Return(hashedAdmin(t, "admin", "admin123"), nil)

	svc := NewAuthService(repo, "test-secret", 1)
	token, err := svc.Login(models.LoginInput{Username: "admin", Password: "admin123"})

	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestLogin_TrimsUsername(t *testing.T) {
	repo := new(mockAdminRepository)
	repo.On("FindByUsername", "admin").Return(hashedAdmin(t, "admin", "admin123"), nil)

	svc := NewAuthService(repo, "test-secret", 1)
	_, err := svc.Login(models.LoginInput{Username: "  admin  ", Password: "admin123"})

	require.NoError(t, err)
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := new(mockAdminRepository)
	repo.On("FindByUsername", "admin").Return(hashedAdmin(t, "admin", "admin123"), nil)

	svc := NewAuthService(repo, "test-secret", 1)
	_, err := svc.Login(models.LoginInput{Username: "admin", Password: "wrong"})

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLogin_UnknownUser(t *testing.T) {
	repo := new(mockAdminRepository)
	repo.On("FindByUsername", "ghost").Return(nil, gorm.ErrRecordNotFound)

	svc := NewAuthService(repo, "test-secret", 1)
	_, err := svc.Login(models.LoginInput{Username: "ghost", Password: "x"})

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLogin_EmptyInput(t *testing.T) {
	repo := new(mockAdminRepository)
	svc := NewAuthService(repo, "test-secret", 1)

	_, err := svc.Login(models.LoginInput{Username: "", Password: ""})

	assert.ErrorIs(t, err, ErrInvalidCredentials)
	repo.AssertNotCalled(t, "FindByUsername", mock.Anything)
}

func TestLogin_RepositoryError(t *testing.T) {
	repo := new(mockAdminRepository)
	repo.On("FindByUsername", "admin").Return(nil, errors.New("db down"))

	svc := NewAuthService(repo, "test-secret", 1)
	_, err := svc.Login(models.LoginInput{Username: "admin", Password: "admin123"})

	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrInvalidCredentials)
}
