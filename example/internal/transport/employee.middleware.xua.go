// Code generated by "xua server" from example/employee.go; DO NOT EDIT.
package transport

import (
	"context"

	employee "github.com/a-tho/grad-proj/example/internal"
	"github.com/a-tho/grad-proj/example/internal/types"
)

type employeeStorageCreate func(ctx context.Context, name, userAgent, bio string) (id int, err error)
type employeeStorageRead func(ctx context.Context, id int) (employee types.Employee, err error)
type employeeStorageUpdate func(ctx context.Context, id, vacationDays int, bio string) (err error)
type employeeStorageDelete func(ctx context.Context, id int) (err error)

type WrapEmployeeStorage func(next employee.EmployeeStorage) employee.EmployeeStorage

type WrapEmployeeStorageCreate func(next employeeStorageCreate) employeeStorageCreate
type WrapEmployeeStorageRead func(next employeeStorageRead) employeeStorageRead
type WrapEmployeeStorageUpdate func(next employeeStorageUpdate) employeeStorageUpdate
type WrapEmployeeStorageDelete func(next employeeStorageDelete) employeeStorageDelete
