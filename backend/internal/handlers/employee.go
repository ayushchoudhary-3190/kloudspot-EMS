package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"EMS/internal/models"
	"EMS/internal/repository"
	"EMS/internal/service"
)

// EmployeeHandler exposes employee CRUD endpoints.
type EmployeeHandler struct {
	svc *service.EmployeeService
}

// NewEmployeeHandler constructs an EmployeeHandler.
func NewEmployeeHandler(svc *service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{svc: svc}
}

func (h *EmployeeHandler) filterFrom(r *http.Request) repository.ListFilter {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("pageSize"))
	return repository.ListFilter{
		Search:     q.Get("search"),
		Department: q.Get("department"),
		Status:     q.Get("status"),
		Page:       page,
		PageSize:   pageSize,
	}
}

func pathID(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid employee id")
	}
	return uint(id), nil
}

// List returns a paginated, filtered employee list.
func (h *EmployeeHandler) List(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.List(h.filterFrom(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// Get returns a single employee.
func (h *EmployeeHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	emp, err := h.svc.GetByID(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, emp)
}

// Create adds a new employee.
func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in models.EmployeeInput
	if err := decodeJSON(r, &in); err != nil {
		badRequest(w, "invalid request body")
		return
	}
	emp, err := h.svc.Create(in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, emp)
}

// Update edits an existing employee.
func (h *EmployeeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	var in models.EmployeeInput
	if err := decodeJSON(r, &in); err != nil {
		badRequest(w, "invalid request body")
		return
	}
	emp, err := h.svc.Update(id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, emp)
}

// Delete removes an employee.
func (h *EmployeeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	if err := h.svc.Delete(id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Departments returns the distinct departments.
func (h *EmployeeHandler) Departments(w http.ResponseWriter, r *http.Request) {
	deps, err := h.svc.Departments()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, deps)
}

// Export streams matching employees as a CSV file.
func (h *EmployeeHandler) Export(w http.ResponseWriter, r *http.Request) {
	employees, err := h.svc.Export(h.filterFrom(r))
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="employees.csv"`)
	w.WriteHeader(http.StatusOK)

	cw := csv.NewWriter(w)
	defer cw.Flush()

	_ = cw.Write([]string{
		"Employee ID", "Full Name", "Email", "Phone", "Department",
		"Designation", "Date of Joining", "Salary", "Status",
	})
	for _, e := range employees {
		doj := ""
		if !e.DateOfJoining.IsZero() {
			doj = e.DateOfJoining.Format("2006-01-02")
		}
		_ = cw.Write([]string{
			strconv.FormatUint(uint64(e.ID), 10),
			e.FullName,
			e.Email,
			e.Phone,
			e.Department,
			e.Designation,
			doj,
			strconv.FormatUint(e.Salary, 10),
			e.Status,
		})
	}
}
