package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

// Date is a date-only type that (de)serializes as "YYYY-MM-DD" in JSON and
// persists as a date string in the database.
type Date struct {
	time.Time
}

func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return json.Marshal(nil)
	}
	return json.Marshal(d.Time.Format(dateLayout))
}

func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		return nil
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return fmt.Errorf("invalid date %q, expected YYYY-MM-DD", s)
	}
	d.Time = t
	return nil
}

// Value implements driver.Valuer so GORM can persist the date.
func (d Date) Value() (driver.Value, error) {
	if d.Time.IsZero() {
		return nil, nil
	}
	return d.Time.Format(dateLayout), nil
}

// Scan implements sql.Scanner so GORM can load the date.
func (d *Date) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		d.Time = v
	case string:
		return d.parse(v)
	case []byte:
		return d.parse(string(v))
	default:
		return fmt.Errorf("unsupported date type %T", value)
	}
	return nil
}

func (d *Date) parse(s string) error {
	if s == "" {
		return nil
	}
	// Accept both date-only and full RFC3339 timestamps.
	layouts := []string{dateLayout, time.RFC3339}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			d.Time = t
			return nil
		}
	}
	return fmt.Errorf("cannot parse date %q", s)
}

// Employment status values.
const (
	StatusActive   = "Active"
	StatusInactive = "Inactive"
)

// Employee is a persisted employee record.
type Employee struct {
	ID            uint      `gorm:"primaryKey" json:"employeeId"`
	FullName      string    `gorm:"not null" json:"fullName"`
	Email         string    `gorm:"not null;uniqueIndex" json:"email"`
	Phone         string    `json:"phone"`
	Department    string    `gorm:"index" json:"department"`
	Designation   string    `json:"designation"`
	DateOfJoining Date      `json:"dateOfJoining"`
	Salary        uint64    `json:"salary"`
	Status        string    `gorm:"index" json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// EmployeeInput is the request payload for creating/updating an employee.
type EmployeeInput struct {
	FullName      string `json:"fullName"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Department    string `json:"department"`
	Designation   string `json:"designation"`
	DateOfJoining Date   `json:"dateOfJoining"`
	Salary        uint64 `json:"salary"`
	Status        string `json:"status"`
}

// ToEmployee converts input into an Employee model.
func (in EmployeeInput) ToEmployee() Employee {
	return Employee{
		FullName:      strings.TrimSpace(in.FullName),
		Email:         strings.TrimSpace(strings.ToLower(in.Email)),
		Phone:         strings.TrimSpace(in.Phone),
		Department:    strings.TrimSpace(in.Department),
		Designation:   strings.TrimSpace(in.Designation),
		DateOfJoining: in.DateOfJoining,
		Salary:        in.Salary,
		Status:        in.Status,
	}
}

// Admin is the administrator account allowed to manage employees.
type Admin struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"not null;uniqueIndex" json:"username"`
	PasswordHash string    `gorm:"not null" json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// LoginInput is the request payload for authentication.
type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// PaginatedEmployees is the response envelope for the employee list endpoint.
type PaginatedEmployees struct {
	Data       []Employee `json:"data"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"pageSize"`
	TotalPages int        `json:"totalPages"`
}
