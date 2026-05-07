# SaaS Baseon Go

SaaS Baseon rebuilt with Go + Gin + GORM + PostgreSQL. Product requirements, menu structure, permission codes, documents and frontend behavior are kept from the original project; this repository changes the technical stack and project architecture only.

## Stack

- Go 1.22+ compatible, Docker build uses Go 1.23
- Gin
- GORM
- PostgreSQL
- Redis
- DDD-style package boundaries
- TDD-first module workflow
- Vue 3 frontend copied from the original project

## Quick Start

```bash
cp .env.example .env
docker compose up --build
```

Health check:

```bash
curl http://127.0.0.1:8081/health
```

Production-like runs can disable development AutoMigrate:

```bash
DB_AUTO_MIGRATE=false docker compose up --build
```

The current PostgreSQL schema baseline is stored at `internal/infrastructure/persistence/postgres/schema/current_schema.sql`.

Frontend:

```text
http://127.0.0.1:8082
```

## Structure

```text
cmd/api                         application entrypoint
internal/bootstrap              config, db, redis, router wiring
internal/interfaces/http        handlers, middleware, response contracts
internal/application            use cases
internal/domain                 domain entities and repository contracts
internal/infrastructure         postgres/redis implementations
tests                           integration and contract tests
frontend                        Vue 3 management UI
docs                            copied project documents updated to current Go architecture
```

## Migration Rule

Keep API response compatible with the current frontend:

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

## Demo Accounts

```text
E10001 / 112233
E10100 / 112233
```

## Current Build Status

Current implementation contains:

- Docker Compose runtime for API, Web, PostgreSQL and Redis
- Gin router with frontend-compatible API response envelope
- PostgreSQL schema migration through GORM AutoMigrate for the rebuilt modules
- seed data for platform tenant, demo users, role, menu/button permissions, SaaS plans, dictionaries, params and logs
- login with password verification and signed bearer token
- captcha generation, login failure tracking, protected API authentication and permission-code gate checks
- system management APIs for tenant, plan, organization, position, business unit, user, role, menu, dictionary, parameter, operation log and login log pages
- monitor APIs for health, server info, service overview, scheduled jobs, Redis stats and Redis key scan
- file upload/download/delete and CSV batch import/export APIs
- explicit API route coverage with strict 404 for unknown API paths
- frontend build and Playwright E2E coverage for the current visible pages
