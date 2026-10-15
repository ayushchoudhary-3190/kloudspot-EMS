package EMS

import (
	"time"
)

type Employee struct{
	empID 		uint
	fullName 	string
	email		string
	phone		string
	department	string
	designation string
	doj			time.Time
	salary		uint
	status		string
}