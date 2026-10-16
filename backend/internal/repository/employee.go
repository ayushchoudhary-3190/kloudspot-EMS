package repository

import (
	"strings"

	"EMS/internal/models"

	"gorm.io/gorm"
)

// ListFilter describes search, filtering and pagination options.
type ListFilter struct {
	Search     string
	Department string
	Status     string
	Page       int
	PageSize   int
}

// EmployeeRepository persists employees with GORM.
type EmployeeRepository struct {
	DB *gorm.DB
}

// NewEmployeeRepository constructs a repository.
func NewEmployeeRepository(db *gorm.DB) *EmployeeRepository {
	return &EmployeeRepository{DB: db}
}

// Create inserts a new employee.
func (r *EmployeeRepository) Create(e *models.Employee) error {
	return r.DB.Create(e).Error
}

// GetByID returns an employee or gorm.ErrRecordNotFound.
func (r *EmployeeRepository) GetByID(id uint) (*models.Employee, error) {
	var e models.Employee
	if err := r.DB.First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

// Update saves all fields of an existing employee.
func (r *EmployeeRepository) Update(e *models.Employee) error {
	return r.DB.Save(e).Error
}

// Delete removes an employee, returning gorm.ErrRecordNotFound if absent.
func (r *EmployeeRepository) Delete(id uint) error {
	res := r.DB.Delete(&models.Employee{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// List returns a paginated, filtered page of employees plus the total count.
func (r *EmployeeRepository) List(f ListFilter) ([]models.Employee, int64, error) {
	query := applyFilters(r.DB.Model(&models.Employee{}), f)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	employees := make([]models.Employee, 0)
	offset := (f.Page - 1) * f.PageSize
	if err := query.Order("id asc").Limit(f.PageSize).Offset(offset).Find(&employees).Error; err != nil {
		return nil, 0, err
	}
	return employees, total, nil
}

// All returns every employee matching the filters (used for CSV export).
func (r *EmployeeRepository) All(f ListFilter) ([]models.Employee, error) {
	query := applyFilters(r.DB.Model(&models.Employee{}), f)
	employees := make([]models.Employee, 0)
	err := query.Order("id asc").Find(&employees).Error
	return employees, err
}

// ExistsByEmail reports whether another employee already uses the email.
func (r *EmployeeRepository) ExistsByEmail(email string, excludeID uint) (bool, error) {
	query := r.DB.Model(&models.Employee{}).Where("email = ?", email)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// DistinctDepartments returns the sorted list of departments in use.
func (r *EmployeeRepository) DistinctDepartments() ([]string, error) {
	deps := make([]string, 0)
	err := r.DB.Model(&models.Employee{}).
		Distinct().
		Where("department <> ''").
		Order("department asc").
		Pluck("department", &deps).Error
	return deps, err
}

func applyFilters(query *gorm.DB, f ListFilter) *gorm.DB {
	if s := strings.TrimSpace(f.Search); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		query = query.Where("LOWER(full_name) LIKE ? OR LOWER(email) LIKE ?", like, like)
	}
	if d := strings.TrimSpace(f.Department); d != "" {
		query = query.Where("department = ?", d)
	}
	if st := strings.TrimSpace(f.Status); st != "" {
		query = query.Where("status = ?", st)
	}
	return query
}
