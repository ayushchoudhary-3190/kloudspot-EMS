package EMS

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
)

func CreateEmployee(e *Employee) error{
	if e.fullName == ""{
		return errors.New("Empty name field")
	}
	if e.email == ""{
		return errors.New("empty email field")
	}
	err := repository.CreateEmployee(e)
	return err
}

func RemoveEmployee(employeeID uint) (error){
	err:= repository.DeleteEmployee(employeeID)
	if err!=nil{
		if errors.Is(err, gorm.ErrRecordNotFound){
			return fmt.Errorf("employee with ID not found")
		}
		return fmt.Errorf("failed to delete employee")
	}
	return nil
}

func DisplayDetails(page int) ([]Employee,error){
	result,err:= repository.FetchDetails(page)
	if err!=nil{
		return nil,fmt.Errorf("failed to fetch employee details : %w",err)
	}
	return result,nil
}
