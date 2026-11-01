# User Age API

A RESTful API in Go that manages users (`name`, `dob`) and returns each user's **age calculated dynamically** from their date of birth.

**Stack:** Go, GoFiber, PostgreSQL, SQLC, Uber Zap, go-playground/validator

## Project structure

```
cmd/server/main.go        # entrypoint, wiring, graceful shutdown
config/                   # env-based configuration
db/migrations/            # SQL schema (users table)
db/query/                 # SQLC query definitions
db/sqlc/                  # SQLC-generated data access layer
internal/handler/         # HTTP handlers, validation, error mapping
internal/service/         # business logic + age calculation (+ unit tests)
internal/repository/      # persistence abstraction over SQLC
internal/routes/          # route registration
internal/middleware/      # request ID + request duration logging
internal/models/          # request/response DTOs and domain errors
internal/logger/          # Zap logger setup
```

## Option A: run with Docker (easiest)

```bash
docker compose up --build
```

The API is available at `http://localhost:3000`. The compose file starts Postgres and creates the `users` table automatically.

## Option B: run locally

Prerequisites: Go 1.22+, PostgreSQL 14+.

```bash
# 1. Install dependencies
go mod tidy

# 2. Create the database and table
createdb usersdb
psql usersdb -f db/migrations/000001_create_users_table.up.sql

# 3. Configure (optional; defaults match the line below)
cp .env.example .env
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/usersdb?sslmode=disable"

# 4. Run
go run ./cmd/server
```

### Environment variables

| Variable       | Default                                                                  | Description                       |
|----------------|--------------------------------------------------------------------------|-----------------------------------|
| `PORT`         | `3000`                                                                   | HTTP port                         |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/usersdb?sslmode=disable`    | Postgres connection string        |
| `APP_ENV`      | `development`                                                            | `production` gives JSON logs      |
| `LOG_LEVEL`    | `info`                                                                   | `debug`, `info`, `warn`, `error`  |

## Regenerating SQLC code

The generated code is committed in `db/sqlc`. To regenerate after changing `db/query/*.sql` or the migrations:

```bash
# install: https://docs.sqlc.dev/en/latest/overview/install.html
sqlc generate
```

## Tests

```bash
go test ./... -v
```

Includes table-driven unit tests for the age calculation (birthday passed / not yet / today, leap day, newborn, future date).

## API

| Method | Path          | Success                          | Errors            |
|--------|---------------|----------------------------------|-------------------|
| POST   | `/users`      | `201` `{id,name,dob}`            | `400`             |
| GET    | `/users/:id`  | `200` `{id,name,dob,age}`        | `400`, `404`      |
| PUT    | `/users/:id`  | `200` `{id,name,dob}`            | `400`, `404`      |
| DELETE | `/users/:id`  | `204` (no body)                  | `400`, `404`      |
| GET    | `/users`      | `200` `[{id,name,dob,age}, ...]` | `400`             |
| GET    | `/health`     | `200` `{"status":"ok"}`          |                   |

### Examples

```bash
# Create
curl -i -X POST localhost:3000/users -H 'Content-Type: application/json' \
  -d '{"name":"Alice","dob":"1990-05-10"}'

# Get (includes age)
curl localhost:3000/users/1

# Update
curl -X PUT localhost:3000/users/1 -H 'Content-Type: application/json' \
  -d '{"name":"Alice Updated","dob":"1991-03-15"}'

# List with pagination
curl -i "localhost:3000/users?page=1&limit=10"

# Delete
curl -i -X DELETE localhost:3000/users/1
```

### Pagination

`GET /users?page=1&limit=20` (defaults: `page=1`, `limit=20`, max `limit=100`). The body remains a plain JSON array; metadata is returned in headers: `X-Total-Count`, `X-Page`, `X-Limit`, `X-Total-Pages`.

### Validation and errors

Errors use a consistent JSON shape:

```json
{ "error": "validation failed", "details": { "dob": "dob must be a valid date in YYYY-MM-DD format" } }
```

- `name`: required, 1–100 characters
- `dob`: required, `YYYY-MM-DD`, must be a real date and not in the future
- `:id`: must be a positive integer

### Middleware

- **Request ID:** every response carries an `X-Request-ID` header (an incoming one is reused, otherwise a UUID is generated).
- **Request logging:** each request is logged via Zap with method, path, status, duration, and request ID.

## Design notes

- `dob` is stored as a `DATE`; `age` is never stored. It is computed per request with `service.CalculateAge` using Go's `time` package (UTC calendar dates).
- Layering is handler → service → repository → SQLC, with the service and repository behind interfaces so they can be tested with fakes.
