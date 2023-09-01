package employee

import (
	"context"
)

type Employee struct {
	Name         string // from a cookie
	ID           int    // from a query arg
	VacationDays int    // from a query arg

	UserAgent string // from a header
	Bio       string // from a request body
}

// @xua http-server
// @xua http-prefix=api/v1
// @xua domain-name=employee
type EmployeeStorage interface {
	// @xua http-method=POST
	// @xua http-path=employee/create
	// @xua http-cookie=name|x-name
	// @xua http-header=userAgent|User-Agent
	// @xua http-success=200
	Create(ctx context.Context, name, userAgent, bio string) (id int, err error)

	// @xua http-method=GET
	// @xua http-query=`id|id`
	// @xua http-path=employee/read
	// @xua http-success=200
	Read(ctx context.Context, id int) (employee Employee, err error)

	// @xua http-method=PATCH
	// @xua http-query=`id|id,vacationDays|days`
	// @xua http-path=employee/update
	// @xua http-success=204
	Update(ctx context.Context, id, vacationDays int, bio string) (err error)

	// @xua http-method=DELETE
	// @xua http-query=`id|id`
	// @xua http-path=employee/delete
	// @xua http-success=204
	Delete(ctx context.Context, id int) (err error)
}
