# Woles — Agent Task List

> This file is the single source of truth for all implementation tasks.
> Reference files: `docs/system-design.md`, `docs/woles-prd.md`, `docs/ui/*.png`, `docs/logo/`.
> The backend and frontend live in **separate repositories**.
> - Backend repo: `woles-backend` (Go + Prabogo, Hexagonal Architecture)
> - Frontend repo: `woles-frontend` (Next.js 15, TypeScript, Tailwind CSS)

---

## Repository Setup

### TASK-001 — Initialize backend repository
**Repo:** `woles-backend`

- Run `prabogo init woles-backend` to scaffold a new Prabogo project.
- Set Go module name to `github.com/woles/woles-backend`.
- Confirm the generated structure matches `docs/system-design.md` Section 5 exactly:
  - `cmd/main.go`
  - `internal/domain/`, `internal/application/`, `internal/port/`, `internal/adapter/`
  - `internal/migration/`, `internal/model/`, `utils/`, `tests/`
- Add domains: `reminder`, `document`, `subscription`, `goal`, `timeline`, `notification`, `billing`, `identity`, `whatsapp`, `family`, `chat`.
- Add application packages for each domain listed above plus `intent`.
- Create `.env.example` with all required variables:
  - `DATABASE_URL`, `REDIS_URL`, `RABBITMQ_URL`
  - `JWT_PRIVATE_KEY_PATH`, `JWT_PUBLIC_KEY_PATH`
  - `WHATSAPP_PROVIDER`, `WHATSAPP_WEBHOOK_SECRET`
  - `AI_PROVIDER`, `AI_API_KEY`
  - `STORAGE_ENDPOINT`, `STORAGE_BUCKET`, `STORAGE_ACCESS_KEY`, `STORAGE_SECRET_KEY`
  - `PAYMENT_PROVIDER`, `PAYMENT_SECRET_KEY`, `PAYMENT_WEBHOOK_SECRET`
  - `APP_ENV`, `APP_PORT`, `APP_BASE_URL`
- Create `Makefile` with targets: `run`, `build`, `test`, `migrate-up`, `migrate-down`, `generate-mock`, `lint`.
- Add `.gitignore`: ignore `.env`, `bin/`, compiled binaries, private keys.
- Create `docker-compose.yml` with services: `api`, `worker`, `scheduler`, `postgres`, `redis`, `rabbitmq`, `minio`.
  - Postgres: port 5432, volume `pgdata`.
  - Redis: port 6379.
  - RabbitMQ: port 5672 + management UI 15672.
  - MinIO: port 9000 + console 9001.
- Create `Dockerfile` for the backend binary (multi-stage build, non-root user, `scratch` or `alpine` final image).

---

### TASK-002 — Initialize frontend repository
**Repo:** `woles-frontend`

- Run `npx create-next-app@latest woles-frontend --typescript --tailwind --app --src-dir=false --import-alias="@/*"`.
- Remove boilerplate: clear `app/page.tsx`, delete default CSS content.
- Install dependencies:
  - `next-themes`, `clsx`, `tailwind-merge`
  - `@tanstack/react-query` (v5), `axios`
  - `react-hook-form`, `zod`, `@hookform/resolvers`
  - `recharts` (for spending trend and progress charts)
  - `lucide-react` (icons)
  - `jose` (JWT decode on client, no verification)
  - `js-cookie`
  - `next/font` (already built-in)
- Install dev dependencies: `@types/js-cookie`, `prettier`, `eslint-config-prettier`.
- Create `.env.local.example`:
  - `NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1`
  - `NEXT_PUBLIC_APP_URL=http://localhost:3000`
- Create `tailwind.config.ts` with the exact design tokens from `docs/system-design.md` Section 14.2:
  - Colors: `primary: '#003527'`, `primary-container: '#064e3b'`, `surface: '#f8faf6'`, etc.
  - Font families: `display: ['Geist', 'sans-serif']`, `body: ['Inter', 'sans-serif']`.
  - Border radius: `sm: '4px'`, `DEFAULT: '8px'`, `md: '12px'`, `lg: '16px'`, `full: '9999px'`.
- Create `styles/tokens.css` with all CSS custom properties matching the design token tables in Section 14.2.
- Create `styles/globals.css` importing `tokens.css` and Tailwind directives.
- Set up `next/font/google` in `app/layout.tsx` loading Geist (weights 600, 700) and Inter (weights 400, 500).
- Add `.gitignore`: ignore `.env.local`, `.next/`, `node_modules/`, `out/`.
- Create `Makefile` with targets: `dev`, `build`, `lint`, `type-check`.

---

## Backend — Database Migrations

### TASK-003 — Write all PostgreSQL migration files
**Repo:** `woles-backend`

Create migration files under `internal/migration/` using Prabogo's migration generator. Each migration must be numbered sequentially (`0001_`, `0002_`, etc.) and include both `up` and `down` SQL.

Create the following tables in order (respecting FK dependencies):

**0001 — users**
- All fields from `docs/system-design.md` Section 7.1.
- Indexes: `UNIQUE(email)`, `UNIQUE(phone)`, index on `account_status`, index on `locked_until`.

**0002 — refresh_tokens**
- All fields from Section 8.1 `refresh_tokens` table.
- Index on `(user_id, revoked_at)`, index on `family_id`.

**0003 — user_sessions**
- All fields from Section 7.9.
- Index on `(user_id, last_active_at)`.

**0004 — whatsapp_identities**
- All fields from Section 7.2.
- `UNIQUE(phone, provider)`, index on `user_id`.

**0005 — family_members**
- All fields from Section 7.10.
- Index on `owner_user_id`.

**0006 — reminders**
- All fields from Section 7.3.
- Indexes: `(user_id, status)`, `(next_run_at)` where status = 'active'.

**0007 — reminder_occurrences**
- All fields from Section 7.4.
- Indexes: `(reminder_id)`, `(scheduled_at, status)`, `(user_id, status)`.

**0008 — documents**
- All fields from Section 7.5.
- Indexes: `(user_id, vault_category)`, `(expiry_date)` where status = 'active', `(family_member_id)`.

**0009 — subscriptions**
- All fields from Section 7.6.
- Indexes: `(user_id, status)`, `(next_billing_at)` where status = 'active'.

**0010 — goals**
- All fields from Section 7.7.
- Index on `(user_id, status)`.

**0011 — notifications**
- All fields from Section 7.8.
- Indexes: `(user_id, status)`, `(scheduled_at, status)`, `UNIQUE(idempotency_key)`.

**0012 — inbound_messages**
- All fields from Section 7.13.
- `UNIQUE(provider_message_id)`, index on `(user_id, processing_status)`.

**0013 — chat_messages**
- All fields from Section 7.11.
- Index on `(user_id, created_at DESC)`.

**0014 — chat_usage**
- All fields from Section 7.12.
- `UNIQUE(user_id, month)`.

**0015 — audit_logs**
- All fields from Section 12.8.
- Indexes: `(user_id, created_at DESC)`, `(entity_type, entity_id)`, `(action)`.

**0016 — plans_and_usage_limits**
- Table `plans`: `id uuid PK`, `name varchar UNIQUE`, `price_idr numeric`, `reminder_limit int`, `document_limit int`, `subscription_limit int`, `goal_tracker bool`, `timeline bool`, `family_account bool`, `ai_chat bool`, `ai_chat_quota int`, `ocr bool`, `created_at timestamp`.
- Table `usage_limits`: `id uuid PK`, `user_id uuid FK users.id UNIQUE`, `reminders_used int DEFAULT 0`, `documents_used int DEFAULT 0`, `subscriptions_used int DEFAULT 0`, `updated_at timestamp`.
- Seed `plans` with Free (0, limits: 20/5/5), Premium (39000, unlimited), Advanced (99000, unlimited + family + AI + OCR).

---

## Backend — Domain Models

### TASK-004 — Define domain entities and value objects
**Repo:** `woles-backend`

In `internal/domain/{domain_name}/`, create Go structs and interfaces for each domain. Do **not** import any infrastructure package.

**identity**
- `User` struct: all fields from table 7.1, plus computed methods `IsLocked() bool`, `CanLogin() bool`.
- `WhatsAppIdentity` struct from table 7.2.

**reminder**
- `Reminder` struct from table 7.3.
- `RecurrenceType` string enum: `one_time`, `daily`, `weekly`, `monthly`, `yearly`, `custom_interval`.
- `ReminderStatus` string enum: `active`, `paused`, `archived`.
- `ReminderOccurrence` struct from table 7.4.
- Method `CalculateNextRunAt(from time.Time) time.Time` on `Reminder`.

**document**
- `Document` struct from table 7.5.
- `VaultCategory` string enum: `vehicles`, `identity`, `insurance`, `financials`, `other`.
- `StorageType` string enum: `physical`, `digital`, `scan_verified`.
- Method `DaysUntilExpiry() int` on `Document`.
- Method `ExpiryRisk() string` returning `safe`, `upcoming` (≤30d), `urgent` (≤7d), `expired`.

**subscription**
- `Subscription` struct from table 7.6.
- `BillingCycle` string enum: `monthly`, `yearly`, `custom`.

**goal**
- `Goal` struct from table 7.7.
- `GoalIcon` string enum: `love`, `emergency`, `vehicle`, `home`, `travel`, `other`.
- Method `ProgressPercent() float64` on `Goal`.
- Method `Remaining() float64` on `Goal`.

**notification**
- `Notification` struct from table 7.8.
- `NotificationStatus` enum: `scheduled`, `sending`, `sent`, `failed`, `canceled`.
- Method `BuildIdempotencyKey() string` returning `notification:{id}:channel:{channel}`.

**family**
- `FamilyMember` struct from table 7.10.
- `MemberRole` string enum: `primary`, `spouse`, `parent`, `child`, `other`.

**chat**
- `ChatMessage` struct from table 7.11.
- `ChatUsage` struct from table 7.12.
- `MessageRole` string enum: `user`, `assistant`.

**billing**
- `Plan` struct.
- `UsageLimit` struct.
- Method `IsWithinLimit(resource string, current int) bool`.

---

## Backend — Outbound Ports

### TASK-005 — Define all outbound port interfaces
**Repo:** `woles-backend`

In `internal/port/outbound/`, create Go interfaces. No implementations here.

**database/**
- `UserRepository`: `Create`, `FindByID`, `FindByEmail`, `FindByPhone`, `UpdateFailedLoginCount`, `UpdateLockedUntil`, `UpdatePlan`, `Delete`, `SoftDelete`.
- `RefreshTokenRepository`: `Create`, `FindByID`, `FindByFamilyID`, `Revoke`, `RevokeAllForUser`, `RevokeFamily`.
- `UserSessionRepository`: `Create`, `FindAllByUser`, `FindByID`, `UpdateLastActive`, `Delete`, `DeleteAllForUser`.
- `ReminderRepository`: `Create`, `FindByID`, `FindAllByUser` (with pagination + filters), `Update`, `SoftDelete`, `FindDueOccurrences` (for scheduler).
- `ReminderOccurrenceRepository`: `Create`, `FindByReminderID`, `UpdateStatus`, `ClaimForSending` (FOR UPDATE SKIP LOCKED).
- `DocumentRepository`: `Create`, `FindByID`, `FindAllByUser` (pagination + vault_category filter), `Update`, `SoftDelete`, `FindExpiringSoon`, `GetStorageUsage`, `GetVaultHealth`.
- `SubscriptionRepository`: `Create`, `FindByID`, `FindAllByUser`, `Update`, `SoftDelete`, `GetMonthlyCostSummary`.
- `GoalRepository`: `Create`, `FindByID`, `FindAllByUser`, `FindActiveGoal`, `UpdateProgress`, `Update`, `SoftDelete`.
- `NotificationRepository`: `Create`, `FindByID`, `FindAllByUser` (pagination + filters), `ClaimDue` (FOR UPDATE SKIP LOCKED), `UpdateStatus`, `IncrementRetry`, `GetStats`, `ExportRange`.
- `FamilyMemberRepository`: `Create`, `FindByID`, `FindAllByOwner`, `Update`, `Delete`.
- `ChatMessageRepository`: `Create`, `FindAllByUser` (pagination), `DeleteAllByUser`.
- `ChatUsageRepository`: `GetOrCreate`, `Increment`, `GetQuota`.
- `InboundMessageRepository`: `Create`, `FindByProviderMessageID`, `UpdateStatus`.
- `AuditLogRepository`: `Create`, `FindAllByUser`.
- `UsageLimitRepository`: `Get`, `Increment`, `Decrement`, `IsWithinLimit`.
- `TimelineRepository`: `GetTimelineItems(userID, from, to time.Time, page, perPage int)`.

**cache/**
- `RateLimiter`: `Allow(key string, limit int, window time.Duration) (bool, int, error)`.
- `IdempotencyStore`: `Set(key string, ttl time.Duration) error`, `Exists(key string) (bool, error)`.
- `OTPStore`: `Set(phone, hashedOTP string, ttl time.Duration) error`, `Get(phone string) (string, error)`, `Delete(phone string) error`.
- `SessionTokenCache`: `Set(key, value string, ttl time.Duration) error`, `Get(key string) (string, error)`, `Delete(key string) error`.

**whatsapp/**
- `WhatsAppSender`: `SendMessage(to, templateName string, params map[string]string) (providerMsgID string, error)`.
- `WhatsAppVerifier`: `VerifySignature(payload []byte, signature string) bool`.

**ai/**
- `IntentExtractor`: `Extract(text, lang string) (*IntentResult, error)`.
- `IntentResult` struct: `Intent string`, `Confidence float64`, `Payload map[string]interface{}`, `RawText string`.

**storage/**
- `FileStore`: `Upload(key, mimeType string, data io.Reader) (url string, error)`, `Delete(key string) error`, `SignedURL(key string, ttl time.Duration) (string, error)`.

**payment/**
- `PaymentProvider`: `CreateCheckout(plan, userID string) (checkoutURL string, error)`, `VerifyWebhook(payload []byte, signature string) (*PaymentEvent, error)`.

---

## Backend — Application Services

### TASK-006 — Identity application service
**Repo:** `woles-backend`

In `internal/application/identity/service.go`:

- `RegisterWithEmail(email, password, name, timezone string) (*User, error)`:
  - Validate email format, phone format, password min length 8.
  - Check email uniqueness (return generic error to prevent enumeration).
  - Hash password with argon2id (memory=65536KB, iterations=3, parallelism=2, saltLen=16, keyLen=32).
  - Create user with `account_status=active`, `plan=free`.
  - Create `usage_limits` row with zero counters.
  - Write audit log: `action=register`.
  - Return user (without password_hash in response).

- `LoginWithEmail(email, password, ip, userAgent string) (*TokenPair, error)`:
  - Find user by email. If not found, run argon2id verify against dummy hash to normalize response time (prevent timing attack).
  - Check `IsLocked()`. If locked, return generic error with `locked_until`.
  - Verify password with constant-time argon2id compare.
  - On failure: increment `failed_login_count`. If count ≥ 5, set `locked_until = now + 15min`. Write audit log: `action=login_failed`.
  - On success: reset `failed_login_count=0`. Generate access token (RS256 JWT, 15min). Generate refresh token (32 random bytes, base64url). Hash refresh token with bcrypt. Insert `refresh_tokens` row with `family_id`. Create `user_sessions` row. Write audit log: `action=login`. Return `TokenPair{AccessToken, RefreshToken}`.

- `RefreshToken(rawRefreshToken, ip, userAgent string) (*TokenPair, error)`:
  - Hash the provided token with bcrypt and find matching `refresh_tokens` row.
  - If not found or `revoked_at IS NOT NULL`: revoke entire family (token theft detected). Write audit log: `action=token_reuse_detected`. Return error.
  - If `expires_at` passed: return error.
  - Revoke old token. Generate new token pair. Create new `refresh_tokens` row with same `family_id`. Update `user_sessions.last_active_at`. Return new pair.

- `Logout(refreshTokenID, userID string) error`: Mark refresh token revoked. Write audit log: `action=logout`.

- `LogoutAllSessions(userID string) error`: Revoke all refresh tokens for user. Delete all `user_sessions`. Write audit log.

- `RequestOTP(phone string) error`: Generate 6-digit OTP. Hash it. Store in Redis with 5-minute TTL key `otp:{phone}`. Send via WhatsApp provider.

- `VerifyOTP(phone, otp string) (*TokenPair, error)`: Retrieve hash from Redis. Compare constant-time. If match: delete Redis key. Find or create user by phone. Issue token pair.

- `ChangePassword(userID, oldPassword, newPassword string) error`: Verify old password. Hash new password. Update `password_hash`. Revoke all refresh tokens. Write audit log.

- `Enable2FA(userID string) (*TOTPSetup, error)`: Generate TOTP secret (base32). Encrypt with AES-256-GCM using app secret key. Store encrypted secret in `users.totp_secret`. Return `{Secret, QRCodeURL}` for user to scan.

- `Verify2FA(userID, totpCode string) error`: Decrypt stored secret. Validate TOTP code with ±1 window. If valid, set `totp_enabled=true`. Write audit log.

- `Disable2FA(userID, password string) error`: Verify password. Set `totp_secret=NULL`, `totp_enabled=false`. Write audit log.

- `GetActiveSessions(userID string) ([]*UserSession, error)`.

- `RevokeSession(sessionID, userID string) error`: Verify ownership. Revoke refresh token linked to session. Delete session row.

---

### TASK-007 — Reminder application service
**Repo:** `woles-backend`

In `internal/application/reminder/service.go`:

- `CreateReminder(userID string, req CreateReminderRequest) (*Reminder, error)`:
  - Validate plan limits: call `UsageLimitRepository.IsWithinLimit(userID, "reminders")`. If at limit for free plan, return 403 with plan limit error.
  - Validate `recurrence_type` is one of the allowed values.
  - Validate `next_run_at` is in the future for new reminders.
  - Sanitize `title` (strip HTML, max 200 chars) and `original_text` (max 4000 chars).
  - Insert reminder row with `status=active`.
  - Increment usage counter.
  - Calculate first occurrence: create `reminder_occurrences` row with `status=scheduled`.
  - Create `notifications` row with `scheduled_at = occurrence.scheduled_at`, `idempotency_key = notification:{notif_id}:channel:whatsapp`, `status=scheduled`.
  - Publish `reminder.created` event to RabbitMQ.
  - Write audit log.

- `GetReminders(userID string, filter ReminderFilter, page, perPage int) (*PaginatedResult, error)`: Query with ownership check. Apply status, category filters. Sort by `next_run_at ASC`. Return paginated result.

- `GetReminderByID(userID, reminderID string) (*Reminder, error)`: Check ownership. Return 404 if not found or not owned.

- `UpdateReminder(userID, reminderID string, req UpdateReminderRequest) (*Reminder, error)`: Ownership check. Sanitize inputs. Recalculate next occurrence if recurrence changed. Publish `reminder.updated`.

- `DeleteReminder(userID, reminderID string) error`: Ownership check. Soft delete. Cancel pending notifications for this reminder. Decrement usage counter. Write audit log.

- `PauseReminder(userID, reminderID string) error`: Set `status=paused`. Cancel pending notifications.

- `ResumeReminder(userID, reminderID string) error`: Set `status=active`. Recalculate next occurrence. Create new notification.

- `CompleteOccurrence(userID, reminderID string) error`: Mark latest occurrence as `done`. Calculate next occurrence if recurring. Create next notification. Write audit log.

- `CalculateNextRunAt(reminder *Reminder, from time.Time) time.Time`: Pure function. Handles all recurrence types. For `custom_interval`: read `recurrence_rule.interval_days` from JSONB.

---

### TASK-008 — Document application service
**Repo:** `woles-backend`

In `internal/application/document/service.go`:

- `CreateDocument(userID string, req CreateDocumentRequest) (*Document, error)`:
  - Check plan limit for documents.
  - Validate `document_type` against allowed values. Validate `expiry_date` is a valid date.
  - Sanitize `title` (max 200 chars), `notes` (max 2000 chars).
  - Derive `vault_category` from `document_type` if not provided:
    - stnk/bpkb/vehicle_insurance → vehicles
    - sim/passport/visa/ktp → identity
    - health_insurance/life_insurance → insurance
    - tax/investment → financials
  - Set default `reminder_offsets = [30, 7, 1]` if not provided.
  - Insert document row.
  - For each offset in `reminder_offsets`: if `expiry_date - offset` is in the future, create a `notifications` row with `scheduled_at = expiry_date - offset days`, idempotency key `document:{docID}:offset:{offset}:channel:whatsapp`.
  - Increment usage counter. Write audit log.

- `UploadDocumentFile(userID, documentID string, file io.Reader, mimeType string, sizeBytes int) (*Document, error)`:
  - Ownership check.
  - Validate MIME type: only `application/pdf`, `image/jpeg`, `image/png`.
  - Validate size ≤ 10MB (10485760 bytes).
  - Generate storage key: `documents/{userID}/{documentID}/{uuid}.{ext}`.
  - Upload via `FileStore.Upload`. Get URL.
  - Update document: `file_url`, `file_size_bytes`, `file_mime_type`, `storage_type=digital`.
  - Write audit log.

- `GetDocumentFileURL(userID, documentID string) (string, error)`: Ownership check. Generate signed URL with 15-minute expiry via `FileStore.SignedURL`.

- `DeleteDocumentFile(userID, documentID string) error`: Ownership check. Delete from storage. Clear file fields on document. Write audit log.

- `GetDocuments(userID string, filter DocumentFilter, page, perPage int) (*PaginatedResult, error)`.

- `GetDocumentByID(userID, documentID string) (*Document, error)`: Ownership check.

- `UpdateDocument(userID, documentID string, req UpdateDocumentRequest) (*Document, error)`: If `expiry_date` changed, cancel old expiry notifications and recreate.

- `DeleteDocument(userID, documentID string) error`: Soft delete. Cancel notifications. Delete file from storage if present. Decrement usage counter. Write audit log.

- `GetStorageUsage(userID string) (*StorageUsage, error)`: Sum `file_size_bytes` where `user_id = userID AND file_url IS NOT NULL`. Return `{UsedBytes, UsedMB, UsedGB, FileCount}`.

- `GetVaultHealth(userID string) (*VaultHealth, error)`: Count documents per `vault_category`. Return completeness score: if user has at least one doc in vehicles, identity, insurance = 100%; partial = proportional.

---

### TASK-009 — Subscription application service
**Repo:** `woles-backend`

In `internal/application/subscription/service.go`:

- `CreateSubscription(userID string, req CreateSubscriptionRequest) (*Subscription, error)`: Check plan limit. Validate `billing_cycle`. Sanitize `name` (max 200 chars). If `next_billing_at` not provided, calculate from `billing_cycle`. Create notification for billing date. Increment usage counter. Write audit log.

- `GetSubscriptions(userID string, filter SubscriptionFilter, page, perPage int) (*PaginatedResult, error)`.

- `GetSubscriptionByID(userID, subscriptionID string) (*Subscription, error)`: Ownership check.

- `UpdateSubscription(userID, subscriptionID string, req UpdateSubscriptionRequest) (*Subscription, error)`: If `next_billing_at` changed, reschedule notification.

- `DeleteSubscription(userID, subscriptionID string) error`: Soft delete. Cancel notifications. Decrement counter. Write audit log.

- `ArchiveSubscription(userID, subscriptionID string) error`: Set `status=archived`. Cancel notifications.

- `GetMonthlyCostSummary(userID string) (*MonthlyCostSummary, error)`: Sum amounts by currency for active subscriptions. Return `[{Currency, TotalAmount, SubscriptionCount}]`.

---

### TASK-010 — Goal application service
**Repo:** `woles-backend`

In `internal/application/goal/service.go`:

- `CreateGoal(userID string, req CreateGoalRequest) (*Goal, error)`: Premium plan required (check `users.plan != 'free'`). Validate `target_amount > 0`. Validate `icon` against allowed values. Sanitize `title`. Insert goal. Write audit log.

- `GetGoals(userID string, filter GoalFilter, page, perPage int) (*PaginatedResult, error)`.

- `GetGoalByID(userID, goalID string) (*Goal, error)`: Ownership check.

- `UpdateGoal(userID, goalID string, req UpdateGoalRequest) (*Goal, error)`.

- `DeleteGoal(userID, goalID string) error`: Soft delete. Write audit log.

- `UpdateProgress(userID, goalID string, newAmount float64) (*Goal, error)`: Ownership check. Update `current_amount`. If `current_amount >= target_amount`, set `status=completed`. Write audit log.

- `GetGoalHistory(userID string, page, perPage int) (*PaginatedResult, error)`: Return completed and archived goals.

- `GetActiveGoalWithTip(userID string) (*GoalWithTip, error)`: Fetch first active goal. Generate tip text based on `(target_amount - current_amount) / monthly_target` → estimated months remaining. If on track, return positive tip. If behind, return motivational tip.

---

### TASK-011 — Timeline application service
**Repo:** `woles-backend`

In `internal/application/timeline/service.go`:

- `GetTimeline(userID string, from, to time.Time, page, perPage int) (*PaginatedResult[TimelineItem], error)`:
  - Query upcoming `reminder_occurrences` where `scheduled_at BETWEEN from AND to AND user_id = userID AND status = 'scheduled'`. Join `reminders` for title and category.
  - Query `documents` where `expiry_date BETWEEN from AND to AND user_id = userID AND status = 'active'`. For each active reminder offset, compute alert date.
  - Query `subscriptions` where `next_billing_at BETWEEN from AND to AND user_id = userID AND status = 'active'`.
  - Query `goals` where `target_date BETWEEN from AND to AND user_id = userID AND status = 'active'`.
  - Normalize all into `TimelineItem{ID, Type, Title, DueAt, Status, EntityID}`.
  - Sort by `DueAt ASC`.
  - Apply offset pagination.
  - Return paginated result.

- `GetTimelineByRange(userID, rangeStr string, page, perPage int) (*PaginatedResult[TimelineItem], error)`: Parse range strings: `7d`, `30d`, `90d`, `this_month`, `next_month`. Call `GetTimeline` with computed from/to.

---

### TASK-012 — Notification application service
**Repo:** `woles-backend`

In `internal/application/notification/service.go`:

- `GetNotifications(userID string, filter NotificationFilter, page, perPage int) (*PaginatedResult, error)`: Filter by `entity_type` (maps to category), date range, status.

- `GetStats(userID string) (*NotificationStats, error)`: Return `{TotalSent, TotalFailed, DeliveryRatePct, TopCategory}`.

- `ExportNotifications(userID, format, rangeStr string) ([]byte, error)`: Generate PDF or CSV. For PDF use a Go PDF library. Include user name, date range, and notification table.

---

### TASK-013 — Financial overview application service
**Repo:** `woles-backend`

In `internal/application/billing/finance_service.go` (extends billing domain for financial overview):

- `GetFinancialSummary(userID, period string) (*FinancialSummary, error)`:
  - Parse period string (`monthly`, `YYYY-MM`).
  - `total_expenses`: sum `subscriptions.amount` where active + any manual transaction entries in the period.
  - `income`, `savings`: from manual entries only (V2: bank sync). Return 0 for MVP if no manual entries.
  - `change_vs_last_period_pct`: compare with previous period total.

- `GetSpendingByCategory(userID, period string) (*SpendingBreakdown, error)`:
  - Group subscriptions by `category`. Map to display categories: entertainment → Household, productivity → Utilities, bill → Utilities, vehicle subscriptions → Transport, other → Others.
  - Calculate percentage for each.

- `GetSpendingTrend(userID, period string) (*SpendingTrend, error)`: For `weekly`: return spending totals per week of the current month. Derive from subscription `next_billing_at` falling in each week.

- `GetUpcomingBills(userID string, page, perPage int) (*PaginatedResult, error)`:
  - Query subscriptions with `next_billing_at <= NOW() + 30 days` and `status = 'active'`.
  - For each: compute STATUS: `URGENT` if `next_billing_at <= NOW() + 3 days`, `PENDING` if `<= 7 days`, `SCHEDULED` otherwise.
  - Sort by `next_billing_at ASC`.

- `ExportFinances(userID, format, period string) ([]byte, error)`: CSV export of subscriptions and summary for the period.

---

### TASK-014 — Family application service
**Repo:** `woles-backend`

In `internal/application/family/service.go`:

- Gate all methods: check `users.plan = 'advanced'`. Return 403 with upgrade prompt otherwise.
- `CreateMember(ownerUserID string, req CreateMemberRequest) (*FamilyMember, error)`: Validate `role`. Sanitize `name`, `relation_label`. Max 10 family members per user.
- `GetMembers(ownerUserID string) ([]*FamilyMember, error)`: Include `active_reminder_count` per member (join reminders).
- `GetMemberByID(ownerUserID, memberID string) (*FamilyMember, error)`: Ownership check.
- `UpdateMember(ownerUserID, memberID string, req UpdateMemberRequest) (*FamilyMember, error)`.
- `DeleteMember(ownerUserID, memberID string) error`: Soft delete. Reassign member's reminders/documents to owner. Write audit log.
- `GetSharedReminders(ownerUserID string, page, perPage int) (*PaginatedResult, error)`: Return reminders and document expiry alerts across all family members. Include `owner_name` and `owner_avatar` fields.

---

### TASK-015 — AI Chat application service
**Repo:** `woles-backend`

In `internal/application/chat/service.go`:

- `SendMessage(userID, content string) (*ChatMessage, error)`:
  - Check chat quota: `ChatUsageRepository.IsWithinLimit(userID)`. Free plan limit = 10/month. If exceeded, return 403 with `quota_exceeded` error and upgrade CTA.
  - Sanitize content (max 4000 chars, strip HTML).
  - Store user message with `role=user`.
  - Call `IntentExtractor.Extract(content, "id")`.
  - Store assistant response with `role=assistant`, `detected_intent=result.Payload`.
  - Increment `chat_usage.messages_used`.
  - Return both messages.

- `GetMessages(userID string, page, perPage int) (*PaginatedResult[ChatMessage], error)`: Return messages sorted by `created_at ASC`.

- `DeleteAllMessages(userID string) error`: Hard delete all chat messages for user.

- `GetUsage(userID string) (*ChatUsageResponse, error)`: Return `{MessagesUsed, Quota, Remaining, PlanName}`.

- `GetDetectedIntents(userID string) ([]*DetectedIntent, error)`: Aggregate unique `detected_intent.intent` values from recent messages. Group by intent type. Return top 5.

---

## Backend — Inbound Ports and HTTP Adapter

### TASK-016 — Generate Fiber HTTP adapter with Prabogo CLI
**Repo:** `woles-backend`

Run `prabogo generate http` for all route groups. Each handler file lives in `internal/adapter/inbound/http_fiber/`.

Create the following handler files:

**auth_handler.go** — routes under `/api/v1/auth/`:
- All endpoints from Section 8.2 Authentication Endpoints.
- Apply CSRF middleware to all POST/DELETE methods.
- Apply rate limiting middleware (Section 8.5 limits).

**reminder_handler.go** — routes under `/api/v1/reminders/`.

**document_handler.go** — routes under `/api/v1/documents/`.
- For file upload endpoint: use `multipart/form-data` parser. Enforce 10MB limit via Fiber body limit config.

**subscription_handler.go** — routes under `/api/v1/subscriptions/`.

**goal_handler.go** — routes under `/api/v1/goals/`.

**finance_handler.go** — routes under `/api/v1/finances/`.

**timeline_handler.go** — routes under `/api/v1/timeline/`.

**notification_handler.go** — routes under `/api/v1/notifications/`.

**family_handler.go** — routes under `/api/v1/family/`.

**chat_handler.go** — routes under `/api/v1/chat/`.

**account_handler.go** — routes under `/api/v1/account/`.

**billing_handler.go** — routes under `/api/v1/billing/`.

**webhook_handler.go** — route `POST /webhooks/whatsapp/{provider}`.
- No CORS headers on this route (server-to-server).
- No CSRF check.
- Verify HMAC-SHA256 signature before processing.
- Return HTTP 200 immediately after signature verification.
- Publish `whatsapp.message_received` to RabbitMQ for async processing.

---

### TASK-017 — Implement security middleware
**Repo:** `woles-backend`

In `internal/adapter/inbound/http_fiber/middleware/`:

**jwt_middleware.go**:
- Parse `Authorization: Bearer <token>` header.
- Verify RS256 signature with public key loaded from `JWT_PUBLIC_KEY_PATH`.
- Validate expiry, issuer, and audience claims.
- On valid token: set `userID`, `plan`, `tz` in Fiber context locals.
- On invalid/expired token: return HTTP 401 `{"error":"unauthorized","message":"Token is expired or invalid"}`.

**rate_limit_middleware.go**:
- Accept `limit int` and `window time.Duration` parameters.
- Use Redis sliding window: key `ratelimit:{identifier}:{endpoint_hash}`.
- For unauthenticated routes: identifier = IP address.
- For authenticated routes: identifier = `userID`.
- Set `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, `Retry-After` response headers on every response.
- Return HTTP 429 on exceeded limit.

**csrf_middleware.go**:
- On GET requests: issue a new CSRF token (32-byte random, base64url encoded) in a `csrf_token` cookie: `SameSite=Strict; Secure; HttpOnly=false; Path=/; Max-Age=86400`.
- On POST/PATCH/DELETE requests: read `X-CSRF-Token` header and `csrf_token` cookie. Compare with constant-time equality. Return HTTP 403 on mismatch or missing header.
- Skip for `/webhooks/` paths.

**cors_middleware.go**:
- Allowed origins: read from `CORS_ALLOWED_ORIGINS` env var (comma-separated). Default to `https://woles.id,https://www.woles.id`.
- On preflight OPTIONS: respond HTTP 204 with CORS headers.
- `Access-Control-Allow-Credentials: true`.
- Never set wildcard origin.

**security_headers_middleware.go**:
- Set on every response:
  - `Content-Security-Policy: default-src 'self'; script-src 'self'; object-src 'none'; frame-ancestors 'none'`
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: DENY`
  - `Referrer-Policy: strict-origin-when-cross-origin`
  - `Permissions-Policy: geolocation=(), microphone=(), camera=()`
  - `Strict-Transport-Security: max-age=63072000; includeSubDomains; preload`

**plan_gate_middleware.go**:
- Accept `requiredPlan string` parameter (`premium` or `advanced`).
- Compare `ctx.Locals("plan")` with required plan.
- Return HTTP 403 with `{"error":"plan_required","required_plan":"premium","upgrade_url":"/billing/checkout"}` if insufficient.

---

### TASK-018 — Implement PostgreSQL adapter
**Repo:** `woles-backend`

Run `prabogo generate postgres` for each repository interface. Implement all repository methods defined in TASK-005.

In `internal/adapter/outbound/postgres/`:

- Use `pgx/v5` as the driver. Use connection pool with `pgxpool`.
- All queries use parameterized statements (never string concatenation with user input).
- `claim_for_sending.go`: implement `FOR UPDATE SKIP LOCKED` pattern for `NotificationRepository.ClaimDue` and `ReminderOccurrenceRepository.ClaimForSending`.
- Implement soft-delete pattern: never hard-delete user data rows (set `deleted_at` or `status=archived/deleted`).
- All list queries accept `PaginationParams{Page, PerPage, Sort, Order}`. Validate `Sort` against an allowlist of column names before using in ORDER BY. Never interpolate sort column from user input directly.
- `GetVaultHealth`: return count per vault_category and calculate completeness score.

---

### TASK-019 — Implement Redis adapter
**Repo:** `woles-backend`

In `internal/adapter/outbound/redis/`:

- Use `go-redis/v9`.
- Implement all cache port interfaces from TASK-005.
- `RateLimiter`: use `ZADD` + `ZREMRANGEBYSCORE` + `ZCARD` in a Lua script for atomic sliding window. Key format: `rl:{identifier}:{endpoint}`. TTL = window duration.
- `IdempotencyStore`: `SET key 1 NX EX {ttl_seconds}`. Return `ErrAlreadyProcessed` if key already exists.
- `OTPStore`: `SET otp:{phone} {hashedOTP} EX 300`. `GET`. `DEL`.

---

### TASK-020 — Implement RabbitMQ adapter and workers
**Repo:** `woles-backend`

In `internal/adapter/outbound/rabbitmq/` and `internal/adapter/inbound/rabbitmq/`:

**Publisher** (outbound):
- Declare exchanges: `woles.events` (topic, durable).
- Routing keys: `reminder.created`, `reminder.updated`, `notification.scheduled`, `notification.send_requested`, `notification.sent`, `notification.failed`, `whatsapp.message_received`, `intent.extraction_requested`.
- All messages are persistent (delivery mode 2).
- Implement `Publish(routingKey string, payload interface{}) error`.

**Consumer** (inbound):
- Declare queues binding to `woles.events` exchange.
- `intent_worker_queue` — binds `whatsapp.message_received`.
- `notification_send_queue` — binds `notification.send_requested`.
- `notification_result_queue` — binds `notification.sent`, `notification.failed`.
- Declare dead-letter exchange `woles.dlx` and DLQ `woles.dlq` for failed messages after max retries.

**Intent Worker** (`cmd/intent_worker/main.go`):
- Consume `whatsapp.message_received`.
- Parse `InboundMessage` from payload.
- Call `IntentExtractor.Extract`.
- Based on intent: call the appropriate application service (create reminder, create subscription, create goal, create document, query timeline).
- Update `inbound_messages.processing_status`.
- If confidence < 0.7: send clarification message via WhatsApp.
- Handle errors: ack message on permanent failure. Nack with requeue=false on transient failure (moves to DLQ after max retries set by RabbitMQ policy).

**Notification Worker** (`cmd/notification_worker/main.go`):
- Consume `notification.send_requested`.
- Lookup notification by ID. Verify `status=sending` (double-check to avoid duplicates).
- Call `WhatsAppSender.SendMessage`.
- On success: update `status=sent`, `sent_at=now`, `provider_message_id`. Publish `notification.sent`.
- On failure: increment `retry_count`. If `retry_count < 3`: update status back to `scheduled`, let scheduler retry. If `retry_count >= 3`: update `status=failed`, `failure_reason`. Publish `notification.failed`.

**Scheduler** (`cmd/scheduler/main.go`):
- Run every 60 seconds using `time.Ticker`.
- Call `NotificationRepository.ClaimDue(batchSize=50)` using `FOR UPDATE SKIP LOCKED`.
- For each claimed notification: update `status=sending`. Publish `notification.send_requested` to RabbitMQ.
- Log: count of claimed notifications, any errors.

---

### TASK-021 — Implement MinIO storage adapter
**Repo:** `woles-backend`

In `internal/adapter/outbound/storage_provider/minio.go`:

- Use `minio-go/v7`.
- Implement `FileStore` interface.
- `Upload`: call `client.PutObject`. Set `ContentType` from mimeType param. Return the object key (not a public URL).
- `Delete`: call `client.RemoveObject`.
- `SignedURL`: call `client.PresignedGetObject` with `ttl` duration. Return the pre-signed URL.
- Bucket policy: bucket is private. All file access requires signed URLs.

---

### TASK-022 — Write OpenAPI 3.1 spec and Postman collection
**Repo:** `woles-backend`

Create `docs/openapi.yaml`:

- Document all endpoints from Section 8.2 and 8.3.
- Every endpoint must have: `operationId`, `summary`, `description`, `tags`, `security` (bearerAuth except auth endpoints), request body (`$ref`), and responses for 200, 400, 401, 403, 404, 422, 429, 500.
- Define reusable `$ref` schemas for every request and response body under `components/schemas/`.
- Key schemas: `PaginationMeta`, `ErrorResponse`, `User`, `Reminder`, `Document`, `Subscription`, `Goal`, `Notification`, `FamilyMember`, `ChatMessage`, `TimelineItem`, `CreateReminderRequest`, `UpdateReminderRequest`, etc.
- Document multipart file upload for `POST /documents/{id}/file`: use `content: multipart/form-data` with `file` property of type `string, format: binary`.
- Document webhook endpoint with required `X-Webhook-Signature` header.

Generate Postman collection:
- Use `openapi-to-postman` or equivalent tool to generate `docs/postman/woles.postman_collection.json` from the OpenAPI spec.
- Create `docs/postman/woles.postman_environment.json` with variables: `base_url`, `access_token`, `refresh_token`, `user_id`.
- Add a Postman pre-request script on `POST /auth/refresh` that reads the `access_token` variable, checks if it is expired (decode JWT expiry without verifying signature), and calls the refresh endpoint automatically if expired.
- Commit both files to the repository.

---

### TASK-023 — Write backend tests
**Repo:** `woles-backend`

In `tests/`:

**Unit tests** (`tests/unit/`):
- `reminder_recurrence_test.go`: test `CalculateNextRunAt` for all recurrence types including edge cases (month-end dates, leap years, DST boundaries in Asia/Jakarta).
- `document_expiry_test.go`: test `DaysUntilExpiry`, `ExpiryRisk`, and notification offset calculation.
- `subscription_billing_test.go`: test next billing date calculation for monthly, yearly, custom cycles.
- `timeline_aggregation_test.go`: test item normalization and sorting with mixed entity types.
- `plan_limit_test.go`: test `IsWithinLimit` returns correct results for all plans.
- `notification_idempotency_test.go`: test `BuildIdempotencyKey` uniqueness.
- `argon2id_test.go`: test hash and verify with known vectors.
- `account_lockout_test.go`: test lockout after 5 failures, unlock after 15 minutes.
- `jwt_test.go`: test access token generation, expired token rejection, tampered payload rejection, `none` algorithm rejection.
- `refresh_token_rotation_test.go`: test old token rejected after rotation, reuse triggers family revocation.
- `csrf_test.go`: test missing header returns 403, mismatched header returns 403.
- `ownership_test.go`: test user A cannot access user B's reminder (expects 404).
- `rate_limit_test.go`: test 11th login attempt within 15 minutes returns 429.
- `input_validation_test.go`: test 201-char title returns 422, SQL injection payload is safely parameterized.
- `webhook_signature_test.go`: test invalid HMAC returns 401.

**Integration tests** (`tests/integration/`): require a real Postgres and Redis (use Docker in CI).
- `postgres_reminder_test.go`: create reminder, fetch, update, delete, verify soft delete.
- `postgres_notification_claim_test.go`: concurrent claim with FOR UPDATE SKIP LOCKED does not double-claim.
- `redis_rate_limiter_test.go`: 10 allowed, 11th blocked within window.
- `redis_otp_test.go`: set, get, expired after TTL.
- `refresh_token_store_test.go`: issue, rotate, revoke, detect reuse.
- `audit_log_test.go`: verify audit rows written on login, logout, password change.

---

### TASK-046 — Implement HTTP handler logic for all REST API endpoints
**Repo:** `woles-backend`

All handler files in `internal/adapter/inbound/http_fiber/` currently return `501 Not Implemented`. This task wires every handler to its application service and implements request parsing, validation, service call, error mapping, and response marshaling.

#### Step 1 — Dependency injection refactor

Refactor `router.go` and all handler files to use constructor-based dependency injection:

- Define a `Services` struct in `internal/adapter/inbound/http_fiber/services.go`:
  ```go
  type Services struct {
      Identity     *identity.Service
      Reminder     *reminder.Service
      Document     *document.Service
      Subscription *subscription.Service
      Goal         *goal.Service
      Timeline     *timeline.Service
      Notification *notification.Service
      Family       *family.Service
      Chat         *chat.Service
      Finance      *billing.FinancialService
  }
  ```
- Change `RegisterRoutes(app *fiber.App)` → `RegisterRoutes(app *fiber.App, svc *Services)`.
- Thread `svc` into each `Register*Routes(router, svc)` call.
- Convert each handler group to a struct (e.g., `authHandler{svc *identity.Service}`) with methods instead of bare functions. Wire via the `Register*Routes` constructor.
- Update `cmd/main.go` to construct all services and pass a populated `Services` to `RegisterRoutes`.

#### Step 2 — Shared helpers

Create `internal/adapter/inbound/http_fiber/response.go` with:
- `userIDFromCtx(c *fiber.Ctx) string` — read `userID` from Fiber locals (set by JWT middleware). Return empty string and send 401 if missing.
- `sendError(c *fiber.Ctx, status int, code, message string) error` — unified JSON error response `{"error": code, "message": message}`.
- `mapServiceError(c *fiber.Ctx, err error) error` — map common service errors to HTTP status codes:
  - `ErrNotFound` → 404
  - `ErrForbidden` / ownership violation → 403
  - `ErrInvalidCredentials` / `ErrAccountLocked` → 400
  - `ErrTokenReused` / `ErrTokenInvalid` → 401
  - `ErrAlreadyProcessed` → 409
  - `quota_exceeded` errors → 403
  - default → 500

#### Step 3 — JWT middleware wiring in router

In `RegisterRoutes`, apply `middleware.JWTMiddleware()` to all protected route groups:
- `/api/v1/auth/logout`, `/api/v1/auth/password/change`, `/api/v1/auth/me`, `/api/v1/auth/2fa/*`, `/api/v1/auth/sessions*` — require JWT.
- All `/api/v1/reminders`, `/api/v1/documents`, `/api/v1/subscriptions`, `/api/v1/goals`, `/api/v1/finances`, `/api/v1/timeline`, `/api/v1/notifications`, `/api/v1/family`, `/api/v1/chat`, `/api/v1/account`, `/api/v1/billing` — require JWT.
- `/api/v1/auth/register`, `/api/v1/auth/login`, `/api/v1/auth/refresh`, `/api/v1/auth/otp/*`, `/api/v1/auth/password/reset/*` — public, no JWT.
- `/webhooks/*` — no JWT, no CORS.

#### Step 4 — Auth handlers (`auth_handler.go`)

Implement each handler. The refresh token is stored in an `HttpOnly; Secure; SameSite=Strict; Path=/; Max-Age=2592000` cookie named `refresh_token`. Clear it on logout.

- `handleRegister`: parse `{email, password, name, timezone}`. Call `identity.RegisterWithEmail`. On success: set refresh cookie, return `201 {"user": ..., "tokens": {"access_token": ...}}`.
- `handleLogin`: parse `{email, password}`. Read `X-Forwarded-For`/`RemoteAddr` for IP and `User-Agent` header. Call `identity.LoginWithEmail`. Set refresh cookie. Return `200 {"user": ..., "tokens": {"access_token": ...}}`.
- `handleRefreshToken`: read `refresh_token` cookie. Read IP + UA. Call `identity.RefreshToken`. Set new refresh cookie. Return `200 {"access_token": ...}`.
- `handleLogout`: read `sub` (userID) and `jti` (token family) from JWT locals. Read refresh token cookie. Call `identity.Logout`. Clear cookie. Return `200 {"message": "Logged out successfully"}`.
- `handleRequestOTP`: parse `{phone}`. Call `identity.RequestOTP`. Return `200 {"message": "OTP sent to WhatsApp"}`.
- `handleVerifyOTP`: parse `{phone, otp}`. Call `identity.VerifyOTP`. Set refresh cookie. Return `200 {"user": ..., "tokens": {"access_token": ...}}`.
- `handlePasswordResetRequest`: parse `{email}`. Stub — return `200 {"message": "If the email exists, a reset link has been sent"}` (constant-time response, do not leak existence).
- `handlePasswordResetConfirm`: parse `{token, new_password}`. Stub — return `501`.
- `handleChangePassword`: parse `{old_password, new_password}`. Call `identity.ChangePassword`. Return `200 {"message": "Password changed"}`.
- `handleGetMe`: read userID from JWT locals. Lookup user from DB via identity repo. Return `200 {"user": ...}`.
- `handleEnable2FA`: call `identity.Enable2FA`. Return `200 {"totp_uri": ..., "secret": ...}`.
- `handleVerify2FA`: parse `{totp_code}`. Call `identity.Verify2FA`. Return `200 {"message": "2FA enabled"}`.
- `handleDisable2FA`: parse `{password}`. Call `identity.Disable2FA`. Return `200 {"message": "2FA disabled"}`.
- `handleGetSessions`: call `identity.GetActiveSessions`. Return `200 {"sessions": [...]}`.
- `handleRevokeSession`: read `:id`. Call `identity.RevokeSession`. Return `200 {"message": "Session revoked"}`.
- `handleRevokeAllSessions`: call `identity.LogoutAllSessions`. Clear cookie. Return `200 {"message": "All sessions revoked"}`.

#### Step 5 — Reminder handlers (`reminder_handler.go`)

- `handleCreateReminder`: parse body into `reminder.CreateReminderRequest` (fields: `title`, `category`, `recurrence_type`, `recurrence_rule`, `next_run_at`, `timezone`). Call `reminder.CreateReminder`. Return `201 {"reminder": ...}`.
- `handleListReminders`: read query params `page` (default 1), `per_page` (default 20, max 100), `status`, `category`, `sort` (allowlist: `next_run_at`, `created_at`), `order` (`asc`/`desc`). Build `database.ReminderFilter`. Call `reminder.GetReminders`. Return `200 {"reminders": [...], "meta": {...pagination...}}`.
- `handleGetReminder`: read `:id`. Call `reminder.GetReminderByID`. Return `200 {"reminder": ...}`.
- `handleUpdateReminder`: read `:id`, parse partial body into `reminder.UpdateReminderRequest`. Call `reminder.UpdateReminder`. Return `200 {"reminder": ...}`.
- `handleDeleteReminder`: read `:id`. Call `reminder.DeleteReminder`. Return `200 {"message": "Reminder deleted"}`.
- `handlePauseReminder`: read `:id`. Call `reminder.PauseReminder`. Return `200 {"message": "Reminder paused"}`.
- `handleResumeReminder`: read `:id`. Call `reminder.ResumeReminder`. Return `200 {"message": "Reminder resumed"}`.
- `handleCompleteOccurrence`: read `:id`. Call `reminder.CompleteOccurrence`. Return `200 {"message": "Occurrence completed"}`.

#### Step 6 — Document handlers (`document_handler.go`)

- `handleCreateDocument`: parse body into `document.CreateDocumentRequest`. Call `document.CreateDocument`. Return `201 {"document": ...}`.
- `handleListDocuments`: query params `page`, `per_page`, `status`, `vault_category`, `sort` (allowlist: `expiry_date`, `created_at`), `order`. Call `document.GetDocuments`. Return `200 {"documents": [...], "meta": {...}}`.
- `handleGetDocument`: read `:id`. Call `document.GetDocumentByID`. Return `200 {"document": ...}`.
- `handleUpdateDocument`: read `:id`, parse body. Call `document.UpdateDocument`. Return `200 {"document": ...}`.
- `handleDeleteDocument`: read `:id`. Call `document.DeleteDocument`. Return `200 {"message": "Document deleted"}`.
- `handleUploadDocumentFile`: read `:id`. Parse `multipart/form-data` with `c.FormFile("file")`. Validate MIME type (`application/pdf`, `image/jpeg`, `image/png`) and size (max 10 MB). Call `document.UploadDocumentFile`. Return `200 {"document": ...}`.
- `handleGetDocumentFileURL`: read `:id`. Call `document.GetDocumentFileURL`. Return `200 {"url": ..., "expires_in": 900}`.
- `handleDeleteDocumentFile`: read `:id`. Call `document.DeleteDocumentFile`. Return `200 {"message": "File deleted"}`.
- `handleGetStorageUsage`: call `document.GetStorageUsage`. Return `200 {"storage": ...}`.
- `handleGetVaultHealth`: call `document.GetVaultHealth`. Return `200 {"vault": ...}`.

#### Step 7 — Subscription handlers (`subscription_handler.go`)

- `handleCreateSubscription`: parse body into `subscription.CreateSubscriptionRequest`. Call `subscription.CreateSubscription`. Return `201 {"subscription": ...}`.
- `handleListSubscriptions`: query params `page`, `per_page`, `status`, `sort` (allowlist: `next_billing_at`, `name`, `created_at`), `order`. Call `subscription.GetSubscriptions`. Return `200 {"subscriptions": [...], "meta": {...}}`.
- `handleGetSubscription`: `:id` → `subscription.GetSubscriptionByID`. Return `200 {"subscription": ...}`.
- `handleUpdateSubscription`: `:id` + body → `subscription.UpdateSubscription`. Return `200 {"subscription": ...}`.
- `handleDeleteSubscription`: `:id` → `subscription.DeleteSubscription`. Return `200 {"message": "Subscription deleted"}`.
- `handleArchiveSubscription`: `:id` → `subscription.ArchiveSubscription`. Return `200 {"message": "Subscription archived"}`.

#### Step 8 — Goal handlers (`goal_handler.go`)

- `handleCreateGoal`: parse body into `goal.CreateGoalRequest`. Call `goal.CreateGoal`. Return `201 {"goal": ...}`.
- `handleListGoals`: query params `page`, `per_page`, `status`. Call `goal.GetGoals`. Return `200 {"goals": [...], "meta": {...}}`.
- `handleGetGoal`: `:id` → `goal.GetGoalByID`. Return `200 {"goal": ...}`.
- `handleUpdateGoal`: `:id` + body → `goal.UpdateGoal`. Return `200 {"goal": ...}`.
- `handleDeleteGoal`: `:id` → `goal.DeleteGoal`. Return `200 {"message": "Goal deleted"}`.
- `handleUpdateGoalProgress`: `:id`, parse `{current_amount}`. Call `goal.UpdateProgress`. Return `200 {"goal": ...}`.
- `handleGetGoalHistory`: query params `page`, `per_page`. Call `goal.GetGoalHistory`. Return `200 {"goals": [...], "meta": {...}}`.

#### Step 9 — Finance handlers (`finance_handler.go`)

- `handleGetFinancialSummary`: query param `period` (default `monthly`). Call `billing.GetFinancialSummary`. Return `200 {"summary": ...}`.
- `handleGetSpending`: query param `period`. Call `billing.GetSpendingByCategory`. Return `200 {"categories": [...]}`.
- `handleGetSpendingTrend`: query param `period` (default `weekly`). Call `billing.GetSpendingTrend`. Return `200 {"weeks": [...]}`.
- `handleGetUpcomingBills`: query params `page`, `per_page`. Call `billing.GetUpcomingBills`. Return `200 {"bills": [...], "meta": {...}}`.
- `handleExportFinances`: query params `format` (`csv`), `period`. Call `billing.ExportFinances`. Set `Content-Type: text/csv` and `Content-Disposition: attachment; filename="finances.csv"`. Write raw bytes.

#### Step 10 — Timeline handlers (`timeline_handler.go`)

- `handleGetTimeline`: query params `range` (values: `7d`, `30d`, `90d`, `this_month`, `next_month`; default `30d`), `page`, `per_page`. Call `timeline.GetTimelineByRange`. Return `200 {"items": [...], "meta": {...}}`.

#### Step 11 — Notification handlers (`notification_handler.go`)

- `handleListNotifications`: query params `page`, `per_page`, `entity_type`, `status`, `from`, `to`. Build `database.NotificationFilter`. Call `notification.GetNotifications`. Return `200 {"notifications": [...], "meta": {...}}`.
- `handleGetNotificationStats`: call `notification.GetStats`. Return `200 {"stats": ...}`.
- `handleExportNotifications`: query params `format` (`csv`, `pdf`), `range`. Call `notification.ExportNotifications`. Set appropriate `Content-Type` and `Content-Disposition`. Write raw bytes.

#### Step 12 — Family handlers (`family_handler.go`)

- `handleCreateFamilyMember`: parse body into `family.CreateMemberRequest`. Call `family.CreateMember`. Return `201 {"member": ...}`.
- `handleListFamilyMembers`: call `family.GetMembers`. Return `200 {"members": [...]}`.
- `handleGetFamilyMember`: `:id` → `family.GetMemberByID`. Return `200 {"member": ...}`.
- `handleUpdateFamilyMember`: `:id` + body → `family.UpdateMember`. Return `200 {"member": ...}`.
- `handleDeleteFamilyMember`: `:id` → `family.DeleteMember`. Return `200 {"message": "Member deleted"}`.
- `handleGetSharedReminders`: query params `page`, `per_page`. Call `family.GetSharedReminders`. Return `200 {"items": [...], "meta": {...}}`.

#### Step 13 — Chat handlers (`chat_handler.go`)

- `handleSendMessage`: parse `{content}`. Call `chat.SendMessage`. Return `200 {"messages": [user_message, assistant_message]}`.
- `handleListMessages`: query params `page`, `per_page`. Call `chat.GetMessages`. Return `200 {"messages": [...], "meta": {...}}`.
- `handleDeleteAllMessages`: call `chat.DeleteAllMessages`. Return `200 {"message": "Chat history cleared"}`.
- `handleGetChatUsage`: call `chat.GetUsage`. Return `200 {"usage": ...}`.
- `handleGetDetectedIntents`: call `chat.GetDetectedIntents`. Return `200 {"intents": [...]}`.

#### Step 14 — Account handlers (`account_handler.go`)

The account service reuses the identity service for user reads/writes. Pass `*identity.Service` (and `*storage.FileStore` for avatar).

- `handleGetProfile`: read userID from JWT. Look up user via identity service. Return `200 {"user": ...}`.
- `handleUpdateProfile`: parse `{name, timezone}`. Update user fields via identity service (add `UpdateProfile` method if not present). Return `200 {"user": ...}`.
- `handleUploadAvatar`: parse multipart `avatar` file (JPEG/PNG, max 2 MB). Call `storage.Upload`. Update `users.avatar_url`. Return `200 {"avatar_url": ...}`.
- `handleExportAccountData`: generate a JSON export of the user's data (profile, reminders count, documents count, subscriptions count). Return `200` with `Content-Disposition: attachment; filename="account-export.json"`.
- `handleDeleteAccount`: call identity soft-delete (set `account_status=deleted`). Revoke all sessions. Clear cookie. Return `200 {"message": "Account deleted"}`.

#### Step 15 — Billing handlers (`billing_handler.go`)

- `handleGetCurrentPlan`: read `plan` from JWT locals. Query `plans_and_usage_limits` table for plan details and current usage. Return `200 {"plan": "free", "limits": {...}, "usage": {...}}`.
- `handleCreateCheckout`: parse `{plan}`. Call payment provider service to create a checkout session/link. Return `200 {"checkout_url": ...}`. For MVP if payment provider not yet wired, return `501`.
- `handlePaymentWebhook`: verify HMAC-SHA256 signature (`X-Payment-Signature` header vs `PAYMENT_WEBHOOK_SECRET`). Parse event. On successful payment: update `users.plan`. Return `200 {"received": true}`.

#### Step 16 — Webhook handler (`webhook_handler.go`)

- `handleWhatsAppWebhook`: already partially defined. Complete implementation:
  1. Read `X-Webhook-Signature` header. Verify HMAC-SHA256 against raw body using `WHATSAPP_WEBHOOK_SECRET`. Return `401` on mismatch.
  2. Parse body into `WhatsAppWebhookPayload`.
  3. Publish `whatsapp.message_received` event to RabbitMQ with the raw payload.
  4. Return `200 {"status": "received"}` immediately (async processing happens in the intent worker).

#### Acceptance criteria

- `go build ./...` passes with no errors.
- All 114 registered handlers return real responses (no more `501 Not Implemented`), except `handlePasswordResetConfirm` and `handleCreateCheckout` which may remain `501` until their respective provider integrations are complete.
- `POST /api/v1/auth/register` + `POST /api/v1/auth/login` return `200`/`201` with valid JWT.
- `GET /api/v1/auth/me` with a valid Bearer token returns the authenticated user.
- All protected endpoints return `401` when called without a valid Bearer token.
- All ownership-protected endpoints (reminders, documents, etc.) return `404` when accessed with another user's ID.

---

## Frontend — Foundation

### TASK-024 — Set up API client layer
**Repo:** `woles-frontend`

In `lib/api/`:

**client.ts**:
- Create an `axios` instance with `baseURL = process.env.NEXT_PUBLIC_API_BASE_URL`.
- Request interceptor: read `access_token` from memory (React context or Zustand store). If present, attach `Authorization: Bearer {token}` header.
- Request interceptor: read `csrf_token` cookie using `js-cookie`. Attach `X-CSRF-Token: {token}` header on POST/PATCH/DELETE.
- Response interceptor: on 401 response, attempt silent token refresh via `POST /auth/refresh`. If refresh succeeds, retry original request with new token. If refresh fails, clear tokens and redirect to `/login`.
- `withCredentials: true` on all requests (required for HttpOnly refresh token cookie).

**auth.ts**: typed functions for `register`, `login`, `logout`, `refreshToken`, `requestOTP`, `verifyOTP`, `changePassword`, `resetPasswordRequest`, `resetPasswordConfirm`, `getMe`, `enable2FA`, `verify2FA`, `disable2FA`, `getSessions`, `revokeSession`, `revokeAllSessions`.

**reminders.ts**: typed functions for all reminder endpoints. Request/response types derived from OpenAPI schema.

**documents.ts**: typed functions including `uploadFile(documentID, file: File)` using `FormData`.

**subscriptions.ts**: typed functions for all subscription endpoints.

**goals.ts**: typed functions for all goal endpoints.

**finances.ts**: typed functions for summary, spending, trend, upcoming-bills, export.

**timeline.ts**: typed functions for timeline.

**notifications.ts**: typed functions for notifications list, stats, export.

**family.ts**: typed functions for family members and shared reminders.

**chat.ts**: typed functions for chat messages, usage, intents.

**account.ts**: typed functions for profile, avatar, export, delete.

**billing.ts**: typed functions for plan, checkout.

---

### TASK-025 — Set up auth state management
**Repo:** `woles-frontend`

In `lib/auth/`:

**AuthContext.tsx**:
- React context holding: `user: User | null`, `accessToken: string | null`, `isLoading: boolean`, `isAuthenticated: boolean`.
- `login(email, password)`: call API, store access token in memory (React state), store user in state. Refresh token is in HttpOnly cookie (set by backend, not accessible in JS).
- `logout()`: call API logout, clear all state, redirect to `/login`.
- `refreshAccessToken()`: call `POST /auth/refresh`. On success, update `accessToken` in state.
- Auto-refresh: on mount, if there's a stale access token, attempt refresh once. If no refresh token cookie, redirect to `/login`.
- Wrap entire dashboard layout with `AuthProvider`.

**useAuth.ts**: convenience hook to consume `AuthContext`.

**withAuth.tsx**: HOC (or use Next.js middleware) to protect dashboard routes. Redirect unauthenticated users to `/login`.

`middleware.ts` (Next.js middleware):
- Protect all routes under `/(dashboard)`.
- Check for refresh token cookie existence. If missing, redirect to `/login`.
- Allow public routes and auth routes without redirection.

---

### TASK-026 — Build layout components
**Repo:** `woles-frontend`

In `components/layout/`:

**sidebar.tsx**:
- Fixed left sidebar, width 240px, visible only on `lg:` breakpoint and above.
- Top: Woles logo (`docs/logo/woles_horizontal_logo_uniform.png`) at 120px wide.
- Navigation items with icons (lucide-react): Dashboard, Reminders, Documents, Finances, AI Chat Hub, Family (only if `plan === 'advanced'`), Settings.
- Each nav item: icon + label. Active state: `bg-primary-container text-on-primary-container rounded-md`.
- Footer: user avatar (circle, 32px), display name, plan badge pill (`FREE`, `PREMIUM`, `ADVANCED`).
- "Upgrade Now" button: shown only if `plan === 'free'`. Deep emerald green background. Links to `/billing/checkout`.
- AI Chat Hub item: show usage indicator "(N/10)" in a small muted label next to the item if on free plan.
- Family item: hidden on free and premium plans.

**mobile-nav.tsx**:
- Fixed bottom navigation bar, visible only below `lg:` breakpoint.
- Five items: Dashboard, Reminders, Documents, Finances, More.
- "More" opens a bottom sheet drawer with: Chat, Family, Settings, and "Sign Out" in red.
- Active item uses primary color underline and tinted icon.
- Touch targets: minimum 44×44px per item.

**topbar.tsx**:
- Shown only on mobile (hidden on `lg:`).
- Left: Woles app icon (`docs/logo/woles_app_icon_uniform.png`) at 32px.
- Right: notification bell icon (links to `/notifications`) and avatar circle (opens profile dropdown: Settings, Sign Out).

**dashboard-layout.tsx** (`app/(dashboard)/layout.tsx`):
- Render sidebar (desktop) + topbar (mobile) + main content area.
- `min-h-screen bg-surface flex`.
- Main content: `flex-1 overflow-y-auto p-6 lg:ml-60` (account for sidebar width).
- Wrap with `AuthProvider` and `QueryClientProvider`.

---

## Frontend — Auth Pages

### TASK-027 — Build login and register pages
**Repo:** `woles-frontend`

**`app/(auth)/login/page.tsx`**:
- White card centered on the page, max-width 440px, `rounded-lg shadow-sm border`.
- Top: Woles primary logo, centered.
- Form: Email input, Password input (with show/hide toggle), "Remember me" checkbox.
- "Sign In" button: full-width, primary emerald green.
- Error state: show generic error banner (red border, error message text) on failed login. Do not reveal whether account exists.
- "Forgot password?" link below form → `/login/forgot`.
- "Don't have an account? Sign up" link → `/register`.
- On submit: call `login(email, password)`. On success, redirect to `/dashboard`.
- Rate limit UX: if 429 response, show "Too many attempts. Please try again in X minutes."
- Lockout UX: if `account_locked` error, show "Account temporarily locked. Try again later."

**`app/(auth)/register/page.tsx`**:
- Same card layout.
- Form: Full Name, Email, Password (min 8 chars, show strength indicator), Confirm Password, Timezone selector (default Asia/Jakarta from list of Indonesian cities).
- "Create Account" button: full-width, primary green.
- Submit calls `register(...)`. On success: show "Check your email to verify your account" and redirect to `/login`.
- "Already have an account? Sign in" link.

**`app/(auth)/verify/page.tsx`**:
- Handle email verification link: extract token from URL query param, call verify endpoint, show success or expired message.

**`app/(auth)/login/forgot/page.tsx`**: Email input, "Send Reset Link" button. Show success state.

**`app/(auth)/login/reset/page.tsx`**: New password + confirm form. Token from URL.

---

## Frontend — Dashboard Pages

### TASK-028 — Build dashboard overview page
**Repo:** `woles-frontend`

**`app/(dashboard)/dashboard/page.tsx`**:

- Page title: "Dashboard". Welcome greeting: "Good morning, {user.name}. Here's your life admin summary."
- **Stats row** (3 columns desktop, 2 columns tablet, 1 column mobile via `grid-cols-1 sm:grid-cols-2 lg:grid-cols-3`):
  - Card 1: "Active Reminders" — count from `GET /api/v1/reminders?status=active&per_page=1` (use `pagination.total`). Icon: bell. Green accent.
  - Card 2: "Documents" — count from documents endpoint. Icon: file. Blue accent.
  - Card 3: "Subscriptions" — count + monthly total cost. Icon: credit card. Amber accent.
  - Each stat card: white bg, rounded-lg, 1px border, label-sm text above, headline-md number, subtle icon top-right.
- **Upcoming items** section (below stats):
  - Title: "This Month" with "View All →" link to `/timeline`.
  - Fetch `GET /api/v1/timeline?range=30d&per_page=10`.
  - Each item: colored left-border by type (reminder=green, document=blue, subscription=amber, goal=teal), type badge, title, due date formatted as "in N days" or "on DD MMM".
  - If empty: illustrated empty state with text "Nothing due this month. You're all clear!" and "Add Reminder" button.
- **Quick actions** row: three buttons — "New Reminder", "Add Document", "Set Goal". Each opens respective creation drawer.
- **Recent activity** section: last 5 audit log events. Small table: icon, action label, date. Fetch from `GET /api/v1/account/activity?per_page=5` (add this lightweight endpoint or derive from notifications feed).
- Responsive: on mobile, stats stack to 1 column, quick actions stack to a 1×3 vertical list.

---

### TASK-029 — Build reminders page
**Repo:** `woles-frontend`

**`app/(dashboard)/reminders/page.tsx`**:

- Page header: "Reminders" title + FAB (`+ New Reminder`) top-right.
- **Filter tabs** (horizontally scrollable on mobile): "All Reminders (N)" | "Upcoming (N)" | "Overdue (N)" | "Completed (N)". Each tab count is fetched with a separate status filter.
- **Search bar**: full-width input with search icon, debounced 300ms, passes `search` query param to API.
- **Card grid**: `grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4`.
- **Reminder card** (`components/reminders/reminder-card.tsx`):
  - Category icon top-left (bell for bill, car for vehicle, file for document, star for custom).
  - Status badge top-right: "DUE IN N DAYS" (amber if ≤3 days), "UPCOMING" (green), "OVERDUE" (red), "COMPLETED" (grey), "PAUSED" (blue-grey).
  - Title (title-lg font, 2-line clamp).
  - Recurrence description (label-md, muted): "Every month on the 10th", "One-time, 15 May 2027", etc.
  - Due date (label-sm, muted): "Next: DD MMM YYYY".
  - 3-dot action menu: Edit, Pause/Resume, Mark as Done, Delete (with confirmation).
  - On hover: subtle `shadow-md` lift effect.
- API: `GET /api/v1/reminders?status={tab}&search={q}&sort=next_run_at&order=asc&page={p}&per_page=20`.
- Pagination: "Load More" button at bottom (or infinite scroll with Intersection Observer).
- **Create/Edit Reminder Drawer** (`components/reminders/reminder-drawer.tsx`):
  - Slide-in from right (desktop) / bottom sheet (mobile).
  - Fields: Title, Category (select), Recurrence Type (select), recurrence details (conditional: day of week for weekly, day of month for monthly, interval days for custom), Date/Time picker, Source note (optional).
  - Submit: POST to create or PATCH to update. On success: close drawer, invalidate query cache.
- Empty state: illustrated empty state with "No reminders yet. Send a WhatsApp message like 'Ingatkan bayar internet tiap tanggal 10' to get started."

---

### TASK-030 — Build document vault page
**Repo:** `woles-frontend`

**`app/(dashboard)/documents/page.tsx`**:

- Page header: "Document Vault".
- **Stats row** (3 columns → 1 column on mobile):
  - "Expiring Soon": count from `GET /api/v1/documents?status=active&expiry_within_days=30&per_page=1` (pagination.total).
  - "Security Status": static "Protected" badge with shield icon (green).
  - "Cloud Storage": from `GET /api/v1/documents/storage/usage`. Show "X.X MB used".
- **Filter tabs**: All Files | Vehicles | Identity | Insurance | Financials.
- **Upload button**: "Upload New" top-right, primary green, file icon.
- **Document grid** (`grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4`):
  - **Document card** (`components/documents/document-card.tsx`):
    - Thumbnail: if `file_url` exists, show file icon (PDF/image indicator). If no file, show category illustration.
    - Category badge chip: "VEHICLE" (amber), "IDENTITY" (blue), "INSURANCE" (teal), "FINANCIAL" (purple), "OTHER" (grey).
    - Title (title-lg, 1-line clamp).
    - Last updated date (label-sm, muted).
    - Expiry indicator: "Expires in N days" (amber if ≤30 days, red if ≤7 days), "Valid until YYYY" (green), "EXPIRED" (red chip).
    - 3-dot menu: View File (opens signed URL in new tab), Edit, Delete.
  - **Add New Document card**: dashed border, "+" icon, "Add New Document" text. Same size as regular cards. Opens upload modal on click.
- **Vault Health widget** (right sidebar or below grid on mobile):
  - Title: "Vault Health". Progress bar.
  - Shows completeness: Vehicles ✓/✗, Identity ✓/✗, Insurance ✓/✗, Financials ✓/✗.
  - Data from `GET /api/v1/documents/vault/health`.
- **Woles Tip card** (below grid): light surface card, lightbulb icon. Contextual tip based on missing vault categories.
- **Recent Activity** section (below tip): last 5 document-related audit events.
- **Upload Modal** (`components/documents/upload-modal.tsx`):
  - Fields: Document Type (select with all allowed values), Title, Expiry Date (date picker), Notes (textarea, max 2000 chars), Storage Type, File upload (drag-and-drop + click-to-browse, shows preview for images, filename for PDF).
  - Validate: file ≤10MB, only PDF/JPG/PNG.
  - Submit: first POST to create document, then POST file if file is selected.
- Responsive: grid collapses to 2→1 columns. Stats row stacks. Vault Health moves below grid on mobile.

---

### TASK-031 — Build finances: goal tracker page
**Repo:** `woles-frontend`

**`app/(dashboard)/finances/goals/page.tsx`**:

- Shared finance sub-nav tab bar (rendered in `finances/layout.tsx`):
  - Two tabs: "Goal Tracker" | "Financial Overview". Active tab: bottom border underline in primary color.
- Page header: "Financial Goal Tracker", subtitle: "Tracking your calm journey to financial freedom."
- Top-right button: "Update Progress" (primary green, pencil icon). Opens progress update drawer.
- FAB bottom-left: "+ New Goal" (secondary action button). Opens goal creation drawer.
- **Two-column layout** (desktop): left 65% / right 35%. Stacks on mobile.
- **Left column — top: Active Goal hero card** (`components/finances/goals/active-goal-card.tsx`):
  - `ACTIVE GOAL` chip badge (teal, label-sm).
  - Goal name: `headline-md` Geist font.
  - Target amount: `Rp X.XXX.XXX.XXX` formatted.
  - Progress bar: full-width, emerald fill, animated on load.
  - Below progress bar: "XX.X% Completed" label right-aligned.
  - Stats row (3 mini-cards): "Current Savings | Rp X", "Remaining | Rp X", "Monthly Target | Rp X".
  - Fetch: `GET /api/v1/goals?status=active&per_page=1`.
  - If no active goal: show "No active goal. Set a goal to start tracking!" empty state with "+ New Goal" CTA.
- **Right column — top: Woles Pro promo card** (`components/finances/goals/pro-card.tsx`):
  - Background: `bg-primary-container` (dark emerald `#064e3b`). Rounded-xl. Padding 24px.
  - Icon: chart/rocket icon (white).
  - Title: "Advanced Goal Analytics" (white, title-lg).
  - Description: "Predict exactly when you'll reach your goal based on spending patterns." (muted green `#80bea6`, body-md).
  - Button: "Unlock Insights" (white background, primary-container text, rounded-full). On click: navigate to billing checkout.
  - **Hide** this card when `plan === 'advanced'`.
- **Left column — bottom: All Financial Goals** (`components/finances/goals/goals-list.tsx`):
  - Header: "All Financial Goals" + "View History >" link (fetches `/goals/history`).
  - Each goal row: goal icon (colored circle matching `icon` field), goal name, progress bar (emerald if ≥70%, blue if 30-70%, amber if <30%), "XX%" label.
  - Clicking a row expands an inline detail panel with full stats and Edit/Delete actions.
  - Fetch: `GET /api/v1/goals`.
- **Right column — bottom: Trajectory Outlook widget** (`components/finances/goals/trajectory-widget.tsx`):
  - Card with `PRO` badge chip (emerald, top-right).
  - Locked state: padlock icon, blurred placeholder chart (use CSS `filter: blur(4px)` on a static chart image).
  - Text: "Predictive Growth — Unlock the ability to see future projections based on inflation and savings rates."
  - "Upgrade to Pro" button (full-width, primary green).
  - **Unlock** (show real chart) when `plan === 'advanced'` and use data from `GET /api/v1/finances/trend`.
- **Woles Finance Tip card** (full-width, below all):
  - Surface-low background, lightbulb icon, tip text.
  - Fetch from `GET /api/v1/goals?status=active&include_tip=true`.
- **Create/Update Goal Drawer**:
  - Fields: Title, Icon (visual icon picker with options: love ❤️, emergency 🚨, vehicle 🚗, home 🏠, travel ✈️, other ⭐), Target Amount (number input, formatted with IDR), Monthly Target (optional), Target Date (date picker).
  - For Update Progress: single amount input ("New total savings amount"), submit calls `POST /api/v1/goals/{id}/progress`.

---

### TASK-032 — Build finances: financial overview page
**Repo:** `woles-frontend`

**`app/(dashboard)/finances/overview/page.tsx`**:

- Finance sub-nav tab bar (same as TASK-031 layout).
- Page header: "Financial Overview", subtitle: "Manage your spending and monthly commitments effortlessly."
- Top-right actions: "Export Data" button (secondary style, download icon) + "Add Transaction" button (primary green, plus icon).
- **Two-column layout** (desktop): left 55% / right 45%. Stacks on mobile (hero first).
- **Left column — top: Total Expenses hero card** (`components/finances/overview/total-expenses-card.tsx`):
  - Background: `bg-primary-container` (#064e3b). White/muted-green text.
  - Label: "TOTAL EXPENSES • THIS MONTH" (label-sm, uppercase, muted).
  - Amount: `headline-lg` white, formatted as Rp XX.XXX.XXX.
  - Change indicator: green ↓ if lower, red ↑ if higher than last month. "N% less than last month".
  - Sub-stats row: "Income | Rp X" and "Savings | Rp X" (muted, label-md).
  - Fetch: `GET /api/v1/finances/summary?period=monthly`.
- **Right column — top: Linked Accounts card** (`components/finances/overview/linked-accounts-card.tsx`):
  - White card.
  - "Linked Accounts" header (title-lg).
  - MVP empty state: bank icon, "No accounts linked yet.", "+ Connect Bank (Coming Soon)" button (disabled/outlined, no click action).
  - When V2 bank sync is available: show list of accounts with masked card numbers.
- **Left column — bottom: Category Breakdown card** (`components/finances/overview/category-breakdown-card.tsx`):
  - Header: "Category Breakdown" + period selector dropdown ("Current Month", "Last Month").
  - `grid-cols-2 gap-3` of category tiles: each tile has category icon (colored), name, percentage (label-md bold), IDR amount (label-sm muted).
  - Categories: Household, Utilities, Transport, Others.
  - Fetch: `GET /api/v1/finances/spending?period=monthly`.
- **Right column — bottom: Spending Trend card** (`components/finances/overview/spending-trend-card.tsx`):
  - Header: "Spending Trend".
  - Bar chart (use `recharts` BarChart): x-axis = week labels ("Week 1", "Week 2", etc.), y-axis = IDR amounts (formatted in millions "XM").
  - Bar color: emerald (#003527).
  - Responsive: use `ResponsiveContainer width="100%" height={200}`.
  - Fetch: `GET /api/v1/finances/trend?period=weekly`.
- **Upcoming Bills table** (full-width, below two columns) (`components/finances/overview/upcoming-bills-table.tsx`):
  - Header: "Upcoming Bills" + count badge (e.g., "You have 3 payments due this week").
  - Table (desktop): SERVICE | DUE DATE | AMOUNT | STATUS | ACTION.
  - SERVICE column: subscription logo placeholder (colored initial), name, billing type label.
  - DUE DATE: formatted as "DD MMM YYYY".
  - AMOUNT: formatted as "Rp X.XXX" or "USD X.XX".
  - STATUS badges: `PENDING` (amber chip), `URGENT` (red chip), `SCHEDULED` (grey chip).
  - ACTION: icon buttons (eye for view detail, check for mark paid).
  - Fetch: `GET /api/v1/finances/upcoming-bills?per_page=20`.
  - **Mobile**: table becomes stacked cards. Each card: service name top, due date + amount middle, status badge + action bottom.
  - Pagination: Load More button.
- **Budget Status widget** (right sidebar, desktop only, sticky): small card showing "XX% of monthly goal". Progress bar. Shows `total_expenses / monthly_target_from_active_goal`.

---

### TASK-033 — Build notification history page
**Repo:** `woles-frontend`

**`app/(dashboard)/notifications/page.tsx`**:

- Page header: "Notification History". Right side: "Export PDF" button (secondary style, download icon).
- **Filter row** (two groups on same row, wraps on mobile):
  - Category filter chips (horizontally scrollable): All | Vehicle | Bills | Documents | Property.
  - Date range selector (dropdown): "Last 7 Days" | "Last 30 Days" | "Last 90 Days" | "Custom Range" (opens date-picker popover).
- **Notifications table** (desktop) / **card list** (mobile):
  - Desktop table columns: STATUS | DATE & TIME | CATEGORY | MESSAGE CONTENT | ACTIONS.
  - STATUS badges: "Sent" (green chip), "Read" (blue chip), "Action Taken" (teal chip), "Failed" (red chip).
  - DATE & TIME: "DD MMM YYYY, HH:MM" in user's timezone.
  - CATEGORY: category chip matching the entity_type (vehicle=amber, bills=blue, documents=teal, property=green).
  - MESSAGE CONTENT: truncated to 60 chars with ellipsis, expand on click.
  - ACTIONS: "View Message" (eye icon), "View Receipt" (receipt icon, only for subscription), "Renew Now" (link icon, for documents).
  - Mobile card: each row becomes a stacked card with STATUS badge top-right, message preview, date, and action links below.
- Pagination: standard page selector at bottom.
- **Stats row below table** (2 cards): "Delivery Rate: XX%" (green indicator), "Top Category: Vehicle" (with count).
- Fetch: `GET /api/v1/notifications?category={f}&range={r}&page={p}&per_page=20`.
- On "Export PDF": call `GET /api/v1/notifications/export?format=pdf&range={r}`. Trigger file download via `window.open` or `a[download]`.

---

### TASK-034 — Build family management page
**Repo:** `woles-frontend`

**`app/(dashboard)/family/page.tsx`**:

- Plan gate: if `plan !== 'advanced'`, show upgrade prompt page instead of the actual content. Upgrade prompt: illustration, "This feature requires the Advanced plan." description, "Upgrade to Advanced" button (primary green, links to billing).
- Page header: "Family Management". Right side: "+ Add Family Member" button (primary green).
- **Family Members grid** (`grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4`):
  - **Member card** (`components/family/member-card.tsx`):
    - Avatar circle (64px): initials placeholder with colored background, or actual avatar image.
    - Role badge chip: "Primary" (deep green), "Spouse" (blue), "Parent" (teal), "Child" (amber), "Other" (grey).
    - Name (title-lg).
    - Relation label (label-md muted).
    - "N active reminders" (label-sm muted).
    - "View Details" button (secondary/ghost).
    - 3-dot menu: Edit, Remove (with confirmation: "Removing this member will transfer their items to your account.").
- **Shared Reminders table** (below grid):
  - Header: "Shared Reminders".
  - Table columns: TASK NAME | OWNER (avatar + name) | DUE DATE | STATUS | ACTION.
  - STATUS badges: "Urgent" (red), "Upcoming" (amber), "Safe" (green).
  - Mobile: table becomes stacked card list.
  - Fetch: `GET /api/v1/family/reminders`.
- **Add/Edit Member Drawer**:
  - Fields: Name, Role (select), Relation Label (text, e.g., "Wife", "Dad"), Avatar upload (optional).
  - Submit: POST to create or PATCH to update.

---

### TASK-035 — Build AI chat hub page
**Repo:** `woles-frontend`

**`app/(dashboard)/chat/page.tsx`**:

- **Three-panel layout** on desktop (`lg:` and above):
  - Panel 1 (left, 260px fixed): navigation sidebar (same as main sidebar, but narrower inside chat layout — or reuse main sidebar).
  - Panel 2 (center, 280px): message list panel + intent panel.
  - Panel 3 (right, remaining width): chat area.
- **Mobile**: only the chat area is shown. A back arrow in the topbar navigates to the message list.
- **Message list panel** (`components/chat/message-list-panel.tsx`):
  - Header: "Conversations".
  - List of previous conversation threads (grouped by day). Each thread: date label, first message preview.
  - Clicking a thread scrolls to that date in the main chat area.
  - Below threads: **Intents panel** ("Detected Intents"):
    - Grouped intent chips: "💰 Saving Goals: Dana Nikah 200jt", "📅 Monthly Bills: Netflix, Spotify".
    - Fetch: `GET /api/v1/chat/intents`.
  - Footer: usage indicator: "N of 10 free messages used" with a mini progress bar. "Upgrade to Pro ›" link. Fetch: `GET /api/v1/chat/usage`.
- **Chat area** (`components/chat/chat-area.tsx`):
  - **Free plan quota banner** (sticky top, dismissible): "You are on the Free Plan. N messages remaining this month. [Go Unlimited →]". Show only if `plan === 'free'`.
  - Message list (`overflow-y-auto flex-1`):
    - User messages: right-aligned bubble, `bg-primary` (dark green), white text, `rounded-tl-xl rounded-tr-xl rounded-bl-xl`.
    - Assistant messages: left-aligned bubble, `bg-surface-container`, on-surface text, `rounded-tr-xl rounded-br-xl rounded-bl-xl`.
    - Timestamp below each bubble (label-sm muted).
    - Auto-scroll to bottom on new message.
  - **Input area** (sticky bottom):
    - Text input: `placeholder="Type a message or describe what you want to track..."`.
    - Right icons: paperclip (attachment, V2), microphone (Voice Command, V2 — show as disabled/greyed out), send button (primary green).
    - "Suggest Intent" button (secondary/ghost, sparkle icon) above input: calls `GET /api/v1/chat/intents` and shows a popover of suggestions the user can tap to pre-fill the input.
  - Submit: `POST /api/v1/chat/messages`. Append both user and assistant response messages to the list. Invalidate chat usage query.
  - On quota exceeded (403 response): show inline error with upgrade CTA.

---

### TASK-036 — Build settings page
**Repo:** `woles-frontend`

**`app/(dashboard)/settings/page.tsx`**:

- Three-section layout: secondary sidebar (desktop) or tabs (mobile): "Account Details" | "Connectivity" | "Security & Privacy".
- Desktop: left secondary sidebar (180px) with section list. Right: section content. Mobile: horizontal tabs at top.

**Account Details section**:
- Avatar: circle (80px), current avatar image or initials. "Change Photo" overlay on hover. On click: opens file picker. Upload via `POST /api/v1/account/avatar`. Preview immediately with `URL.createObjectURL`.
- Form fields: First Name, Last Name (split from `name` on load, join on save), Email (read-only, show verified badge), Timezone (select from IANA timezone list, filtered to Indonesian cities + common zones).
- Plan badge: shows current plan with label.
- "Save Changes" button: primary green, full-width. Calls `PATCH /api/v1/account/profile`.
- Success toast: "Profile updated."

**Connectivity Integrations section**:
- Three integration cards:
  - Google Workspace: toggle (off, disabled), "Coming Soon" badge.
  - Financial (Plaid): toggle (off, disabled), "Coming Soon" badge.
  - Smart Home Hub: toggle (off, disabled), "Coming Soon" badge.
- Each card: integration logo, name, description, status label ("Disconnected"), toggle switch.
- MVP: all toggles are visually rendered but non-functional (disabled + tooltip "Coming in a future update").

**Security & Privacy section**:
- **Password**: "Last changed: X days ago." + "Update Password" button → opens password change modal (current password + new password fields).
- **Two-Factor Authentication**: toggle switch + "RECOMMENDED" green badge. When enabled: show TOTP QR code in a modal for setup, then verify TOTP code. On enable: call `POST /auth/2fa/enable` then `POST /auth/2fa/verify`. When disabled: require password confirmation.
- **Active Sessions**: "View Sessions" button → expands a list of sessions (device name, last active time, IP masked). Each session has a "Revoke" button. "Sign Out All Other Sessions" button at bottom (calls `DELETE /auth/sessions`).
- **Danger Zone** (red border section, bottom):
  - "Delete Account" button: red outlined. On click: confirmation modal with "Type DELETE to confirm" text input. On confirm: call `DELETE /api/v1/account`. On success: logout and redirect to landing page.

---

## Frontend — Public Pages

### TASK-037 — Build landing page
**Repo:** `woles-frontend`

**`app/(public)/page.tsx`** — server component, fully server-rendered for SEO:

- `<title>Woles — Semua Urusan Hidup, Satu Chat Aja</title>`.
- Meta description, og:title, og:description, og:image (Woles logo), og:locale: id_ID.
- **Navbar**: Woles horizontal logo left. "Masuk" (Login) + "Mulai Gratis" (Start Free, primary CTA) right. Hamburger on mobile.
- **Hero section**:
  - Headline: "Semua urusan hidupmu, satu chat WhatsApp." (`display-lg` on desktop, `headline-lg-mobile` on mobile).
  - Subheadline: "Ingatkan tagihan, dokumen penting, dan target keuangan — cukup dari WhatsApp. Tanpa install app." (body-lg, muted).
  - CTA buttons: "Mulai via WhatsApp" (primary green, WhatsApp icon, links to WA deep link) and "Lihat Demo" (secondary/ghost).
  - Right side: mockup of AI assistant chat preview card (show example conversation from `docs/woles-prd.md` Section 6.2 journey).
  - Responsive: hero text stacks above mockup on mobile.
- **Problem section**: "Pernah lupa bayar pajak kendaraan? STNK mati mendadak?"
  - Pain point icons grid: 6 tiles with icons — SIM expired, STNK, pajak mobil, servis, subscription, tagihan.
- **How it works**: "3 langkah mudah" — three cards with numbered steps: "Kirim pesan WA" → "Woles simpan & jadwalkan" → "Dapat reminder tepat waktu."
- **Categories grid**: 8 tiles — STNK, SIM, Pajak, Servis Mobil, Passport, Tagihan, Langganan, Cicilan. Each with icon and label.
- **Trust section**: 3 cards — "Data Pribadi Aman" (lock icon), "WhatsApp-First" (WA icon), "Reliable & Human" (heart icon).
- **Pricing section**: 3 pricing cards matching `docs/woles-prd.md` Section 12. Free (Rp0), Premium (Rp39.000/bulan), Advanced (Rp99.000/bulan). Premium card: highlighted with `bg-primary` border. Features list per plan.
- **FAQ section**: accordion with 6-8 FAQ items. Each item is also rendered as JSON-LD `FAQPage` schema in `<script type="application/ld+json">`.
- **CTA section**: "Siap untuk hidup lebih santai?" + "Mulai Gratis via WhatsApp" button.
- **Footer**: Logo, tagline, navigation links (Dashboard, Features, Pricing), legal links (Privacy Policy, Terms), social links (Instagram, Twitter, TikTok).

---

### TASK-038 — Build SEO landing pages
**Repo:** `woles-frontend`

Create one page component per use case. All are server-rendered (`app/(public)/reminder-{slug}/page.tsx`):

- `/reminder-pajak-mobil`
- `/reminder-stnk`
- `/reminder-sim`
- `/reminder-servis-mobil`
- `/reminder-passport`
- `/reminder-subscription`
- `/whatsapp-reminder`

Each page follows the same structure:
1. `<h1>`: keyword-focused (e.g., "Reminder Pajak Mobil Otomatis via WhatsApp").
2. Problem section: explain the pain (1-2 paragraphs, Indonesian).
3. Solution section: how Woles solves it with a WhatsApp command example (show the exact message user would send, styled as a WhatsApp bubble).
4. Benefits: 3-4 bullet points.
5. How it works: same 3-step flow as landing page (reusable component).
6. FAQ: 4-6 questions specific to that use case. Include JSON-LD `FAQPage` schema.
7. CTA: "Mulai Reminder {topic} via WhatsApp" button.
8. Internal links to related pages.

Shared `next/metadata` export per page with unique `title`, `description`, `keywords`, `openGraph`, `canonical`.

---

## Frontend — Shared Components

### TASK-039 — Build shared UI primitives
**Repo:** `woles-frontend`

In `components/ui/`:

**button.tsx**: variants: `primary` (bg-primary text-on-primary), `secondary` (bg-secondary-fixed text-on-secondary), `ghost` (transparent, primary text), `danger` (bg-error text-on-error), `outline` (border-outline). Sizes: `sm`, `md`, `lg`. Loading state with spinner. All variants: `rounded-md` (8px).

**card.tsx**: white background, `rounded-lg` (12px), 1px border (`border-outline-variant`), optional `hover:shadow-md` transition.

**badge.tsx**: pill shape (`rounded-full`). Color variants: `green`, `blue`, `amber`, `red`, `grey`, `teal`. Size: `sm` (12px) and `md` (14px).

**input.tsx**: `rounded-md`, 1px border. Focus state: border transitions to `border-primary` with `ring-2 ring-primary/20`. Error state: `border-error`. Label above input (label-md). Helper text below (label-sm muted). Optional icon left.

**select.tsx**: same styling as input. Dropdown using native `<select>` or a custom dropdown. Mobile: always use native `<select>` for touch UX.

**textarea.tsx**: same as input, multiline. Character count display if `maxLength` prop provided.

**drawer.tsx**: slide-in from right on desktop (320px width), full-screen bottom sheet on mobile. Overlay backdrop. Close on ESC, backdrop click. `framer-motion` or CSS transition for animation.

**modal.tsx**: centered modal on desktop, full-screen on mobile. Header with title and close button. Footer slot for action buttons.

**bottom-sheet.tsx**: slides up from bottom on mobile. Drag handle. Snap points.

**toast.tsx**: top-right toast notifications. Variants: success (green), error (red), info (blue), warning (amber). Auto-dismiss after 4 seconds. Max 3 toasts stacked.

**progress-bar.tsx**: horizontal progress bar. Color prop. Animated fill with `transition-all duration-700`. Shows percentage label optionally.

**avatar.tsx**: circle image with fallback to initials (colored background from name hash). Sizes: `sm` (32px), `md` (48px), `lg` (80px).

**icon-picker.tsx**: grid of icon options for goal icons. Emoji or SVG icons. Selected state: primary border ring.

**pagination.tsx**: "Previous" / "Next" buttons + page number display. Matches the pagination standard from Section 8.4.

**empty-state.tsx**: centered illustration placeholder (SVG), title, description, optional CTA button.

---

### TASK-040 — Integrate React Query for data fetching
**Repo:** `woles-frontend`

In `lib/hooks/`:

- Create a custom hook per resource using `@tanstack/react-query` `useQuery` and `useMutation`:
  - `useReminders(filter, pagination)`: uses `reminders.ts` API client. Key: `['reminders', filter, pagination]`.
  - `useReminder(id)`: key `['reminders', id]`.
  - `useCreateReminder()`: mutation that invalidates `['reminders']` on success.
  - `useUpdateReminder()`, `useDeleteReminder()`, `usePauseReminder()`, `useResumeReminder()`, `useCompleteReminder()`.
  - Same pattern for: documents, subscriptions, goals, notifications, family, chat.
  - `useTimeline(params)`.
  - `useFinancialSummary()`, `useSpendingBreakdown()`, `useSpendingTrend()`, `useUpcomingBills()`.
  - `useChatMessages()`, `useSendMessage()`, `useChatUsage()`, `useChatIntents()`.
  - `useGoalProgress()`.

- Configure `QueryClient` in `app/(dashboard)/layout.tsx` with:
  - `staleTime: 30 * 1000` (30 seconds default).
  - `retry: 1`.
  - `refetchOnWindowFocus: false`.

- Use `useInfiniteQuery` for paginated lists with "Load More" button (reminders, documents, notifications).

---

## Infrastructure and DevOps

### TASK-041 — Configure CI/CD pipelines
**Repos:** both `woles-backend` and `woles-frontend`

**Backend CI** (`.github/workflows/ci-backend.yml`):
- Trigger: `push` to `main` and `develop`, `pull_request` to `main`.
- Jobs:
  1. `lint`: run `golangci-lint`.
  2. `test-unit`: run `go test ./tests/unit/...`.
  3. `test-integration`: spin up Postgres + Redis using `docker-compose`, run `go test ./tests/integration/... -tags integration`.
  4. `build`: `go build ./cmd/...`.
  5. `docker-build`: build Docker image and push to registry (on `main` branch only).

**Frontend CI** (`.github/workflows/ci-frontend.yml`):
- Trigger: same as backend.
- Jobs:
  1. `lint`: `npm run lint`.
  2. `type-check`: `npm run type-check`.
  3. `build`: `npm run build`.

---

### TASK-042 — Configure Docker production setup
**Repo:** `woles-backend`

Create `docker-compose.prod.yml`:
- Services: `api`, `worker`, `scheduler`, `postgres`, `redis`, `rabbitmq`, `minio`, `nginx`.
- `api`: image from CI registry, env file `.env.prod`, restart `always`, health check `GET /health`.
- `worker`: same image as `api`, entrypoint `./woles-worker`.
- `scheduler`: same image, entrypoint `./woles-scheduler`.
- `nginx`: `nginx:alpine`, mount `nginx.conf`, ports 80 and 443, SSL via Certbot.
- All services on a private Docker network. Only `nginx` exposes public ports.
- `postgres`: named volume `pgdata`, no public port exposure.
- `redis`: named volume `redisdata`, no public port exposure.
- `minio`: named volumes, console on port 9001 (internal only, behind nginx for admin).

Create `nginx.conf`:
- HTTPS only. HTTP → HTTPS redirect.
- `server_name api.woles.id`: proxy to `api:8080`.
- `server_name www.woles.id woles.id`: proxy to `frontend:3000` (or serve static files if exported).
- Set `proxy_set_header` for `X-Real-IP` and `X-Forwarded-For`.

---

## Quality Assurance

### TASK-043 — Frontend responsive testing checklist
**Repo:** `woles-frontend`

For each page listed below, manually verify on Chrome DevTools at breakpoints 375px (iPhone SE), 768px (iPad), 1280px (desktop), 1440px (large desktop). All items must pass before marking a page done.

Pages to verify: Dashboard, Reminders, Document Vault, Goal Tracker, Financial Overview, Notification History, Family Management, AI Chat Hub, Settings, Landing Page.

Checklist per page:
- [ ] No horizontal scroll at 375px.
- [ ] Bottom nav visible and all touch targets ≥44×44px at 375px.
- [ ] Sidebar hidden at 375px. Sidebar visible at 1280px.
- [ ] All tables collapse to card lists at 375px.
- [ ] All grids respect `grid-cols-1 sm:grid-cols-2 lg:grid-cols-3` breakpoints.
- [ ] Modals and drawers: full-screen at 375px, centered at 1280px.
- [ ] Typography: `headline-lg-mobile` (28px) at mobile, `headline-lg` (32px) at desktop.
- [ ] Woles logo correct variant: horizontal logo in sidebar, app icon in mobile topbar.
- [ ] No text overflow or truncation loss on any card.
- [ ] Empty states render correctly when no data.
- [ ] Loading skeletons show on all data-fetching components.
- [ ] Toast notifications appear at correct position and auto-dismiss.

---

### TASK-044 — Security review checklist
**Repos:** both

Before any public launch, verify all of the following:

**Backend:**
- [ ] All passwords hashed with argon2id, never stored plain.
- [ ] Account lockout after 5 failed logins.
- [ ] JWT uses RS256, not HS256 or none.
- [ ] Refresh token stored as bcrypt hash, not plain.
- [ ] Refresh token rotation active; reuse triggers family revocation.
- [ ] TOTP secret stored encrypted (AES-256-GCM).
- [ ] All list endpoints return ownership-filtered results only.
- [ ] IDOR: accessing another user's resource returns 404.
- [ ] CSRF: all non-GET mutation endpoints require X-CSRF-Token.
- [ ] Rate limiting active on all auth endpoints.
- [ ] WhatsApp webhook HMAC-SHA256 signature verified before processing.
- [ ] Payment webhook signature verified before processing.
- [ ] File uploads: MIME type + file size validated server-side.
- [ ] Signed URLs for document files (no public bucket access).
- [ ] No secrets committed to repository.
- [ ] Security headers set on all responses (CSP, HSTS, X-Content-Type-Options, X-Frame-Options).
- [ ] All SQL queries parameterized.
- [ ] User input sanitized (HTML stripped) before storage.
- [ ] Audit log written for: login, logout, password change, account deletion, plan change.

**Frontend:**
- [ ] Access token stored in memory only (no localStorage, no sessionStorage).
- [ ] Refresh token in HttpOnly cookie (not accessible via JS).
- [ ] `dangerouslySetInnerHTML` never used with user-supplied content.
- [ ] CSP header configured in Next.js `headers()`.
- [ ] All forms validate client-side with Zod (defense in depth alongside server validation).
- [ ] Sensitive pages redirect to `/login` if unauthenticated (Next.js middleware).
- [ ] CORS `withCredentials: true` and no wildcard origin.

---

## File and Asset Management

### TASK-045 — Copy logo assets to frontend repository
**Repo:** `woles-frontend`

Copy all logo files from `docs/logo/` into the frontend repo:
- `public/logo/woles_primary_logo.png` — used on login/register pages and landing page.
- `public/logo/woles_horizontal_logo.png` — used in the sidebar.
- `public/logo/woles_app_icon.png` — used in the mobile topbar and as favicon.
- `public/logo/woles_icon_only.png` — used as the PWA icon and og:image fallback.

Set up Next.js metadata in `app/layout.tsx`:
- `icons.icon`: `/logo/woles_app_icon.png`
- `icons.apple`: `/logo/woles_app_icon.png`
- `openGraph.images`: `/logo/woles_primary_logo.png`

Create `public/favicon.ico` from `woles_app_icon.png` (convert to multi-size ICO).

---

## Done Criteria

A milestone is complete when:
1. All tasks in that milestone pass their respective test suites.
2. All API endpoints in the milestone are documented in `openapi.yaml` and Postman collection is regenerated.
3. All UI pages in the milestone pass the responsive checklist from TASK-043.
4. The security checklist in TASK-044 passes for all affected endpoints.
5. Docker Compose (`docker-compose.yml`) starts all services cleanly with `docker compose up`.

---

## Milestone Mapping

| Milestone | Tasks |
|---|---|
| M1: Foundation | TASK-001, TASK-002, TASK-003, TASK-004, TASK-024, TASK-025, TASK-026, TASK-039, TASK-041, TASK-042, TASK-045 |
| M2: Auth + Security | TASK-005 (identity ports), TASK-006, TASK-016 (auth handler), TASK-017, TASK-018 (users/refresh_tokens), TASK-019, TASK-027, TASK-044 |
| M3: Reminder MVP | TASK-007, TASK-008 (reminder part), TASK-013 (reminder queries), TASK-016 (reminder handler), TASK-020, TASK-022, TASK-029, TASK-040 (reminder hooks), TASK-043 (reminders page) |
| M4: Documents + Subscriptions | TASK-008 (full), TASK-009, TASK-016 (doc+sub handlers), TASK-021, TASK-030, TASK-031 (subscription part), TASK-043 (documents page) |
| M5: Finances + Goals | TASK-010, TASK-011, TASK-013, TASK-016 (goal+finance handlers), TASK-031, TASK-032, TASK-043 (finances pages) |
| M6: Notifications + Timeline | TASK-012, TASK-016 (timeline+notif handlers), TASK-028 (dashboard timeline), TASK-033, TASK-043 (notifications page) |
| M7: Family + Chat | TASK-014, TASK-015, TASK-016 (family+chat handlers), TASK-034, TASK-035, TASK-043 (family+chat pages) |
| M8: Settings + Account | TASK-016 (account handler), TASK-036, TASK-043 (settings page) |
| M9: Public Beta | TASK-022 (full OpenAPI), TASK-023, TASK-037, TASK-038, TASK-044 (full security review) |
