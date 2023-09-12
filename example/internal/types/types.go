package types

type Employee struct {
	Name         string // from a cookie
	ID           int    // from a query arg
	VacationDays int    // from a query arg

	UserAgent string // from a header
	Bio       string // from a request body
}
