package employee

import (
	"context"

	"github.com/a-tho/grad-proj/example/internal/types"
)

// @xua http-server
// @xua http-prefix=api/v1
type EmployeeStorage interface {
	// @xua http-method=POST
	// @xua http-path=employee/create
	// @xua http-cookie=name|x-name
	// @xua http-header=userAgent|User-Agent
	// @xua http-success=200
	Create(ctx context.Context, name, userAgent, bio string) (id int, err error)

	// @xua http-method=GET
	// @xua http-query=id|id
	// @xua http-path=employee/read
	// @xua http-success=200
	Read(ctx context.Context, id int) (employee types.Employee, err error)

	// @xua http-method=PATCH
	// @xua http-query=id|id,vacationDays|days
	// @xua http-path=employee/update
	// @xua http-success=204
	Update(ctx context.Context, id, vacationDays int, bio string) (err error)

	// @xua http-method=DELETE
	// @xua http-query=id|id
	// @xua http-path=employee/delete
	// @xua http-success=204
	Delete(ctx context.Context, id int) (err error)
}
