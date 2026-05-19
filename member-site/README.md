# Member Site

A backend webserver for managing users, written in Go.

```bash
# run any pending migrations
go run ./scripts/migrate up
# run test suite
go test ./...
# start the server
go run ./member-site
# for live reload on changes, use gow
go install github.com/mitranim/gow@latest
gow run ./member-site
```

## Development

Models and other conceptual units are given their own packages, which follow this convention:

- `{model_name}.go` defines the model struct and associated methods
- `queries.go` defines the database operations
- `controller.go` defines the application-level functions/operations that can be performed
- `router.go` defines the API controller to map endpoints to operations

The separation of controllers and routers offers a separation of concerns (for example, a
hypothetical future CLI could expose the same operations as the API by calling the same controllers).

The tests for the router endpoints are defined in the top-level `member-site` package. This separation
is clunky but allows us to run the test end-to-end, checking there are no conflicts between routers
and that the router is properly attached to the main server.

## Framework tools

- Database: [sqlx](https://jmoiron.github.io/sqlx/), [squirrel](https://github.com/masterminds/squirrel)
- Routing: [chi](https://github.com/go-chi/chi)
- Migrations: [golang-migrate](github.com/golang-migrate/migrate)
