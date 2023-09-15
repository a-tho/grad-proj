# grad-proj

Let's say you want to store info on employees. You could [implement](example/internal/employee/storage.go) a simple CRUD storage with an interface

```go
type EmployeeStorage interface {
	Create(ctx context.Context, name, userAgent, bio string) (id int, err error)
	Read(ctx context.Context, id int) (employee types.Employee, err error)
	Update(ctx context.Context, id, vacationDays int, bio string) (err error)
	Delete(ctx context.Context, id int) (err error)
}
```

Then you would need to implement a server that exposes an endpoint for each of the method (`Create`, `Read` etc) of the interface `EmployeeStorage`, specifies what headers/cookies/request bodies/query parameters etc you are interested in.

To minimize writing boilerplate code yourself, simply [add comments](example/internal/employee.go) for each method and the interface itself:

```go
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
```

Then run `xua server --in path/to/pkg/with/interface`, and voilà, you [have](example/transport) all the necessary code to [run](example/cmd/main.go) your app!