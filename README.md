# SaaS Baseon Go

SaaS Baseon rebuilt project skeleton using Go + Gin + GORM + PostgreSQL. Product requirements and business rules are kept from the original project; this repository changes the technical stack and project architecture only.

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

## Current Build Status

The initial Go backend contains:

- application entrypoint
- config loading
- Gin router
- unified response envelope
- request id middleware
- PostgreSQL connection through GORM
- Redis connection
- `/health`
- parameter-management POC module as the DDD/TDD template

Business modules should be ported one by one without changing product logic.
