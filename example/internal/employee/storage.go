package storage

import (
	"context"
	"errors"

	"github.com/a-tho/grad-proj/example/internal/types"
)

type EmployeeStorage struct {
	latestID int
	storage  map[int]types.Employee
}

func New() *EmployeeStorage {
	return &EmployeeStorage{
		storage: make(map[int]types.Employee),
	}
}

func (svc *EmployeeStorage) Create(ctx context.Context, name, userAgent, bio string) (id int, err error) {
	if name == "Voldemort" {
		return 0, errors.New("He-Who-Must-Not-Be-Named: access denied")
	}

	svc.latestID++
	id = svc.latestID

	svc.storage[id] = types.Employee{
		Name:      name,
		ID:        id,
		UserAgent: userAgent,
		Bio:       bio,
	}
	return id, nil
}

func (svc *EmployeeStorage) Read(ctx context.Context, id int) (employee types.Employee, err error) {
	if err = isValidID(id); err != nil {
		return employee, err
	}
	employee, ok := svc.storage[id]
	if !ok {
		err = errors.New("not found")
	}
	return employee, err
}

func (svc *EmployeeStorage) Update(ctx context.Context, id, vacationDays int, bio string) (err error) {
	if err = isValidID(id); err != nil {
		return err
	}
	employee, ok := svc.storage[id]
	if !ok {
		return errors.New("not found")
	}
	employee.VacationDays = vacationDays
	employee.Bio = bio
	svc.storage[id] = employee
	return nil
}

func (svc *EmployeeStorage) Delete(ctx context.Context, id int) (err error) {
	if err = isValidID(id); err != nil {
		return err
	}
	delete(svc.storage, id)
	return nil
}

func isValidID(id int) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	return nil
}
