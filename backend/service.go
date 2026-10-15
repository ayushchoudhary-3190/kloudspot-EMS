package EMS

import "errors"

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

func RemoveEmployee()