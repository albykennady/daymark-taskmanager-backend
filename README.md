# Task Manager API

A starter Go REST API for managing tasks. It uses Chi for routing and SQLite for storage. The current database is SQLite's in-memory database, so tasks are reset when the server stops.

## Run

```sh
go run ./cmd/api
```

At startup, the API opens the SQLite database, checks the connection, creates the tasks table if needed, and inserts three sample tasks. Seeding is idempotent. It listens on `:8080` by default. Set `SERVER_ADDRESS` to change the bind address.

## Endpoints

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/tasks` | Create a task |
| `GET` | `/tasks` | List all tasks |
| `GET` | `/tasks/{id}` | Get one task |
| `PUT` | `/tasks/{id}` | Replace a task using its numeric ID and a JSON body |
| `PATCH` | `/tasks/{id}` | Update selected task fields |
| `DELETE` | `/tasks/{id}` | Delete a task |

Create example:

```sh
curl -X POST http://localhost:8080/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Prepare release","description":"Review the checklist"}'
```

`PUT` requires `title`, `description`, and `completed`. `PATCH` accepts any non-empty subset of those fields. Task responses include an ID and UTC `created_at` and `updated_at` timestamps.

Sample task IDs are `1`, `2`, and `3`. For example, replace task `1` with:

```sh
curl -X PUT http://localhost:8080/tasks/1 \
  -H 'Content-Type: application/json' \
  -d '{"title":"Updated task","description":"Revised details","completed":true}'
```

## Project layout

```text
cmd/api/                 Application entry point and server wiring
internal/database/       SQLite connection setup and schema initialization
internal/handler/        HTTP routes, request validation, and JSON responses
internal/repository/     Repository interface and SQLite data access
internal/task/           Task domain model and repository input types
```

Run the test suite with `go test ./...`.
