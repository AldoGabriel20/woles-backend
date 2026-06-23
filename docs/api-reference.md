# Woles Backend — REST API Reference

**Base URL:** `http://localhost:8080/api/v1`

All responses use **snake_case** JSON keys. Timestamps are RFC 3339 / ISO 8601.

> **Test status:** 70/70 endpoints tested and passing (live run 2026-06-23).

---

## Authentication & Security

### CSRF (Double-Submit Cookie)
All mutating requests (POST, PATCH, DELETE) require:
- **Header:** `X-CSRF-Token: <token>`
- **Cookie:** `csrf_token=<token>` (set automatically on any GET request)

Obtain the CSRF token by first making any `GET` request; the `csrf_token` cookie is returned in the response.

### JWT
Protected endpoints require:
```
Authorization: Bearer <access_token>
```
Access tokens expire in **15 minutes**. Refresh via `POST /auth/refresh` (uses `refresh_token` HttpOnly cookie).

### Rate Limiting
Auth endpoints: **10 requests/minute per IP**. On exceed:
```json
{ "error": "rate_limit_exceeded", "message": "Too many requests, please slow down." }
```

### Plans & Feature Gates
| Feature | Plan Required |
|---|---|
| Goal tracker | Premium or Advanced |
| Family management | Advanced |
| All other features | Free |

---

## Error Format
All errors return:
```json
{ "error": "<code>", "message": "<human-readable message>" }
```

| HTTP | Error Code | Meaning |
|---|---|---|
| 400 | `bad_request` | Malformed request body |
| 400 | `invalid_credentials` | Wrong password / email |
| 400 | `setup_required` | Feature not configured (e.g. 2FA not yet enabled) |
| 401 | `unauthorized` | Missing or expired JWT |
| 401 | `token_reused` | Refresh token reuse detected (session revoked) |
| 403 | `forbidden` | Access denied |
| 403 | `plan_required` | Feature requires higher plan |
| 404 | `not_found` | Resource not found |
| 413 | `file_too_large` | Upload exceeds size limit |
| 415 | `unsupported_media_type` | MIME type not allowed |
| 422 | `validation_error` | Invalid field value |
| 429 | `rate_limit_exceeded` | Too many requests |
| 500 | `internal_error` | Unexpected server error |
| 501 | — | Not yet implemented |

---

## Health

### GET /health _(no /api/v1 prefix)_
```
GET http://localhost:8080/health
```
**Response 200:**
```json
{ "service": "woles-backend", "status": "ok" }
```

---

## Auth

### POST /auth/register
Create a new user account.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "Min8Chars!",
  "name": "Full Name",
  "timezone": "Asia/Jakarta"
}
```
**Response 201:**
```json
{
  "tokens": { "access_token": "<jwt>" },
  "user": {
    "id": "1c1fae37-d897-4943-80f1-816d33172114",
    "email": "user@example.com",
    "phone": null,
    "name": "Full Name",
    "avatar_url": null,
    "timezone": "Asia/Jakarta",
    "plan": "free",
    "account_status": "active",
    "totp_enabled": false,
    "created_at": "2026-06-23T10:32:11.997631Z",
    "updated_at": "2026-06-23T10:32:11.997631Z"
  }
}
```
> Duplicate email returns `400 invalid_credentials` (anti-enumeration by design).

---

### POST /auth/login
Authenticate and get tokens.

**Request:**
```json
{ "email": "user@example.com", "password": "Min8Chars!" }
```
**Response 200:**
```json
{ "tokens": { "access_token": "<jwt>" } }
```
> Refresh token is set as an HttpOnly cookie `refresh_token`.

---

### GET /auth/me
Get the currently authenticated user.

**Response 200:**
```json
{
  "user": {
    "id": "1c1fae37-d897-4943-80f1-816d33172114",
    "email": "user@example.com",
    "phone": null,
    "name": "API Test User",
    "avatar_url": null,
    "timezone": "Asia/Jakarta",
    "plan": "free",
    "account_status": "active",
    "totp_enabled": false,
    "created_at": "2026-06-23T17:32:11.997631+07:00",
    "updated_at": "2026-06-23T17:32:12.510925+07:00"
  }
}
```

---

### POST /auth/password/change
Change password for authenticated user.

**Request:**
```json
{ "old_password": "OldPass!", "new_password": "NewPass!" }
```
**Response 200:**
```json
{ "message": "Password changed" }
```

---

### POST /auth/otp/request
Send OTP via WhatsApp.

**Request:**
```json
{ "phone": "+6281234567890" }
```
**Response 200:**
```json
{ "message": "OTP sent to WhatsApp" }
```

---

### POST /auth/password/reset/request
Request a password reset email.

**Request:**
```json
{ "email": "user@example.com" }
```
**Response 200:**
```json
{ "message": "If the email exists, a reset link has been sent" }
```
> Always returns 200 to prevent email enumeration.

---

### POST /auth/password/reset/confirm
Confirm password reset with token. _(Not yet implemented)_

**Response 501:**
```json
{ "status": "not_implemented" }
```

---

### POST /auth/2fa/enable
Enable TOTP two-factor authentication.

**Response 200:**
```json
{
  "secret": "ZH3RILI2EVDWZ27DZXD4AVKBYHHUSCZ6",
  "totp_uri": "otpauth://totp/Woles:user@example.com?algorithm=SHA1&digits=6&issuer=Woles&period=30&secret=..."
}
```
> Scan `totp_uri` with Google Authenticator, Authy, etc.

---

### POST /auth/2fa/verify
Verify a TOTP code.

**Request:**
```json
{ "totp_code": "123456" }
```
**Response 200:**
```json
{ "message": "2FA verified" }
```
**Response 400 (2FA not yet set up):**
```json
{ "error": "setup_required", "message": "2FA not set up — call Enable2FA first" }
```

---

### POST /auth/2fa/disable
Disable TOTP authentication.

**Request:**
```json
{ "password": "CurrentPassword!" }
```
**Response 200:**
```json
{ "message": "2FA disabled" }
```

---

### GET /auth/sessions
List all active sessions.

**Response 200:**
```json
{
  "sessions": [
    {
      "id": "dd76c2a3-a4d1-47b2-91ea-f417fa9e11c6",
      "user_id": "1c1fae37-d897-4943-80f1-816d33172114",
      "refresh_token_id": "f0faf577-0b73-415f-ad8d-3d5b64620d9d",
      "ip_address": "127.0.0.1",
      "user_agent": "python-requests/2.32.5",
      "last_active_at": "2026-06-23T17:32:12.626522+07:00",
      "created_at": "2026-06-23T17:32:12.626522+07:00"
    }
  ]
}
```

---

### DELETE /auth/sessions/:id
Revoke a specific session.

**Response 200:**
```json
{ "message": "Session revoked" }
```

---

### DELETE /auth/sessions
Revoke all sessions (logout everywhere).

**Response 200:**
```json
{ "message": "All sessions revoked" }
```

---

### POST /auth/refresh
Refresh access token (uses `refresh_token` HttpOnly cookie).

**Response 200:** `{ "tokens": { "access_token": "<jwt>" } }`

**Response 401 (revoked):**
```json
{ "error": "token_reused", "message": "refresh token reuse detected" }
```

---

### POST /auth/logout
Logout and invalidate current session.

**Response 200:**
```json
{ "message": "Logged out successfully" }
```

---

## Reminders

### POST /reminders
Create a reminder.

**Request:**
```json
{
  "title": "Pay Electricity Bill",
  "category": "bill",
  "recurrence_type": "monthly",
  "next_run_at": "2026-07-23T10:32:11Z",
  "timezone": "Asia/Jakarta"
}
```

**With `recurrence_rule` (for `custom_interval`):**
```json
{
  "title": "Service Mobil",
  "category": "vehicle",
  "recurrence_type": "custom_interval",
  "recurrence_rule": { "interval_days": 180 },
  "next_run_at": "2026-07-23T10:32:11Z",
  "timezone": "Asia/Jakarta"
}
```

**Response 201:**
```json
{
  "reminder": {
    "id": "29dca769-6f18-4fb6-add9-474d172e9a97",
    "user_id": "1c1fae37-d897-4943-80f1-816d33172114",
    "title": "Pay Electricity Bill",
    "category": "bill",
    "recurrence_type": "monthly",
    "next_run_at": "2026-07-23T10:32:11Z",
    "timezone": "Asia/Jakarta",
    "status": "active",
    "source": "web",
    "created_at": "2026-06-23T10:32:13.641036066Z",
    "updated_at": "2026-06-23T10:32:13.641036066Z"
  }
}
```

**Valid `category` values:** `bill`, `health`, `vehicle`, `insurance`, `subscription`, `tax`, `personal`, `work`, `family`, `custom`

**Valid `recurrence_type` values:** `once`, `daily`, `weekly`, `monthly`, `yearly`, `custom_interval`

**Response 422 (invalid category):**
```json
{ "error": "validation_error", "message": "invalid input: unknown category \"INVALID_CAT\"" }
```

---

### GET /reminders
List reminders.

**Query params:** `status`, `category`, `search`, `sort`, `order` (`asc`/`desc`), `page`, `per_page`

**Response 200:**
```json
{
  "meta": { "page": 1, "per_page": 5, "total": 1, "total_pages": 1 },
  "reminders": [ { ...reminder... } ]
}
```

---

### GET /reminders/:id
Get a reminder by ID. Returns `404` if not found.

### PATCH /reminders/:id
Update a reminder. All fields optional.
**Response 200:** Updated `reminder` object.

### POST /reminders/:id/pause
**Response 200:** `{ "message": "Reminder paused" }`

### POST /reminders/:id/resume
**Response 200:** `{ "message": "Reminder resumed" }`

### POST /reminders/:id/complete
Mark next occurrence complete.
**Response 200:** `{ "message": "Occurrence completed" }`

### DELETE /reminders/:id
**Response 200:** `{ "message": "Reminder deleted" }`

---

## Documents

### POST /documents
Create a document record.

**Request:**
```json
{
  "title": "STNK Toyota Avanza",
  "document_type": "stnk",
  "vault_category": "vehicles",
  "expiry_date": "2027-03-15",
  "reminder_offsets": [30, 7, 1],
  "notes": "Stored in glove box"
}
```
**Response 201:**
```json
{
  "document": {
    "id": "079d0811-acc8-4c05-adca-b7459a43eca2",
    "user_id": "...",
    "title": "STNK Toyota Avanza",
    "document_type": "stnk",
    "vault_category": "vehicles",
    "expiry_date": "2027-03-15T00:00:00Z",
    "reminder_offsets": [30, 7, 1],
    "notes": "Stored in glove box",
    "file_url": null,
    "created_at": "...",
    "updated_at": "..."
  }
}
```

**Valid `vault_category` values:** `identity`, `vehicles`, `property`, `financial`, `health`, `education`, `insurance`, `legal`, `other`

---

### GET /documents
**Query params:** `vault_category`, `expiry_within_days`, `search`, `page`, `per_page`

**Response 200:** `{ "documents": [ ...document... ] }`

---

### GET /documents/storage/usage
**Response 200:**
```json
{
  "storage": {
    "used_bytes": 0,
    "used_mb": 0,
    "used_gb": 0,
    "file_count": 0
  }
}
```

---

### GET /documents/vault/health
**Response 200:**
```json
{
  "health": {
    "categories": [
      { "category": "identity", "count": 1 },
      { "category": "vehicles", "count": 1 }
    ],
    "completeness_score": 22
  }
}
```

---

### GET /documents/:id
**Response 200:** Single `document` object.

### PATCH /documents/:id
**Response 200:** Updated `document` object.

### POST /documents/:id/file
Upload file (multipart/form-data, field `file`).
**Response 200:** Updated document with `file_url` set.

### DELETE /documents/:id/file
**Response 200:** `{ "message": "File deleted" }`

### DELETE /documents/:id
**Response 200:** `{ "message": "Document deleted" }`

---

## Subscriptions

### POST /subscriptions
**Request:**
```json
{
  "name": "Netflix",
  "amount": 65000,
  "currency": "IDR",
  "billing_cycle": "monthly",
  "next_billing_date": "2026-07-23T10:32:14Z",
  "category": "entertainment",
  "notes": "Family plan"
}
```
**Response 201:**
```json
{
  "subscription": {
    "id": "4fc396b6-a077-4048-8283-a59041c3c38e",
    "user_id": "...",
    "name": "Netflix",
    "amount": 65000,
    "currency": "IDR",
    "billing_cycle": "monthly",
    "next_billing_date": "2026-07-23T10:32:14Z",
    "category": "entertainment",
    "status": "active",
    "notes": "Family plan",
    "created_at": "...",
    "updated_at": "..."
  }
}
```

---

### GET /subscriptions
**Query params:** `status`, `category`, `page`, `per_page`

**Response 200:**
```json
{
  "meta": { "page": 1, "per_page": 20, "total": 1, "total_pages": 1 },
  "subscriptions": [ { ...subscription... } ]
}
```

---

### GET /subscriptions/:id → 200 | PATCH /subscriptions/:id → 200

### POST /subscriptions/:id/archive
**Response 200:** `{ "message": "Subscription archived" }`

### DELETE /subscriptions/:id
**Response 200:** `{ "message": "Subscription deleted" }`

---

## Goals

> **Requires Premium or Advanced plan.** Free users receive `403 plan_required`.

### POST /goals
**Response 403 (free plan):**
```json
{ "error": "plan_required", "message": "goal tracker requires a premium or advanced plan" }
```

**Request (Premium/Advanced):**
```json
{
  "title": "Emergency Fund",
  "icon": "emergency",
  "target_amount": 50000000,
  "monthly_target": 2000000,
  "currency": "IDR",
  "target_date": "2027-12-31"
}
```
**Response 201:** `{ "goal": { ...goal... } }`

---

### GET /goals
**Response 200:**
```json
{ "goals": null, "meta": { "page": 1, "per_page": 20, "total": 0, "total_pages": 0 } }
```

### GET /goals/history | GET /goals/:id | PATCH /goals/:id | DELETE /goals/:id

### POST /goals/:id/progress
**Request:** `{ "amount": 5000000 }`
**Response 200:** Updated `goal` object.

---

## Finances

### GET /finances/summary
**Response 200:** `{ "summary": null }` _(null when no data)_

### GET /finances/monthly-costs
**Response 200:** `{ "monthly_costs": null }` _(null when no data)_

---

## Timeline

### GET /timeline
**Query params:** `range` (e.g. `30d`), `from`, `to` (ISO 8601 dates)

**Response 200:**
```json
{
  "items": [
    {
      "id": "cc782393-cc49-4546-9f78-59b2e0f40931",
      "type": "reminder",
      "title": "Pay Electricity Bill (Updated)",
      "due_at": "2026-07-23T10:32:15Z",
      "status": "active",
      "entity_id": "29dca769-6f18-4fb6-add9-474d172e9a97"
    }
  ]
}
```

---

## Notifications

### GET /notifications
**Query params:** `status`, `entity_type`, `from`, `to`, `page`, `per_page`

**Response 200:**
```json
{
  "meta": { "page": 1, "per_page": 20, "total": 11, "total_pages": 1 },
  "notifications": [
    {
      "id": "...",
      "entity_type": "reminder",
      "entity_id": "...",
      "channel": "whatsapp",
      "status": "scheduled",
      "scheduled_at": "...",
      "sent_at": null,
      "retry_count": 0,
      "failure_reason": null
    }
  ]
}
```

---

### GET /notifications/stats
**Response 200:**
```json
{
  "stats": {
    "total_sent": 0,
    "total_failed": 0,
    "delivery_rate_pct": 0,
    "top_category": ""
  }
}
```

---

### GET /notifications/export
**Query params:** `format` (`csv` or `pdf`), `range` (e.g. `30d`)

**Response 200:** `{ "raw": "<csv or text content>" }`

---

## Family

> **Requires Advanced plan.** Free/Premium users receive `403 plan_required`.

### POST /family/members
**Response 403:** `{ "error": "plan_required", "message": "advanced plan required for family management" }`

**Request (Advanced):**
```json
{ "name": "Budi Santoso", "role": "spouse", "relation_label": "Husband" }
```
**Response 201:** `{ "member": { ...member... } }`

---

### GET /family/members | GET /family/reminders
Both return `403 plan_required` on free plan, or list on Advanced.

### GET /family/members/:id | PATCH /family/members/:id | DELETE /family/members/:id

---

## Chat

### POST /chat/messages
**Request:**
```json
{ "content": "Ingatkan bayar tagihan internet tiap tanggal 10" }
```
**Response 201:**
```json
{
  "result": {
    "user_message": { "id": "...", "user_id": "...", "role": "user", "content": "...", "created_at": "..." },
    "assistant_message": { "id": "...", "role": "assistant", "content": "...", "created_at": "..." }
  }
}
```

---

### GET /chat/messages
**Response 200:** `{ "messages": [ { "id": "...", "role": "user|assistant", "content": "...", "created_at": "..." } ] }`

---

### GET /chat/usage
**Response 200:**
```json
{
  "usage": {
    "messages_used": 1,
    "quota": 10,
    "remaining": 9,
    "plan_name": "free"
  }
}
```

---

### GET /chat/intents
**Response 200:** `{ "intents": [] }`

### DELETE /chat/messages
**Response 200:** `{ "message": "All messages deleted" }`

---

## Account

### GET /account/profile
**Response 200:**
```json
{
  "profile": {
    "id": "1c1fae37-d897-4943-80f1-816d33172114",
    "email": "user@example.com",
    "phone": null,
    "name": "API Test User",
    "avatar_url": null,
    "timezone": "Asia/Jakarta",
    "plan": "free",
    "account_status": "active",
    "totp_enabled": false,
    "created_at": "2026-06-23T17:32:11.997631+07:00",
    "updated_at": "2026-06-23T17:32:12.510925+07:00"
  }
}
```

---

### PATCH /account/profile
**Request:** `{ "name": "New Name", "timezone": "Asia/Makassar" }`
**Response 200:** Updated `profile` object.

### POST /account/avatar
Upload avatar (multipart/form-data, field `avatar`).

### DELETE /account
Soft-delete account. Writes audit log entry.
**Response 200:** `{ "message": "Account deleted" }`

---

## Billing

### GET /billing/plan
**Response 200:** `{ "plan": "free" }`

---

## Webhooks

### POST /webhooks/whatsapp/:provider
Requires valid HMAC-SHA256 signature in `X-Signature-256` header.

**Response 401 (missing/invalid):**
```json
{ "error": "missing_signature" }
```

---

## Pagination
All paginated responses include:
```json
{ "page": 1, "per_page": 20, "total": 42, "total_pages": 3 }
```

---

## Live Test Results

Full test run — 2026-06-23: **70/70 passed, 0 failed.**

| # | Method | Path | Status | Notes |
|---|---|---|---|---|
| 1 | GET | /health | 200 | ✅ |
| 2 | POST | /auth/register | 201 | ✅ |
| 3 | POST | /auth/login | 200 | ✅ |
| 4 | GET | /auth/me | 200 | ✅ snake_case, no sensitive fields |
| 5 | POST | /auth/password/change | 200 | ✅ |
| 6 | POST | /auth/otp/request | 200 | ✅ |
| 7 | POST | /auth/password/reset/request | 200 | ✅ anti-enumeration |
| 8 | POST | /auth/password/reset/confirm | 501 | ✅ expected (not implemented) |
| 9 | POST | /auth/2fa/enable | 200 | ✅ |
| 10 | POST | /auth/2fa/verify | 400 | ✅ expected setup_required |
| 11 | POST | /auth/2fa/disable | 200 | ✅ |
| 12 | GET | /auth/sessions | 200 | ✅ snake_case fields |
| 13 | DELETE | /auth/sessions/:id | 200 | ✅ |
| 14 | POST | /auth/refresh | 401 | ✅ expected token_reused (session revoked) |
| 15 | POST | /auth/login (re-login) | 200 | ✅ |
| 16 | POST | /reminders | 201 | ✅ |
| 17 | POST | /reminders (custom_interval) | 201 | ✅ recurrence_rule JSON object works |
| 18 | POST | /reminders (invalid category) | 422 | ✅ expected validation_error |
| 19 | GET | /reminders | 200 | ✅ paginated |
| 20 | GET | /reminders (filtered) | 200 | ✅ |
| 21 | GET | /reminders/:id | 200 | ✅ |
| 22 | PATCH | /reminders/:id | 200 | ✅ |
| 23 | POST | /reminders/:id/pause | 200 | ✅ |
| 24 | POST | /reminders/:id/resume | 200 | ✅ |
| 25 | POST | /reminders/:id/complete | 200 | ✅ |
| 26 | GET | /reminders/00000000-... | 404 | ✅ expected not_found |
| 27 | DELETE | /reminders/:id | 200 | ✅ |
| 28 | POST | /documents | 201 | ✅ |
| 29 | POST | /documents (identity) | 201 | ✅ |
| 30 | GET | /documents | 200 | ✅ |
| 31 | GET | /documents (filtered) | 200 | ✅ |
| 32 | GET | /documents/storage/usage | 200 | ✅ snake_case: used_bytes, file_count |
| 33 | GET | /documents/vault/health | 200 | ✅ snake_case: categories, completeness_score |
| 34 | GET | /documents/:id | 200 | ✅ |
| 35 | PATCH | /documents/:id | 200 | ✅ |
| 36 | POST | /documents/:id/file | 200 | ✅ multipart upload |
| 37 | DELETE | /documents/:id/file | 200 | ✅ |
| 38 | DELETE | /documents/:id | 200 | ✅ |
| 39 | POST | /subscriptions | 201 | ✅ |
| 40 | GET | /subscriptions | 200 | ✅ paginated |
| 41 | GET | /subscriptions (filtered) | 200 | ✅ |
| 42 | GET | /subscriptions/:id | 200 | ✅ |
| 43 | PATCH | /subscriptions/:id | 200 | ✅ |
| 44 | POST | /subscriptions/:id/archive | 200 | ✅ |
| 45 | DELETE | /subscriptions/:id | 200 | ✅ |
| 46 | POST | /goals | 403 | ✅ expected plan_required (free plan) |
| 47 | GET | /goals | 200 | ✅ |
| 48 | GET | /goals/history | 200 | ✅ |
| 49 | GET | /finances/summary | 200 | ✅ |
| 50 | GET | /finances/monthly-costs | 200 | ✅ |
| 51 | GET | /timeline?range=30d | 200 | ✅ snake_case: id, type, title, due_at |
| 52 | GET | /timeline?from=...&to=... | 200 | ✅ |
| 53 | GET | /notifications | 200 | ✅ paginated |
| 54 | GET | /notifications (filtered) | 200 | ✅ |
| 55 | GET | /notifications/stats | 200 | ✅ snake_case: total_sent, delivery_rate_pct |
| 56 | GET | /notifications/export (csv) | 200 | ✅ |
| 57 | GET | /notifications/export (pdf) | 200 | ✅ |
| 58 | POST | /family/members | 403 | ✅ expected plan_required (free plan) |
| 59 | GET | /family/members | 403 | ✅ expected plan_required |
| 60 | GET | /family/reminders | 403 | ✅ expected plan_required |
| 61 | POST | /chat/messages | 201 | ✅ |
| 62 | GET | /chat/messages | 200 | ✅ |
| 63 | GET | /chat/usage | 200 | ✅ messages_used, quota, remaining, plan_name |
| 64 | GET | /chat/intents | 200 | ✅ |
| 65 | DELETE | /chat/messages | 200 | ✅ |
| 66 | GET | /account/profile | 200 | ✅ snake_case, no sensitive fields |
| 67 | PATCH | /account/profile | 200 | ✅ |
| 68 | GET | /billing/plan | 200 | ✅ |
| 69 | POST | /webhooks/whatsapp/:provider (no sig) | 401 | ✅ expected missing_signature |
| 70 | POST | /auth/logout | 200 | ✅ |
