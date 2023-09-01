package transport

import (
	"context"

	"github.com/a-tho/grad-proj/example/internal"
)

type EmployeeStorageCreate func(ctx context.Context, name, userAgent, bio string) (id int, err error)
type EmployeeStorageRead func(ctx context.Context, id int) (employee employee.Employee, err error)
type EmployeeStorageUpdate func(ctx context.Context, id, vacationDays int, bio string) (err error)
type EmployeeStorageDelete func(ctx context.Context, id int) (err error)

type WrapEmployeeStorage func(next employee.EmployeeStorage) employee.EmployeeStorage

type WrapEmployeeStorageCreate func(next EmployeeStorageCreate) EmployeeStorageCreate
type WrapEmployeeStorageRead func(next EmployeeStorageRead) EmployeeStorageRead
type WrapEmployeeStorageUpdate func(next EmployeeStorageUpdate) EmployeeStorageUpdate
type WrapEmployeeStorageDelete func(next EmployeeStorageDelete) EmployeeStorageDelete
