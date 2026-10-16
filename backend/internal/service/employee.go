package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"EMS/internal/models"
	"EMS/internal/repository"

	"gorm.io/gorm"
)

// Sentinel errors returned by the service layer and mapped to HTTP statuses.
var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("employee not found")
	ErrConflict   = errors.New("conflict")
)

var emailRegex = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// EmployeeRepository is the persistence contract required by EmployeeService.
type EmployeeRepository interface {
	Create(e *models.Employee) error
	GetByID(id uint) (*models.Employee, error)
	Update(e *models.Employee) error
	Delete(id uint) error
	List(f repository.ListFilter) ([]models.Employee, int64, error)
	All(f repository.ListFilter) ([]models.Employee, error)
	ExistsByEmail(email string, excludeID uint) (bool, error)
	DistinctDepartments() ([]string, error)
}

// EmployeeService implements the employee management business logic.
type EmployeeService struct {
	repo EmployeeRepository
}

// NewEmployeeService constructs a service.
func NewEmployeeService(repo EmployeeRepository) *EmployeeService {
	return &EmployeeService{repo: repo}
}

// Create validates and persists a new employee.
func (s *EmployeeService) Create(in models.EmployeeInput) (*models.Employee, error) {
	if err := validate(in); err != nil {
		return nil, err
	}

	email := normalizeEmail(in.Email)
	dup, err := s.repo.ExistsByEmail(email, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
	}
	if dup {
		return nil, fmt.Errorf("%w: an employee with this email already exists", ErrConflict)
	}

	emp := in.ToEmployee()
	emp.Status = in.Status
	if err := s.repo.Create(&emp); err != nil {
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}
	return &emp, nil
}

// GetByID loads a single employee.
func (s *EmployeeService) GetByID(id uint) (*models.Employee, error) {
	emp, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to load employee: %w", err)
	}
	return emp, nil
}

// Update validates and saves changes to an existing employee.
func (s *EmployeeService) Update(id uint, in models.EmployeeInput) (*models.Employee, error) {
	existing, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	if err := validate(in); err != nil {
		return nil, err
	}

	dup, err := s.repo.ExistsByEmail(normalizeEmail(in.Email), id)
	if err != nil {
		return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
	}
	if dup {
		return nil, fmt.Errorf("%w: another employee already uses this email", ErrConflict)
	}

	updated := in.ToEmployee()
	updated.ID = existing.ID
	updated.CreatedAt = existing.CreatedAt
	if err := s.repo.Update(&updated); err != nil {
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}
	return &updated, nil
}

// Delete removes an employee by id.
func (s *EmployeeService) Delete(id uint) error {
	if err := s.repo.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to delete employee: %w", err)
	}
	return nil
}

// List returns a paginated and filtered set of employees.
func (s *EmployeeService) List(f repository.ListFilter) (*models.PaginatedEmployees, error) {
	normalizeFilter(&f)

	employees, total, err := s.repo.List(f)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch employees: %w", err)
	}

	totalPages := 0
	if f.PageSize > 0 {
		totalPages = int((total + int64(f.PageSize) - 1) / int64(f.PageSize))
	}

	return &models.PaginatedEmployees{
		Data:       employees,
		Total:      total,
		Page:       f.Page,
		PageSize:   f.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Export returns all employees matching the filters (for CSV download).
func (s *EmployeeService) Export(f repository.ListFilter) ([]models.Employee, error) {
	employees, err := s.repo.All(f)
	if err != nil {
		return nil, fmt.Errorf("failed to export employees: %w", err)
	}
	return employees, nil
}

// Departments returns the distinct department names.
func (s *EmployeeService) Departments() ([]string, error) {
	deps, err := s.repo.DistinctDepartments()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch departments: %w", err)
	}
	return deps, nil
}

func normalizeFilter(f *repository.ListFilter) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 10
	}
	if f.PageSize > 100 {
		f.PageSize = 100
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validate(in models.EmployeeInput) error {
	if strings.TrimSpace(in.FullName) == "" {
		return fmt.Errorf("%w: full name is required", ErrValidation)
	}
	email := strings.TrimSpace(in.Email)
	if email == "" {
		return fmt.Errorf("%w: email is required", ErrValidation)
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("%w: email address is invalid", ErrValidation)
	}
	if in.Status != models.StatusActive && in.Status != models.StatusInactive {
		return fmt.Errorf("%w: status must be either %q or %q", ErrValidation, models.StatusActive, models.StatusInactive)
	}
	return nil
}
