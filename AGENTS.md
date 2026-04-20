# DeepWrite AGENTS Guide

## Project Overview

DeepWrite is a monorepo project for immersive AI-collaborative research work platform.

**Tech Stack:**
- Frontend: Vue 3 + TypeScript + Vite + TailwindCSS + DaisyUI
- Backend: Go + Gin + GORM
- AI/Worker: Python + FastAPI + Celery
- Database: PostgreSQL + Redis

**Commands:**
- `pnpm dev` - Start all services
- `pnpm build` - Build all projects
- `pnpm lint` - Run lint checks
- `pnpm gen:openapi` - Generate Swagger docs (gateway)

---

## Code Conventions

### Go (services/gateway)

1. Follow standard Go project layout
2. Handlers in `handler/`, models in `models/`
3. Use GORM for database operations
4. Return errors via `pkg/response` package
5. **NEVER use string concatenation for SQL ORDER BY clauses** - validate against whitelist

### TypeScript (apps/web, apps/admin)

1. Strict mode enabled
2. Use Pinia for state management
3. API calls via `src/api/` modules
4. Components in `src/components/`
5. **Always wrap JSON.parse in try-catch for streaming data**

### Python (services/ai, services/worker)

1. Use FastAPI for AI service
2. Use Celery for worker tasks
3. Follow PEP 8 style

---

## Current Development Status

| Module | Status | Notes |
|--------|--------|-------|
| Agent System | Phase 4 Complete | Chat, Sessions, Groups, Share, Export |
| Reference Library | Phase 2 Complete | CRUD, Import/Export, AI Search |
| Admin UI | Redesign Complete | Dark/Light theme, Agent management |

---

## Issues to Fix (from code review)

### HIGH Priority

#### SQL Injection Risk
- **File:** `services/gateway/handler/reference_handler.go:88-94`
- **Issue:** ORDER BY clause built via string concatenation without validation
- **Fix:** Validate `params.SortBy` against whitelist (allowed: `created_at`, `updated_at`, `title`, `year`, `starred`)

```go
allowedSortFields := map[string]bool{
    "created_at": true, "updated_at": true, "title": true, "year": true, "starred": true,
}
if !allowedSortFields[params.SortBy] {
    params.SortBy = "created_at"
}
```

### MEDIUM Priority

#### Stream JSON Parse Failure
- **File:** `apps/web/src/composables/useChatStream.ts:221`
- **Issue:** `JSON.parse(toolCallDelta.function.arguments)` may fail during streaming
- **Fix:** Wrap in try-catch, accumulate string before parsing

#### Stream Message Save Failure
- **File:** `services/gateway/handler/chat_handler.go:608-615`
- **Issue:** Message save only logs error, may lose data on disconnect
- **Fix:** Consider retry mechanism or queue

### LOW Priority

#### Silent Error Swallowing
- **File:** `services/gateway/handler/session_handler.go:143-144`
- **Issue:** Error silently ignored
- **Fix:** Log or return appropriate error

---

## Database Migrations

Run migrations after pulling changes that add new tables/fields:

```sql
-- See docs/agent-system-progress.md for Session fields
-- See docs/reference-library-plan.md for Reference/Collection tables
```

---

## Testing

- Frontend: Vitest unit tests in `apps/web`
- Backend: Go tests in `services/gateway`
- E2E: Playwright

---

## Commit Guidelines

1. Run `pnpm lint` before committing
2. Update Swagger docs if API changes (`pnpm gen:openapi`)
3. Update progress docs if features change