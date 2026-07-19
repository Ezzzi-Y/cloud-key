# Task 3 Report — Auth Context, Route Guards, Admin Layouts (Frontend)

**Status:** DONE

**Commit:** `5ab26bf`

## Files Created/Updated

1. `web/src/hooks/useAuth.tsx` — AuthProvider context with `useAuth` hook. JWT stored in localStorage key `ck_token`. On mount, probes `/super/profile` then `/tenant/profile` to restore session. Provides `login()`, `logout()`, and state (`token`, `role`, `tenantId`, `username`, `isAuthenticated`).

2. `web/src/components/RequireAuth.tsx` — Route guard component. Checks `isAuthenticated` (redirect to `/login` if false) and `role` match (redirects to correct role path on mismatch). Uses `<Outlet />` for nested routing.

3. `web/src/layouts/SuperAdminLayout.tsx` — Sidebar layout with nav: 仪表盘, 租户管理, 平台配置, 个人设置. Responsive: sidebar on md+, Sheet drawer on mobile. Logout button with username display.

4. `web/src/layouts/TenantAdminLayout.tsx` — Sidebar layout with nav: 仪表盘, Key 管理, 校验与扣减, 使用记录, 服务账号, 个人设置. Same responsive pattern.

5. `web/src/App.tsx` — Rewritten with full lazy-loaded routing:
   - `/login` → Login page
   - `/super/*` → RequireAuth(role="super_admin") + SuperAdminLayout with 5 nested routes
   - `/tenant/*` → RequireAuth(role="tenant_admin") + TenantAdminLayout with 6 nested routes
   - `/` and `*` → Navigate to `/login`

## TypeScript Check

`npx tsc --noEmit` passed with **zero errors**. The lazy imports for page components (`@/pages/Login`, `@/pages/super/Dashboard`, etc.) resolve correctly since TypeScript defers resolution for dynamic `import()` expressions. These page components will be created in Tasks 4-6.

## Expected Follow-up

Pages referenced in App.tsx (Login, Dashboard, Tenants, KeyManagement, etc.) to be created in Tasks 4-6. The auth and layout scaffolding is complete and type-safe.
