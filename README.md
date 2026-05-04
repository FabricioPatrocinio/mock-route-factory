# mock-route-factory

HTTP API that stores route mocks (method + path) in PostgreSQL and replays them on demand. Built for QA / staging environments.

## Quick start

### 1. Postgres (Docker)

```bash
docker run -d \
  --name pg-mocks \
  -e POSTGRES_USER=mocks \
  -e POSTGRES_PASSWORD=mocks \
  -e POSTGRES_DB=mocks \
  -p 5432:5432 \
  postgres:16-alpine
```

### 2. Configure

```bash
cp .env.example .env
# edit .env — the defaults in .env.example match the docker command above
```

### 3. Install dependencies and run

```bash
go mod tidy
go run .
```

The server listens on `PORT` (default `8080`) and runs the database migration automatically on startup.

## Environment variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `PORT` | no | `8080` | HTTP listen port |
| `DB_WRITER_HOST` | **yes** | — | Postgres host for writes |
| `DB_READER_HOST` | **yes** | — | Postgres host for reads (can be same as writer) |
| `DB_PORT` | no | `5432` | Postgres port |
| `DB_NAME` | **yes** | — | Database name |
| `DB_SSL_MODE` | no | `disable` | pq SSL mode (`disable`, `require`, `verify-full`, …) |
| `DB_MOCKS_USER` | **yes** | — | Postgres user |
| `DB_MOCKS_PASSWORD` | **yes** | — | Postgres password |
| `DB_MOCKS_SCHEMA` | no | `public` | Schema for the mocks table |
| `PUBLIC_BASE_URL` | no | derived from request | Base URL used when generating curl commands |

`DB_MOCKS_SCHEMA` is validated as a safe SQL identifier (letters, digits, underscore; must start with letter or underscore).

## API reference

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check — `{"status":"ok"}` |
| `GET` | `/swagger` | Swagger UI |
| `GET` | `/swagger/openapi.yaml` | OpenAPI 3.0 spec |
| `POST` | `/admin/mocks` | Create or update a mock |
| `GET` | `/admin/mocks` | List all mocks |
| `GET` | `/admin/mocks/curls` | List mocks with ready-to-run curl commands |
| `DELETE` | `/admin/mocks?method=GET&path=/foo` | Delete a mock (404 if not found) |
| `*` | `/*` | Replay the stored mock for that method+path |

### Register a mock

```bash
curl -s -X POST http://localhost:8080/admin/mocks \
  -H 'Content-Type: application/json' \
  -d '{
    "method": "GET",
    "path": "/api/v1/users",
    "status": 200,
    "response": {"items": [], "total": 0}
  }'
```

### Call the mock

```bash
curl -s http://localhost:8080/api/v1/users
# → {"items":[],"total":0}
```

### Delete a mock

```bash
curl -s -X DELETE 'http://localhost:8080/admin/mocks?method=GET&path=/api/v1/users'
```

## Validation rules

- `method`: one of `GET POST PUT PATCH DELETE HEAD OPTIONS`
- `path`: must start with `/`; cannot be `/health` or start with `/admin`
- `status`: integer 100–599 (defaults to `200`)
- `response`: any valid JSON value (required)

Trailing slashes are stripped from paths (except root `/`).

## Build

```bash
go build -o mock-route-factory .
./mock-route-factory
```

## Limitations & security notes

- **No authentication** on `/admin` routes by default. In production, add an auth middleware (e.g. static token, mTLS) before the admin group.
- Admin paths and `/health` cannot be registered as mocks, but there is no RBAC beyond that.
- The `DB_MOCKS_SCHEMA` value is validated before use; all table/schema names are quoted with `pq.QuoteIdentifier` to prevent SQL injection.
- The server runs the migration on every startup — safe to restart freely.
