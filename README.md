# GoAdminer

A database management web app written in Go — inspired by Adminer.

Connect to a database, browse tables, view and edit rows, and run SQL queries — all from a browser.

## Features

- **Multiple databases**: PostgreSQL and SQLite (extensible via the `driver.Driver` interface)
- **Table browser**: paginated, sortable, searchable
- **Row CRUD**: insert, edit, delete rows with auto-generated forms
- **SQL query runner**: execute arbitrary SQL, see results
- **Single binary**: frontend embedded in the Go binary via `embed.FS`
- **Docker**: multi-stage build, minimal Alpine image

## Quick start

```bash
# Build and run
go build -o goadminer ./cmd/server
./goadminer

# With Docker
docker compose up --build
```

Then open http://localhost:8080.

## Development

```bash
# Serve frontend from disk (hot-reloadable)
go run ./cmd/server -dev
```

## API

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/connect` | Connect to a database |
| `GET` | `/api/status` | Connection status |
| `POST` | `/api/disconnect` | Disconnect |
| `GET` | `/api/databases` | List databases |
| `GET` | `/api/tables` | List tables |
| `GET` | `/api/tables/{table}/schema` | Table schema |
| `GET` | `/api/tables/{table}/rows` | Browse rows |
| `POST` | `/api/tables/{table}/rows` | Insert row |
| `PUT` | `/api/tables/{table}/rows` | Update row |
| `DELETE` | `/api/tables/{table}/rows` | Delete row |
| `POST` | `/api/query` | Execute SQL |

## Adding a new driver

1. Create `internal/driver/<name>.go`
2. Implement the `driver.Driver` interface
3. Register it in `internal/handler/handler.go` drivers map
4. Add the driver import to `cmd/server/main.go`
5. Add the driver toggle button in `web/js/views/connect.js`

## Project structure

```
├── cmd/server/main.go     # Entry point
├── internal/
│   ├── driver/            # Database drivers (PG, SQLite)
│   ├── handler/           # HTTP handlers
│   ├── model/             # Shared types
│   ├── server/            # HTTP server, routes, embed
│   └── session/           # Connection session store
├── web/                   # Frontend (vanilla JS SPA)
│   ├── index.html
│   ├── css/style.css
│   └── js/
│       ├── app.js         # Router, nav
│       ├── api.js         # API client
│       ├── utils.js       # Helpers
│       └── views/         # View components
├── dockerfile
├── compose.yml
└── go.mod
```
