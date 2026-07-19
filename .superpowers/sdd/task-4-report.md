# Task 4 Report: Login Page with Multi-Step TOTP Flow (Frontend)

**Status:** Complete

**Commit:** `964dd5d` — feat: add login page with multi-step TOTP flow

**File:** `web/src/pages/Login.tsx` (155 lines)

**TypeScript Check:** Passed with zero errors (`npx tsc --noEmit`)

**What was built:**
- `web/src/pages/Login.tsx` — a multi-step login page with three states:
  - **Step 1 (`form`):** Username + password form calling `POST /api/auth/login`
    - `require_totp === true` advances to Step 2
    - `need_setup === true` automatically calls `setupTOTPInit` and shows QR code (Step 3)
  - **Step 2 (`totp`):** TOTP code input calling `POST /api/auth/verify-2fa`
    - On success, stores JWT via `useAuth().login(token, role, tenant_id, username)` and navigates to `/super/` or `/tenant/`
  - **Step 3 (`setup`):** TOTP setup wizard with QR code display, calling `POST /api/auth/totp/confirm-init`
    - QR code rendered via `api.qrserver.com`, secret shown as fallback
    - On success, stores JWT and navigates

**Error handling:** All API errors and network failures show Chinese toast notifications via `useToast()`.

**Navigation:** After successful auth, navigates to `/super/` for `super_admin` role, `/tenant/` for `tenant_admin` role.
