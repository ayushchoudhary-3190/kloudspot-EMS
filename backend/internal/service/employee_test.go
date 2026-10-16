package service

import (
	"errors"
	"testing"

	"EMS/internal/models"
	"EMS/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// mockEmployeeRepository is a testify mock of EmployeeRepository.
type mockEmployeeRepository struct {
	mock.Mock
}

func (m *mockEmployeeRepository) Create(e *models.Employee) error {
	args := m.Called(e)
	return args.Error(0)
}

func (m *mockEmployeeRepository) GetByID(id uint) (*models.Employee, error) {
	args := m.Called(id)
	if emp := args.Get(0); emp != nil {
		return emp.(*models.Employee), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEmployeeRepository) Update(e *models.Employee) error {
	args := m.Called(e)
	return args.Error(0)
}

func (m *mockEmployeeRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockEmployeeRepository) List(f repository.ListFilter) ([]models.Employee, int64, error) {
	args := m.Called(f)
	emps := args.Get(0)
	if emps == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return emps.([]models.Employee), args.Get(1).(int64), args.Error(2)
}

func (m *mockEmployeeRepository) All(f repository.ListFilter) ([]models.Employee, error) {
	args := m.Called(f)
	emps := args.Get(0)
	if emps == nil {
		return nil, args.Error(1)
	}
	return emps.([]models.Employee), args.Error(1)
}

func (m *mockEmployeeRepository) ExistsByEmail(email string, excludeID uint) (bool, error) {
	args := m.Called(email, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *mockEmployeeRepository) DistinctDepartments() ([]string, error) {
	args := m.Called()
	deps := args.Get(0)
	if deps == nil {
		return nil, args.Error(1)
	}
	return deps.([]string), args.Error(1)
}

func validInput() models.EmployeeInput {
	return models.EmployeeInput{
		FullName:    "Jane Doe",
		Email:       "Jane.Doe@example.com",
		Phone:       "9876543210",
		Department:  "Engineering",
		Designation: "Backend Developer",
		Salary:      60000,
		Status:      models.StatusActive,
	}
}

func TestCreate_Success_NormalizesEmail(t *testing.T) {
	repo := new(mockEmployeeRepository)
	repo.On("ExistsByEmail", "jane.doe@example.com", uint(0)).Return(false, nil)
	repo.On("Create", mock.MatchedBy(func(e *models.Employee) bool {
		return e.Email == "jane.doe@example.com" && e.FullName == "Jane Doe"
	})).Return(nil)

	svc := NewEmployeeService(repo)
	emp, err := svc.Create(validInput())

	require.NoError(t, err)
	assert.Equal(t, "jane.doe@example.com", emp.Email)
	assert.Equal(t, models.StatusActive, emp.Status)
	repo.AssertExpectations(t)
}

func TestCreate_ValidationErrors(t *testing.T) {
	cases := map[string]models.EmployeeInput{
		"empty name":     {Email: "a@b.com", Status: models.StatusActive},
		"empty email":    {FullName: "John", Status: models.StatusActive},
		"invalid email":  {FullName: "John", Email: "not-an-email", Status: models.StatusActive},
		"invalid status": {FullName: "John", Email: "a@b.com", Status: "Fired"},
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			repo := new(mockEmployeeRepository)
			svc := NewEmployeeService(repo)
			_, err := svc.Create(in)
			assert.ErrorIs(t, err, ErrValidation)
			repo.AssertNotCalled(t, "Create", mock.Anything)
		})
	}
}

func TestCreate_DuplicateEmail(t *testing.T) {
	repo := new(mockEmployeeRepository)
	repo.On("ExistsByEmail", "jane.doe@example.com", uint(0)).Return(true, nil)

	svc := NewEmployeeService(repo)
	_, err := svc.Create(validInput())

	assert.ErrorIs(t, err, ErrConflict)
	repo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestCreate_RepositoryError(t *testing.T) {
	repo := new(mockEmployeeRepository)
	repo.On("ExistsByEmail", mock.Anything, mock.Anything).Return(false, nil)
	repo.On("Create", mock.Anything).Return(errors.New("db down"))

	svc := NewEmployeeService(repo)
	_, err := svc.Create(validInput())

	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrValidation)
}

func TestGetByID_NotFound(t *testing.T) {
	repo := new(mockEmployeeRepository)
	repo.On("GetByID", uint(42)).Return(nil, gorm.ErrRecordNotFound)

	svc := NewEmployeeService(repo)
	_, err := svc.GetByID(42)

	assert.ErrorIs(t, err, ErrNotFound)
}

func TestGetByID_Success(t *testing.T) {
	repo := new(mockEmployeeRepository)
	expected := &models.Employee{ID: 7, FullName: "Jane"}
	repo.On("GetByID", uint(7)).Return(expected, nil)

	svc := NewEmployeeService(repo)
	emp, err := svc.GetByID(7)

	require.NoError(t, err)
	assert.Equal(t, expected, emp)
}

func TestUpdate_Success(t *testing.T) {
	repo := new(mockEmployeeRepository)
	existing := &models.Employee{ID: 5, FullName: "Old"}
	in := validInput()
	in.FullName = "New Name"

	repo.On("GetByID", uint(5)).Return(existing, nil)
	repo.On("ExistsByEmail", "jane.doe@example.com", uint(5)).Return(false, nil)
	repo.On("Update", mock.MatchedBy(func(e *models.Employee) bool {
		return e.ID == 5 && e.FullName == "New Name"
	})).Return(nil)

	svc := NewEmployeeService(repo)
	emp, err := svc.Update(5, in)

	require.NoError(t, err)
	assert.Equal(t, uint(5), emp.ID)
	assert.Equal(t, "New Name", emp.FullName)
}

func TestUpdate_NotFound(t *testing.T) {
	repo := new(mockEmployeeRepository)
	repo.On("GetByID", uint(9)).Return(nil, gorm.ErrRecordNotFound)

	svc := NewEmployeeService(repo)
	_, err := svc.Update(9, validInput())

	assert.ErrorIs(t, err, ErrNotFound)
	repo.AssertNotCalled(t, "Update", mock.Anything)
}

func TestUpdate_Conflict(t *testing.T) {
	repo := new(mockEmployeeRepository)
	repo.On("GetByID", uint(3)).Return(&models.Employee{ID: 3}, nil)
	repo.On("ExistsByEmail", mock.Anything, uint(3)).Return(true, nil)

	svc := NewEmployeeService(repo)
	_, err := svc.Update(3, validInput())

	assert.ErrorIs(t, err, ErrConflict)
	repo.AssertNotCalled(t, "Update", mock.Anything)
}

func TestDelete_Success(t *testing.T) {
	repo := new(mockEmployeeRepository)
	repo.On("Delete", uint(1)).Return(nil)

	svc := NewEmployeeService(repo)
	assert.NoError(t, svc.Delete(1))
}

func TestDelete_NotFound(t *testing.T) {
	repo := new(mockEmployeeRepository)
	repo.On("Delete", uint(2)).Return(gorm.ErrRecordNotFound)

	svc := NewEmployeeService(repo)
	assert.ErrorIs(t, svc.Delete(2), ErrNotFound)
}

func TestList_NormalizesPaginationAndComputesTotalPages(t *testing.T) {
	repo := new(mockEmployeeRepository)
	repo.On("List", mock.MatchedBy(func(f repository.ListFilter) bool {
		return f.Page == 1 && f.PageSize == 10
	})).Return([]models.Employee{{ID: 1}}, int64(25), nil)

	svc := NewEmployeeService(repo)
	res, err := svc.List(repository.ListFilter{Page: 0, PageSize: 0})

	require.NoError(t, err)
	assert.Equal(t, 1, res.Page)
	assert.Equal(t, 10, res.PageSize)
	assert.Equal(t, int64(25), res.Total)
	assert.Equal(t, 3, res.TotalPages)
}

func TestList_CapsPageSize(t *testing.T) {
	repo := new(mockEmployeeRepository)
	repo.On("List", mock.MatchedBy(func(f repository.ListFilter) bool {
		return f.PageSize == 100
	})).Return([]models.Employee{}, int64(0), nil)

	svc := NewEmployeeService(repo)
	res, err := svc.List(repository.ListFilter{Page: 2, PageSize: 500})

	require.NoError(t, err)
	assert.Equal(t, 100, res.PageSize)
	assert.Equal(t, 0, res.TotalPages)
}

func TestExport_Error(t *testing.T) {
	repo := new(mockEmployeeRepository)
	repo.On("All", mock.Anything).Return(nil, errors.New("boom"))

	svc := NewEmployeeService(repo)
	_, err := svc.Export(repository.ListFilter{})

	require.Error(t, err)
}

func TestDepartments(t *testing.T) {
	repo := new(mockEmployeeRepository)
	repo.On("DistinctDepartments").Return([]string{"Engineering", "HR"}, nil)

	svc := NewEmployeeService(repo)
	deps, err := svc.Departments()

	require.NoError(t, err)
	assert.Equal(t, []string{"Engineering", "HR"}, deps)
}
