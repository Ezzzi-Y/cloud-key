# Task 2 Report: API Client Layer (React SPA)

**Status:** DONE

**Commit:** `4dd5cda3498b8f30007c773b6082fa37c1b1dfca`

## Summary
Created the full API client layer for the CloudKey React SPA frontend. All 9 files match the Go backend routes defined in `internal/router/router.go`.

## Files Created (9 total)
| File | Purpose |
|------|---------|
| `web/src/types/index.ts` | All TypeScript interfaces and types (ApiResponse, models, request/response types) |
| `web/src/api/client.ts` | Axios instance with JWT interceptor (`ck_token` from localStorage), 401 redirect |
| `web/src/api/auth.ts` | Auth endpoints: login, 2FA verify, TOTP setup/confirm, profile, password change |
| `web/src/api/keys.ts` | Key endpoints: public status/consume, tenant CRUD, export (CSV + JSON) |
| `web/src/api/tenants.ts` | Super admin tenant management: list, create, get, update, reset password |
| `web/src/api/stats.ts` | Dashboard, overview, trends, top-keys, top-ips endpoints |
| `web/src/api/logs.ts` | Usage logs (list + export) and login logs |
| `web/src/api/service-accounts.ts` | Service account management: list, create, toggle, delete |
| `web/src/api/config.ts` | Super admin system configs: get and update |

## Backend Route Alignment
All API paths were verified against `internal/router/router.go`:
- `/api/auth/*` — public auth routes
- `/api/key/status`, `/api/key/consume` — public key endpoints
- `/api/super/*` — super admin routes
- `/api/tenant/*` — tenant admin routes
- Route parameters (`:id`) match backend Gin path params

## Verification
- `npx tsc --noEmit`: PASSED with zero errors
- All imports use `@/` path aliases matching the Vite/tsconfig configuration from Task 1
- JWT key name `ck_token` matches the spec
- Response interceptor unwraps `res.data` so callers receive `ApiResponse<T>` directly

## Concerns
- None. All types and API paths align with the Go backend.
