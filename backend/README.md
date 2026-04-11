# CyberToolkit Backend

Go backend for CyberToolkit.

## Scope

Current implementation provides:

- Public API for home, categories, tools, tags, submissions
- Admin API for login, current user, categories, tools
- In-memory data store for rapid iteration
- SQL schema file for later PostgreSQL migration

## Run

```bash
cd backend
go run ./cmd/api
```

Default address:

```text
http://localhost:8080
```

## Environment Variables

- `APP_ADDR`: server address, default `:8080`
- `APP_ADMIN_EMAIL`: admin login email, default `admin@cybertoolkit.local`
- `APP_ADMIN_PASSWORD`: admin login password, default `admin123456`
- `APP_ADMIN_TOKEN`: static admin bearer token, default `dev-admin-token`

## Notes

- This version intentionally uses only Go standard library.
- Persistence is currently in memory.
- `backend/sql/schema.sql` contains the relational schema for a future PostgreSQL-backed version.
