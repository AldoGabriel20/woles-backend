# Woles System Design

## 1. Overview

Woles is a WhatsApp-first personal life administration assistant with a companion web dashboard. The system allows users to create reminders, document expiry trackers, subscription records, financial goals, and timeline queries through WhatsApp and web. Notifications are delivered primarily through WhatsApp.

The system must be reliable before it becomes intelligent. A missed reminder is more damaging than a missing AI feature, so the architecture prioritizes durable storage, idempotent scheduling, retryable notification delivery, and clear operational visibility.

## 2. Product Requirements Covered

This design supports:

- WhatsApp onboarding and commands.
- Natural language intent extraction.
- Reminder creation and recurring scheduling.
- Important document expiry reminders.
- Subscription tracking.
- Financial goals.
- Life Timeline aggregation.
- Web dashboard.
- WhatsApp reminder notifications.
- Subscription plan limits.
- SEO-driven public website pages.

## 3. Recommended Technology Stack

## 3.1 Backend

Recommended: **Golang with Prabogo**.

Reasoning:

- Go is reliable, efficient, and simple to deploy.
- Reminder and notification workloads benefit from Go's concurrency model.
- Prabogo uses Hexagonal Architecture, also known as Ports and Adapters.
- Prabogo provides plug-and-play infrastructure components.
- Prabogo has CLI generation for HTTP adapters, PostgreSQL adapters, Redis cache adapters, RabbitMQ adapters, and Temporal workflow adapters.
- The architecture keeps business logic independent from WhatsApp, database, queue, cache, and AI providers.

Relevant Prabogo capabilities from documentation:

- Hexagonal Architecture.
- Domain, port, adapter, and application layers.
- PostgreSQL migration and database adapter generation.
- Fiber HTTP inbound adapter generation.
- RabbitMQ inbound and outbound message adapter generation.
- Redis cache adapter generation.
- Temporal workflow adapter generation.
- Docker support.
- Unit testing and mock generation.

Recommended backend style:

- Use Prabogo for the main backend service.
- Keep domains separated: identity, reminder, document, subscription, goal, timeline, notification, billing, and WhatsApp.
- Start as a modular monolith, not microservices.
- Split services only when scaling pressure is proven.

## 3.2 Frontend

Recommended: **Next.js**.

Reasoning:

- Strong fit for public SEO pages and authenticated app dashboard in one framework.
- Supports server-side rendering and static generation for SEO landing pages.
- Good ecosystem for modern UI, authentication, analytics, and payment flows.
- Easy deployment to Vercel, self-hosted Node, or container platforms.

Suggested frontend setup:

- Next.js App Router.
- TypeScript.
- Tailwind CSS or a carefully customized design system.
- shadcn/ui only if customized enough to avoid a generic SaaS look.
- Server components for SEO and content pages.
- Client components for dashboard interactions.

Alternative:

- SvelteKit can be considered if the team wants a lighter and highly polished frontend experience.
- For this product, Next.js remains the safer recommendation because SEO, hiring, and ecosystem support matter.

## 3.3 Database

Recommended: **PostgreSQL**.

PostgreSQL should be the source of truth for:

- Users.
- WhatsApp identities.
- Reminders.
- Reminder occurrences.
- Documents.
- Subscriptions.
- Goals.
- Notifications.
- Billing plans.
- Audit logs.

Reasoning:

- Strong transactional guarantees.
- Good support for relational data.
- Reliable indexing for timeline queries.
- Works well with migrations and reporting.
- Can scale far enough for the early and mid-stage product.

## 3.4 Cache

Recommended: **Redis**.

Use Redis for:

- Rate limiting.
- Short-lived session or token cache.
- Idempotency keys for webhook processing.
- Temporary intent extraction cache.
- Distributed locks for schedulers, if needed.
- Fast counters for plan limits, with PostgreSQL as source of truth.

Do not use Redis as the primary source of truth for reminders or notifications.

## 3.5 Queue and Workflow Engine

### MVP Recommendation: RabbitMQ

RabbitMQ is a practical MVP choice because Prabogo supports RabbitMQ adapters and it is stable for event-driven messaging.

Use RabbitMQ for:

- `reminder.created`
- `reminder.updated`
- `notification.scheduled`
- `notification.send_requested`
- `notification.sent`
- `notification.failed`
- `whatsapp.message_received`
- `intent.extraction_requested`

RabbitMQ strengths:

- Mature and well-known.
- Good routing model.
- Works well for event consumers and retry queues.
- Easy enough to self-host for MVP.

RabbitMQ limitations:

- It is not a full durable workflow engine.
- Long-running reminders months or years into the future should not rely only on delayed messages.
- Operational care is needed for retry, dead-letter queues, and monitoring.

### Scale Recommendation: Temporal for Durable Workflows

For a larger system, **Temporal** is a strong recommendation for reminder workflows because reminders are long-running, stateful, and need reliability across days, months, or years.

Use Temporal for:

- Long-running reminder schedules.
- Document expiry reminder workflows.
- Recurring reminder workflows.
- Notification retry workflows.
- Human-in-the-loop flows such as clarification and confirmation.

Temporal strengths:

- Durable workflow state.
- Built-in retries.
- Timers that survive process restarts.
- Strong fit for recurring and future scheduled work.
- Good visibility into workflow state.

Temporal tradeoffs:

- More operational complexity than RabbitMQ.
- Requires learning workflow patterns.
- May be overkill for the first private MVP.

### Final Queue Recommendation

Start with RabbitMQ plus PostgreSQL-backed scheduled jobs for MVP. Design ports so the scheduling engine can later move to Temporal without rewriting domain logic.

For production scale, migrate critical reminder scheduling to Temporal while keeping RabbitMQ for general events.

## 3.6 AI Provider

MVP AI should be limited to intent extraction and classification.

Recommended pattern:

- First use deterministic parsing for common Indonesian reminder patterns.
- Use an LLM only when rules cannot confidently parse the message.
- Store extraction confidence and original text.
- Ask for clarification if confidence is low.
- Keep extracted output as strict JSON.

This keeps AI cost predictable and reduces hallucination risk.

## 3.7 WhatsApp Provider

Options:

- Meta WhatsApp Cloud API.
- Twilio WhatsApp API.
- Local Indonesian WhatsApp Business API providers such as Qontak, Mekari, Wablas, or similar.

Recommended selection criteria:

- Official API compliance.
- Template approval support.
- Clear pricing.
- Webhook reliability.
- Delivery status callbacks.
- Indonesian support and billing.
- Ability to handle production volume.

Avoid unofficial WhatsApp automation for a public product because it can cause account bans and reliability issues.

## 4. High-Level Architecture

```text
User WhatsApp
    |
    v
WhatsApp Provider Webhook
    |
    v
API Gateway / Backend HTTP Adapter
    |
    v
WhatsApp Inbound Service
    |
    v
Intent Extraction Service
    |
    v
Application Services
    |
    +--> PostgreSQL
    +--> Redis
    +--> RabbitMQ
              |
              v
       Notification Worker
              |
              v
       WhatsApp Provider API
              |
              v
       User WhatsApp

Web User
    |
    v
Next.js Web App
    |
    v
Backend API
    |
    v
Application Services
    |
    v
PostgreSQL / Redis / Queue
```

## 5. Backend Architecture

Use Prabogo's Hexagonal Architecture.

Recommended structure:

```text
cmd/
  main.go
internal/
  domain/
    reminder/
    document/
    subscription/
    goal/
    timeline/
    notification/
    billing/
    identity/
    whatsapp/
    family/
    chat/
  application/
    reminder/
    document/
    subscription/
    goal/
    timeline/
    notification/
    billing/
    intent/
    family/
    chat/
  port/
    inbound/
      http/
      message/
      command/
    outbound/
      database/
      cache/
      message/
      whatsapp/
      ai/
      payment/
      storage/
  adapter/
    inbound/
      http_fiber/
      rabbitmq/
    outbound/
      postgres/
      redis/
      rabbitmq/
      whatsapp_provider/
      ai_provider/
      payment_provider/
      storage_provider/
  migration/
  model/
utils/
tests/
```

Core rule:

- Domain logic must not import WhatsApp, Redis, RabbitMQ, HTTP, or AI SDKs.
- Application services coordinate domain logic and ports.
- Adapters implement infrastructure details.

## 6. Domain Modules

## 6.1 Identity Domain

Responsibilities:

- User identity.
- WhatsApp phone identity.
- Account status.
- Timezone.
- Plan assignment.

Important rules:

- A WhatsApp phone number can be linked to one primary user account.
- A user may later add email login for web access.
- Timezone must be stored per user for reminder scheduling.

## 6.2 Reminder Domain

Responsibilities:

- Reminder entity.
- Recurrence rules.
- Next occurrence calculation.
- Reminder status: active, paused, archived.
- Completion history.

Supported MVP recurrence:

- one_time
- daily
- weekly
- monthly
- yearly
- custom_interval

## 6.3 Document Domain

Responsibilities:

- Document type.
- Expiry date.
- Reminder offsets.
- Expiry status.
- File attachment (uploaded PDF, JPG, or PNG image of the physical document).
- Storage type: physical only, uploaded digital copy, or scan-verified.
- Vault health score: completeness based on how many documents a user has added.

Default offsets:

- 30 days before expiry.
- 7 days before expiry.
- 1 day before expiry.

Document categories (matching the UI vault grouping):

- vehicles (STNK, BPKB, vehicle insurance)
- identity (SIM, passport, visa, KTP)
- insurance (health, life, vehicle)
- financials (tax, investment statements, other)
- other

## 6.4 Subscription Domain

Responsibilities:

- Subscription record.
- Amount and currency.
- Billing cycle.
- Next billing date.
- Active or archived state.

## 6.5 Goal Domain

Responsibilities:

- Financial goal target.
- Current progress.
- Monthly savings target.
- Target date.
- Goal milestones.
- Goal icon or category for visual grouping (love, emergency, vehicle, home, travel, other).

## 6.6 Timeline Domain

Responsibilities:

- Aggregate all upcoming life-admin items.
- Normalize reminders, documents, subscriptions, and goals into timeline items.
- Support date range filters.
- Sort and group by due date.

The timeline should be read-optimized because it is the product's killer feature.

## 6.7 Notification Domain

Responsibilities:

- Notification scheduling.
- Notification attempt tracking.
- Delivery status.
- Retry rules.
- Duplicate prevention.
- Notification history for the UI: list, filter by category and date range, export as PDF.

Important rule:

- Notification delivery must be idempotent. The same notification occurrence should not be sent twice unless explicitly retried after a failed attempt.

## 6.8 Family Domain

Responsibilities:

- Family member registry (spouse, parent, child, custom relation).
- Each family member is linked to the primary account holder.
- Family members have their own reminder and document ownership.
- Shared reminder view aggregates tasks across all family members.
- Member access management: primary user can view, edit, or remove family members.

This is an Advanced plan feature.

## 6.9 AI Chat Domain

Responsibilities:

- In-app chat interface (AI Chat Hub) as an alternative to WhatsApp for command input.
- Message history per user.
- Intent panel: sidebar that groups detected intents from recent conversations (e.g., Saving Goals, Monthly Bills).
- Usage quota: Free plan gets a limited number of AI chat messages per month. Quota is tracked per user.
- Voice command support (V2).
- Suggest intent feature: surface likely intents before the user types.

This domain shares the intent extraction port with the WhatsApp inbound domain.

## 7. Data Model

## 7.1 users

| Field | Type | Notes |
| --- | --- | --- |
| id | uuid | Primary key |
| email | varchar | Nullable for WhatsApp-only users |
| phone | varchar | Normalized phone number |
| password_hash | varchar | Nullable; argon2id hash for email+password login |
| name | varchar | Optional display name |
| avatar_url | varchar | Nullable; profile photo URL (object storage) |
| timezone | varchar | Default Asia/Jakarta |
| plan | varchar | free, premium, advanced |
| account_status | varchar | active, suspended, deleted |
| failed_login_count | int | Default 0; reset on successful login |
| locked_until | timestamp | Nullable; set after 5 consecutive failed logins |
| totp_secret | varchar | Nullable; encrypted TOTP secret for 2FA |
| totp_enabled | boolean | Default false |
| created_at | timestamp | |
| updated_at | timestamp | |

## 7.2 whatsapp_identities

| Field | Type | Notes |
| --- | --- | --- |
| id | uuid | Primary key |
| user_id | uuid | FK users.id |
| phone | varchar | WhatsApp phone number |
| provider | varchar | meta, twilio, qontak, etc. |
| provider_contact_id | varchar | Nullable |
| status | varchar | active, blocked, disconnected |
| created_at | timestamp | |

## 7.3 reminders

| Field | Type | Notes |
| --- | --- | --- |
| id | uuid | Primary key |
| user_id | uuid | FK users.id |
| title | varchar | Reminder title |
| category | varchar | bill, vehicle, document, custom |
| recurrence_type | varchar | one_time, daily, weekly, monthly, yearly, custom_interval |
| recurrence_rule | jsonb | Flexible recurrence detail |
| next_run_at | timestamp | Indexed |
| timezone | varchar | User timezone at creation |
| status | varchar | active, paused, archived |
| source | varchar | whatsapp, web, system |
| original_text | text | Original user message |
| created_at | timestamp | |
| updated_at | timestamp | |

## 7.4 reminder_occurrences

| Field | Type | Notes |
| --- | --- | --- |
| id | uuid | Primary key |
| reminder_id | uuid | FK reminders.id |
| user_id | uuid | FK users.id |
| scheduled_at | timestamp | Indexed |
| completed_at | timestamp | Nullable |
| status | varchar | scheduled, sent, done, skipped, failed |
| created_at | timestamp | |

## 7.5 documents

| Field | Type | Notes |
| --- | --- | --- |
| id | uuid | Primary key |
| user_id | uuid | FK users.id |
| family_member_id | uuid | Nullable; FK family_members.id if owned by a family member |
| document_type | varchar | stnk, bpkb, vehicle_insurance, sim, passport, visa, ktp, health_insurance, life_insurance, tax, investment, other |
| vault_category | varchar | vehicles, identity, insurance, financials, other |
| title | varchar | Display name |
| expiry_date | date | Indexed; nullable for documents without expiry |
| reminder_offsets | int[] | Days before expiry; default [30, 7, 1] |
| notes | text | Nullable |
| storage_type | varchar | physical, digital, scan_verified |
| file_url | varchar | Nullable; object storage URL of uploaded file |
| file_size_bytes | int | Nullable |
| file_mime_type | varchar | Nullable; application/pdf, image/jpeg, image/png |
| status | varchar | active, archived |
| created_at | timestamp | |
| updated_at | timestamp | |

## 7.6 subscriptions

| Field | Type | Notes |
| --- | --- | --- |
| id | uuid | Primary key |
| user_id | uuid | FK users.id |
| name | varchar | Subscription name |
| amount | numeric | Billing amount |
| currency | varchar | IDR, USD, etc. |
| billing_cycle | varchar | monthly, yearly, custom |
| next_billing_at | timestamp | Indexed |
| category | varchar | entertainment, productivity, bill, other |
| status | varchar | active, archived, canceled |
| created_at | timestamp | |
| updated_at | timestamp | |

## 7.7 goals

| Field | Type | Notes |
| --- | --- | --- |
| id | uuid | Primary key |
| user_id | uuid | FK users.id |
| title | varchar | Goal name |
| icon | varchar | Nullable; love, emergency, vehicle, home, travel, other |
| target_amount | numeric | |
| current_amount | numeric | Default 0 |
| monthly_target | numeric | Nullable; monthly savings contribution target |
| currency | varchar | Default IDR |
| target_date | date | Nullable |
| status | varchar | active, completed, archived |
| created_at | timestamp | |
| updated_at | timestamp | |

## 7.8 notifications

| Field | Type | Notes |
| --- | --- | --- |
| id | uuid | Primary key |
| user_id | uuid | FK users.id |
| entity_type | varchar | reminder, document, subscription, goal |
| entity_id | uuid | Related entity id |
| occurrence_id | uuid | Nullable |
| channel | varchar | whatsapp, email, web_push |
| scheduled_at | timestamp | Indexed |
| sent_at | timestamp | Nullable |
| status | varchar | scheduled, sending, sent, failed, canceled |
| idempotency_key | varchar | Unique |
| provider_message_id | varchar | Nullable |
| failure_reason | text | Nullable |
| retry_count | int | Default 0 |
| created_at | timestamp | |
| updated_at | timestamp | |

## 7.9 user_sessions

Tracks active authenticated sessions for the Security → Active Sessions UI in Settings.

| Field | Type | Notes |
| --- | --- | --- |
| id | uuid | Primary key |
| user_id | uuid | FK users.id |
| refresh_token_id | uuid | FK refresh_tokens.id |
| device_name | varchar | Nullable; e.g., Chrome on MacOS, iPhone Safari |
| ip_address | varchar | Masked or hashed |
| user_agent | varchar | Truncated |
| last_active_at | timestamp | |
| created_at | timestamp | |

## 7.10 family_members

| Field | Type | Notes |
| --- | --- | --- |
| id | uuid | Primary key |
| owner_user_id | uuid | FK users.id; the primary account holder |
| name | varchar | Family member display name |
| role | varchar | primary, spouse, parent, child, other |
| relation_label | varchar | Freeform label: Self, Wife, Father, Son, etc. |
| avatar_url | varchar | Nullable |
| created_at | timestamp | |
| updated_at | timestamp | |

## 7.11 chat_messages

Stores the AI Chat Hub conversation history.

| Field | Type | Notes |
| --- | --- | --- |
| id | uuid | Primary key |
| user_id | uuid | FK users.id |
| role | varchar | user, assistant |
| content | text | Message text |
| detected_intent | jsonb | Nullable; structured intent if extracted |
| created_at | timestamp | |

## 7.12 chat_usage

Tracks monthly AI Chat message quota per user.

| Field | Type | Notes |
| --- | --- | --- |
| id | uuid | Primary key |
| user_id | uuid | FK users.id |
| month | date | First day of the billing month |
| messages_used | int | Default 0 |
| quota | int | From plan: free=10, premium=unlimited |
| updated_at | timestamp | |

## 7.13 inbound_messages

| Field | Type | Notes |
| --- | --- | --- |
| id | uuid | Primary key |
| user_id | uuid | Nullable before user resolution |
| channel | varchar | whatsapp |
| provider_message_id | varchar | Unique |
| from_phone | varchar | |
| raw_text | text | |
| parsed_intent | jsonb | Nullable |
| processing_status | varchar | received, parsed, handled, failed |
| created_at | timestamp | |

## 7.14 plans and usage_limits

Use these tables to enforce Free, Premium, and Advanced limits.

Plan limits can also be cached in Redis, but PostgreSQL remains the source of truth.

## 8. API Design

## 8.1 Authentication

### Password Hashing

- All passwords are hashed with **argon2id** before storing. Never store plaintext passwords. Never log passwords.
- Recommended argon2id parameters: memory=64MB, iterations=3, parallelism=2, salt length=16 bytes, output length=32 bytes.
- Password comparison must use constant-time comparison to prevent timing attacks.
- Minimum password length is 8 characters. Enforce on server side, not only frontend.
- Account lockout: after 5 consecutive failed login attempts, set `locked_until` to 15 minutes from now. Return a generic error; do not reveal whether the account exists.

### Web Users: Email + Password

- Register with email and password. Send email verification link before allowing full access.
- Login endpoint validates email + password, then issues JWT pair if credentials are correct and account is active.

### Web Users: WhatsApp OTP

- WhatsApp-only users authenticate via a 6-digit OTP sent to their registered phone number.
- OTP is valid for 5 minutes and is single-use. Store the OTP as a hash in Redis with a TTL.
- Magic link tokens are cryptographically random (32 bytes), stored as a hash in PostgreSQL, valid for 15 minutes, and single-use.

### JWT Token Design

Issue two tokens on successful authentication:

- **Access token**: short-lived JWT, expires in **15 minutes**. Sent in `Authorization: Bearer <token>` header.
- **Refresh token**: long-lived opaque token, expires in **30 days**. Stored in an `HttpOnly`, `Secure`, `SameSite=Strict` cookie. Also stored as a bcrypt hash in the `refresh_tokens` PostgreSQL table for server-side revocation.

Access token payload example:

```json
{
  "sub": "user-uuid",
  "jti": "unique-token-id",
  "iat": 1750000000,
  "exp": 1750000900,
  "plan": "premium",
  "tz": "Asia/Jakarta"
}
```

JWT signing rules:

- Algorithm: **RS256** (asymmetric RSA). Private key signs; public key verifies.
- Never use `none` or `HS256` with a shared secret in a multi-service environment.
- Include a `kid` (key ID) in the JWT header to support zero-downtime key rotation.
- Rotate signing keys periodically (recommended: every 90 days).

### Refresh Token Rotation

- Issue a new refresh token on every `/auth/refresh` call. Invalidate the old one immediately.
- Detect reuse of an already-invalidated refresh token. When detected, revoke the entire token family for that user (indicates possible token theft). Force re-authentication.

### Token Revocation

- Store active refresh token records in a `refresh_tokens` PostgreSQL table (token hash, user id, expires at, revoked at).
- On logout: mark the refresh token as revoked.
- On password change or account suspension: revoke all refresh tokens for that user.
- Access tokens cannot be revoked individually due to their short lifetime. Use short expiry and check token version or user status in middleware if needed.

### `refresh_tokens` table

| Field | Type | Notes |
| --- | --- | --- |
| id | uuid | Primary key |
| user_id | uuid | FK users.id |
| token_hash | varchar | bcrypt hash of the opaque token |
| family_id | uuid | Token family for rotation/reuse detection |
| expires_at | timestamp | |
| revoked_at | timestamp | Nullable |
| created_at | timestamp | |

### Internal Service-to-Service

- Use a static bearer API key stored in a secret manager, or mTLS.
- Do not use user JWTs for internal service calls.
- Prabogo also supports Authentik JWT if a more complete identity provider is desired later.

## 8.2 Public Web API

All authenticated endpoints require a valid access token in the `Authorization: Bearer <token>` header.

All endpoints are versioned under `/api/v1/`. When a breaking change is introduced, the version increments to `/api/v2/` while `/api/v1/` remains active during the migration window.

### Authentication Endpoints

```http
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/logout
POST /api/v1/auth/refresh
POST /api/v1/auth/otp/request
POST /api/v1/auth/otp/verify
POST /api/v1/auth/password/change
POST /api/v1/auth/password/reset/request
POST /api/v1/auth/password/reset/confirm
GET  /api/v1/auth/me
POST /api/v1/auth/2fa/enable
POST /api/v1/auth/2fa/verify
POST /api/v1/auth/2fa/disable
GET  /api/v1/auth/sessions
DELETE /api/v1/auth/sessions/{id}
DELETE /api/v1/auth/sessions
```

### Reminders

```http
POST   /api/v1/reminders
GET    /api/v1/reminders?page=1&per_page=20&status=active&sort=next_run_at&order=asc
GET    /api/v1/reminders/{id}
PATCH  /api/v1/reminders/{id}
DELETE /api/v1/reminders/{id}
POST   /api/v1/reminders/{id}/pause
POST   /api/v1/reminders/{id}/resume
POST   /api/v1/reminders/{id}/complete
```

### Documents

```http
POST   /api/v1/documents
GET    /api/v1/documents?page=1&per_page=20&status=active&vault_category=vehicles&sort=expiry_date&order=asc
GET    /api/v1/documents/{id}
PATCH  /api/v1/documents/{id}
DELETE /api/v1/documents/{id}
POST   /api/v1/documents/{id}/file
DELETE /api/v1/documents/{id}/file
GET    /api/v1/documents/storage/usage
GET    /api/v1/documents/vault/health
```

File upload notes:

- `POST /api/v1/documents/{id}/file` accepts `multipart/form-data`.
- Allowed MIME types: `application/pdf`, `image/jpeg`, `image/png`.
- Maximum file size: 10 MB.
- Uploaded file is stored in object storage (S3-compatible). The URL is saved in `documents.file_url`.
- File is served through a signed URL with a short expiry to prevent unauthorized access.

### Subscriptions

```http
POST   /api/v1/subscriptions
GET    /api/v1/subscriptions?page=1&per_page=20&status=active&sort=next_billing_at&order=asc
GET    /api/v1/subscriptions/{id}
PATCH  /api/v1/subscriptions/{id}
DELETE /api/v1/subscriptions/{id}
POST   /api/v1/subscriptions/{id}/archive
```

### Goals

```http
POST   /api/v1/goals
GET    /api/v1/goals?page=1&per_page=20&status=active
GET    /api/v1/goals/{id}
PATCH  /api/v1/goals/{id}
DELETE /api/v1/goals/{id}
POST   /api/v1/goals/{id}/progress
GET    /api/v1/goals/history
```

### Financial Overview

```http
GET    /api/v1/finances/summary?period=monthly
GET    /api/v1/finances/spending?period=monthly
GET    /api/v1/finances/trend?period=weekly
GET    /api/v1/finances/upcoming-bills?page=1&per_page=20
GET    /api/v1/finances/export?format=csv&period=monthly
```

Financial summary response:

```json
{
  "period": "2026-06",
  "total_expenses": 12450000,
  "income": 25000000,
  "savings": 4500000,
  "currency": "IDR",
  "change_vs_last_period_pct": -8.4
}
```

Category breakdown response:

```json
{
  "categories": [
    { "name": "Household", "amount": 5600000, "pct": 45 },
    { "name": "Utilities",  "amount": 2700000, "pct": 22 },
    { "name": "Transport", "amount": 2200000, "pct": 18 },
    { "name": "Others",    "amount": 1900000, "pct": 15 }
  ]
}
```

Note: Linked bank accounts (Connect Bank, income/savings auto-sync) is a **V2 feature** using Plaid or a local Indonesian open-banking provider. For MVP, expenses are derived from subscription records and manual entries.

### Timeline

```http
GET /api/v1/timeline?from=2026-09-01&to=2026-09-30&page=1&per_page=50
GET /api/v1/timeline?range=90d&page=1&per_page=50
```

Example response:

```json
{
  "items": [
    {
      "id": "uuid",
      "type": "subscription",
      "title": "Netflix",
      "due_at": "2026-09-10T09:00:00+07:00",
      "status": "upcoming"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 50,
    "total": 120,
    "total_pages": 3
  }
}
```

### Notifications

```http
GET    /api/v1/notifications?page=1&per_page=20&category=vehicle&range=30d&status=sent
GET    /api/v1/notifications/{id}
GET    /api/v1/notifications/stats
GET    /api/v1/notifications/export?format=pdf&range=30d
```

### Family

```http
GET    /api/v1/family/members
POST   /api/v1/family/members
GET    /api/v1/family/members/{id}
PATCH  /api/v1/family/members/{id}
DELETE /api/v1/family/members/{id}
GET    /api/v1/family/reminders?page=1&per_page=20
```

### AI Chat

```http
GET    /api/v1/chat/messages?page=1&per_page=30
POST   /api/v1/chat/messages
DELETE /api/v1/chat/messages
GET    /api/v1/chat/usage
GET    /api/v1/chat/intents
```

### Account

```http
GET    /api/v1/account/profile
PATCH  /api/v1/account/profile
POST   /api/v1/account/avatar
GET    /api/v1/account/export
DELETE /api/v1/account
```

### Billing

```http
GET  /api/v1/billing/plan
POST /api/v1/billing/checkout
POST /api/v1/billing/webhook
```

## 8.3 WhatsApp Webhook API

```http
POST /webhooks/whatsapp/{provider}
```

Webhook requirements:

- Verify provider signature.
- Store inbound message before processing.
- Deduplicate by provider message id.
- Return quickly to provider.
- Process intent asynchronously through queue or worker.

## 8.4 Pagination Standard

All list endpoints use offset-based pagination with the following query parameters:

| Parameter | Type | Default | Description |
| --- | --- | --- | --- |
| page | int | 1 | Page number, 1-indexed |
| per_page | int | 20 | Items per page, max 100 |
| sort | string | created_at | Field to sort by |
| order | string | desc | Sort direction: asc or desc |

All list responses include a `pagination` object:

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 87,
    "total_pages": 5
  }
}
```

Rules:

- `per_page` is capped at 100. Requests above this silently default to 100.
- `page` below 1 returns a 400 error.
- Responses never return all records without pagination.
- Cursor-based pagination can be introduced later for high-volume endpoints such as timeline and notification history.

## 8.5 Rate Limiting

Rate limits are enforced per IP address and per authenticated user using Redis sliding window counters.

### Public and Unauthenticated Endpoints

| Endpoint | Limit |
| --- | --- |
| POST /api/v1/auth/login | 10 requests / 15 min per IP |
| POST /api/v1/auth/register | 5 requests / hour per IP |
| POST /api/v1/auth/otp/request | 5 requests / 15 min per phone number |
| POST /api/v1/auth/password/reset/request | 5 requests / hour per email |
| POST /webhooks/whatsapp/{provider} | 200 requests / min per IP |
| All other public GET | 60 requests / min per IP |

### Authenticated API Endpoints

| Endpoint | Limit |
| --- | --- |
| Default per-user | 120 requests / min per user |
| POST /api/v1/reminders | 30 requests / min per user |
| POST /api/v1/documents | 20 requests / min per user |
| POST /api/v1/goals/{id}/progress | 30 requests / min per user |
| GET /api/v1/timeline | 30 requests / min per user |

### Rate Limit Response

When a rate limit is exceeded, respond with HTTP 429:

```json
{
  "error": "rate_limit_exceeded",
  "message": "Too many requests. Please try again later.",
  "retry_after": 60
}
```

Include the following response headers on all rate-limited endpoints:

```
X-RateLimit-Limit: 120
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1750001000
Retry-After: 60
```

Implementation:

- Use Redis sliding window (not fixed window) to avoid burst at window boundary.
- Rate limit keys expire automatically using Redis TTL.
- Log rate limit violations with user ID or IP for monitoring.
- Provide an admin override for trusted internal callers.

## 8.6 CORS Configuration

Cross-Origin Resource Sharing must be explicitly configured. Do not use wildcard origins in production.

### Allowed Origins

- Production: `https://woles.id` and `https://www.woles.id` only.
- Development: `http://localhost:3000`.
- Staging: the staging domain only.

CORS configuration:

```
Access-Control-Allow-Origin: https://woles.id
Access-Control-Allow-Methods: GET, POST, PATCH, DELETE, OPTIONS
Access-Control-Allow-Headers: Authorization, Content-Type, X-CSRF-Token, X-Request-ID
Access-Control-Allow-Credentials: true
Access-Control-Max-Age: 86400
```

Rules:

- `Access-Control-Allow-Credentials: true` is required because refresh tokens are sent via cookie.
- Preflight OPTIONS requests must be responded to with HTTP 204 within 50ms.
- Never set `Access-Control-Allow-Origin: *` on any endpoint that accepts credentials or sets cookies.
- Webhook endpoints (`/webhooks/`) do not need CORS headers because they are called server-to-server.

## 8.7 API Documentation

Maintain a full OpenAPI 3.1 specification for all `/api/v1/` endpoints. This is the single source of truth for the Postman collection and frontend type generation.

### Specification Location

```text
docs/
  openapi.yaml        ← main OpenAPI 3.1 spec
  postman/
    woles.postman_collection.json   ← generated from OpenAPI spec
    woles.postman_environment.json  ← base URL, tokens
```

### OpenAPI Structure

```yaml
openapi: "3.1.0"
info:
  title: Woles API
  version: "1.0.0"
servers:
  - url: https://api.woles.id/api/v1
    description: Production
  - url: http://localhost:8080/api/v1
    description: Local Development
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
  schemas:
    PaginationMeta:
      type: object
      properties:
        page: { type: integer }
        per_page: { type: integer }
        total: { type: integer }
        total_pages: { type: integer }
    ErrorResponse:
      type: object
      properties:
        error: { type: string }
        message: { type: string }
security:
  - bearerAuth: []
```

### Documentation Standards

- Every endpoint must have a `summary`, `description`, request body schema, and all response codes documented (200, 400, 401, 403, 404, 422, 429, 500).
- All request and response bodies must reference reusable `$ref` schemas.
- Authentication endpoints must be tagged with `auth` and must not require `bearerAuth` security.
- Webhook endpoints must document the expected provider payload shape and required signature header.
- Export and commit a regenerated Postman collection whenever the spec changes.

### Postman Environment Variables

```json
{
  "name": "Woles - Local",
  "values": [
    { "key": "base_url", "value": "http://localhost:8080/api/v1" },
    { "key": "access_token", "value": "" },
    { "key": "refresh_token", "value": "" },
    { "key": "user_id", "value": "" }
  ]
}
```

Use Postman pre-request scripts on the `/auth/refresh` request to automatically renew the access token when it is expired.



## 9.1 WhatsApp Message to Reminder

```text
1. User sends WhatsApp message.
2. WhatsApp provider calls Woles webhook.
3. Backend verifies signature.
4. Backend stores inbound message.
5. Backend resolves or creates user by phone number.
6. Backend publishes whatsapp.message_received event.
7. Intent worker extracts structured intent.
8. Application service validates the command.
9. Reminder service creates reminder in PostgreSQL.
10. Scheduler creates next reminder occurrence and notification record.
11. Bot sends confirmation through WhatsApp.
```

## 9.2 Notification Delivery

```text
1. Scheduler finds due notifications.
2. Scheduler claims notifications using row lock or workflow ownership.
3. Scheduler publishes notification.send_requested.
4. Notification worker sends message through WhatsApp provider.
5. Provider returns message id.
6. Worker updates notification status to sent.
7. Provider delivery callback updates delivery status if available.
8. Failed sends are retried with backoff.
9. Permanent failures move to dead-letter handling.
```

## 9.3 Timeline Query

```text
1. User asks WhatsApp or opens web timeline.
2. Timeline service queries upcoming reminders, documents, subscriptions, and goals.
3. Timeline service normalizes records into timeline items.
4. Items are sorted by due date.
5. Response is formatted for WhatsApp or web.
```

## 10. Scheduling Strategy

## 10.1 MVP Scheduling

Use PostgreSQL as the durable schedule store.

Process:

- Store notification rows with `scheduled_at`.
- Run scheduler worker every minute.
- Worker claims due rows using `FOR UPDATE SKIP LOCKED`.
- Worker publishes send requests to RabbitMQ.
- Worker marks claimed rows as `sending`.
- Notification worker updates status after provider response.

Benefits:

- Simple.
- Durable.
- Easy to inspect.
- No reliance on long delayed queue messages.

## 10.2 Recurrence Calculation

For recurring reminders:

- Store recurrence definition in `reminders.recurrence_rule`.
- Store the next due item as a notification or occurrence.
- After successful send or completion, calculate the next occurrence.
- Keep occurrence history for audit and analytics.

## 10.3 Scale Scheduling

When volume grows, move long-running and recurring reminder workflows to Temporal.

Recommended split:

- Temporal: durable reminder and document expiry workflows.
- RabbitMQ: event fan-out and async integration messages.
- PostgreSQL: source of truth and queryable history.

## 11. Idempotency and Reliability

Idempotency keys should be used for:

- WhatsApp inbound webhook processing.
- Reminder creation from repeated messages.
- Notification sends.
- Payment webhooks.

Example notification idempotency key:

```text
notification:{notification_id}:channel:whatsapp
```

Reliability rules:

- Every inbound message is stored before processing.
- Every notification has one durable row.
- Every send attempt is recorded.
- Failed messages can be retried safely.
- Dead-letter records are visible in an admin view.
- Duplicate provider webhooks must not duplicate user records or notifications.

## 12. Security and Privacy

Woles stores sensitive life-admin data. Security must be part of the MVP, not a later addition.

## 12.1 Transport and Secrets

- Use HTTPS everywhere. TLS 1.2 minimum; TLS 1.3 preferred.
- Store secrets in environment variables or a dedicated secret manager (HashiCorp Vault, AWS Secrets Manager, or similar). Never hardcode secrets or commit them to the repository.
- Encrypt sensitive provider tokens at rest using AES-256-GCM before storing in PostgreSQL.
- Use least-privilege database users in production. The application user must not have `CREATE`, `DROP`, or `TRUNCATE` privileges.

## 12.2 Authentication Security

- Passwords hashed with argon2id. See Section 8.1 for parameters.
- Account lockout after 5 consecutive failed login attempts. Return a generic error message that does not reveal whether the account or email exists.
- OTP and magic links are single-use and short-lived (5 minutes for OTP, 15 minutes for magic link).
- JWT access tokens expire in 15 minutes. Refresh tokens expire in 30 days and rotate on every use.
- Invalidate all refresh tokens on password change, account suspension, or detected token reuse.
- On the `/auth/login` and `/auth/register` endpoints, use the same response time regardless of whether the user exists (constant-time response) to prevent user enumeration.

## 12.3 Authorization

- All authenticated endpoints enforce ownership checks. A user may only access their own reminders, documents, subscriptions, goals, and notifications. Return HTTP 404 (not 403) when accessing another user's resource to avoid resource enumeration.
- Plan limits are enforced in application services, not just the frontend.
- Admin endpoints are protected by a separate admin role claim in the JWT and restricted to internal network access.
- Internal service-to-service calls use a separate bearer API key or mTLS. User JWTs are never used for service-to-service authorization.

## 12.4 CSRF Protection

CSRF attacks are mitigated through two complementary mechanisms:

**Double Submit Cookie pattern for state-changing API calls:**

- On page load, the backend issues a `csrf_token` value as a `SameSite=Strict; Secure; HttpOnly=false` cookie so JavaScript can read it.
- The frontend includes this value in an `X-CSRF-Token` header on every POST, PATCH, DELETE request.
- The backend middleware validates that the header value matches the cookie value before processing.

**SameSite cookie attribute:**

- The refresh token cookie is set with `SameSite=Strict; Secure; HttpOnly=true`.
- `SameSite=Strict` prevents the cookie from being sent on cross-site requests, blocking the most common CSRF vector.

Rules:

- Webhook endpoints (`/webhooks/`) are exempt from CSRF checks because they are called by external providers using signature verification instead.
- GET endpoints do not require CSRF tokens.

## 12.5 XSS Protection

XSS risks are reduced through defence-in-depth:

**Backend:**

- Never reflect raw user input directly in API responses without sanitization.
- Sanitize any field that might contain HTML before storing (title, notes, original_text). Strip all HTML tags from user-supplied strings.
- Store and return data as structured JSON, never as HTML fragments.
- Set `Content-Type: application/json` on all API responses.

**HTTP Security Headers (set on every response):**

```
Content-Security-Policy: default-src 'self'; script-src 'self'; object-src 'none'; frame-ancestors 'none'
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
Strict-Transport-Security: max-age=63072000; includeSubDomains; preload
```

**Frontend (Next.js):**

- Use Next.js built-in output escaping. Do not use `dangerouslySetInnerHTML` with user-supplied content.
- Apply Content Security Policy via Next.js `headers()` configuration.
- Validate and sanitize any user input rendered in the UI.

## 12.6 Input Validation

- Validate all request inputs at the HTTP adapter layer before passing to application services.
- Use strict JSON schema validation. Reject unknown fields.
- Enforce maximum lengths on all string fields (title: 200 chars, notes: 2000 chars, original_text: 4000 chars).
- Validate UUIDs, dates, and numeric ranges. Return HTTP 422 with a structured error on validation failures.
- Never pass raw user input into SQL queries. Use parameterized queries only. Prabogo's PostgreSQL adapter handles this.

## 12.7 Logging and Privacy

- Log request IDs, user IDs, and event types. Do not log request bodies that contain passwords, tokens, or sensitive fields.
- Mask phone numbers and emails in application logs. Store only partial values (e.g., `+62812****5678`).
- Do not log full WhatsApp message content in unprotected log stores. Log intent type and confidence score instead.
- Store provider webhook payloads in a dedicated table for debugging, but do not expose them in public admin views.

## 12.8 Audit Log

Maintain an `audit_logs` table for sensitive operations:

| Field | Type | Notes |
| --- | --- | --- |
| id | uuid | Primary key |
| user_id | uuid | FK users.id; nullable for system events |
| actor_type | varchar | user, system, admin |
| action | varchar | login, logout, password_change, reminder_deleted, plan_changed, etc. |
| entity_type | varchar | user, reminder, document, subscription, goal |
| entity_id | uuid | Nullable |
| ip_address | varchar | Masked or hashed in production |
| user_agent | varchar | Truncated |
| created_at | timestamp | |

Audited events include: login, logout, failed login, password change, account deletion, plan upgrade or downgrade, reminder or document deletion, billing events.

## 12.9 WhatsApp Webhook Security

- Verify the provider HMAC-SHA256 signature on every inbound webhook before processing.
- Return HTTP 200 to the provider immediately after signature verification. Process the message asynchronously.
- Reject any webhook with an invalid or missing signature with HTTP 401.
- Deduplicate by `provider_message_id` using a Redis idempotency key before writing to the database.

## 12.10 Compliance

- Prepare a privacy policy and terms of service before public launch.
- Make it clear that Woles is not a government service.
- Provide a data export endpoint for users (`GET /api/v1/account/export`).
- Provide an account deletion endpoint (`DELETE /api/v1/account`) that anonymizes or deletes all personal data within 30 days.
- Apply rate limits per phone number and per IP address.
- Conduct a basic security review before opening to paying users.

## 13. Plan Limits and Billing

Plan enforcement should happen in application services, not only frontend.

Free plan limits:

- 20 reminders.
- 5 documents.
- 5 subscriptions.

Premium:

- Unlimited reminders.
- Unlimited documents.
- Unlimited subscriptions.
- Goal tracker.
- Life Timeline.

Advanced:

- Family account.
- AI assistant.
- OCR.

Recommended payment providers for Indonesia:

- Xendit.
- Midtrans.

Billing webhook requirements:

- Verify signature.
- Store event payload.
- Process idempotently.
- Update plan state only after successful payment confirmation.

## 14. Frontend System Design

## 14.1 Next.js App Structure

The app uses Next.js App Router with route groups. The navigation from the UI mockups defines the actual page set.

```text
app/
  (public)/
    page.tsx                        ← Landing page (Indonesian, WhatsApp CTA)
    reminder-pajak-mobil/
    reminder-stnk/
    reminder-sim/
    reminder-servis-mobil/
    reminder-passport/
    whatsapp-reminder/
  (auth)/
    login/
    verify/
    register/
  (dashboard)/
    layout.tsx                      ← Sidebar nav + mobile bottom nav
    dashboard/
    reminders/
      page.tsx                      ← Card grid with status tabs
      [id]/
    documents/
      page.tsx                      ← Document Vault with file grid
      [id]/
    finances/
      layout.tsx                    ← Shared Finances layout with sub-nav tabs
      goals/
        page.tsx                    ← Financial Goal Tracker
        [id]/
      overview/
        page.tsx                    ← Financial Overview (subscriptions + spending)
    notifications/
      page.tsx                      ← Notification History
    family/
      page.tsx                      ← Family Management (Advanced plan)
    chat/
      page.tsx                      ← AI Chat Hub
    settings/
      page.tsx                      ← Account, Security, Connectivity tabs
components/
  ui/                               ← Primitive components (button, card, badge)
  layout/
    sidebar.tsx
    mobile-nav.tsx
    topbar.tsx
  reminders/
  documents/
  finances/
    goals/
    overview/
  family/
  chat/
lib/
  api/                              ← Typed API client functions
  auth/                             ← Auth helpers and token management
  hooks/                            ← Custom React hooks
server/
  actions/                          ← Next.js Server Actions (form mutations)
styles/
  globals.css
  tokens.css                        ← CSS custom properties from design tokens
public/
```

## 14.2 Design System

The design system is defined from the UI mockup token files. These values should be implemented as CSS custom properties in `tokens.css` and mapped to Tailwind config.

### Color Tokens

| Token | Value | Usage |
| --- | --- | --- |
| `--color-primary` | `#003527` | Primary buttons, active nav, key CTAs |
| `--color-primary-container` | `#064e3b` | Upgrade/Pro cards, dark CTAs |
| `--color-on-primary` | `#ffffff` | Text on primary backgrounds |
| `--color-on-primary-container` | `#80bea6` | Muted text on dark cards |
| `--color-inverse-primary` | `#95d3ba` | Light accent on dark surface |
| `--color-surface` | `#f8faf6` | Page background |
| `--color-surface-container-lowest` | `#ffffff` | Card backgrounds |
| `--color-surface-container-low` | `#f2f4f1` | Subtle section backgrounds |
| `--color-surface-container` | `#eceeeb` | Input backgrounds, chips |
| `--color-on-surface` | `#191c1b` | Primary body text |
| `--color-on-surface-variant` | `#404944` | Secondary text, labels |
| `--color-outline` | `#707974` | Borders, dividers |
| `--color-outline-variant` | `#bfc9c3` | Subtle borders |
| `--color-error` | `#ba1a1a` | Error states, danger zone |
| `--color-error-container` | `#ffdad6` | Error backgrounds |
| `--color-secondary` | `#0058be` | Links, info states |

### Typography Tokens

| Token | Font | Size | Weight | Usage |
| --- | --- | --- | --- | --- |
| `display-lg` | Geist | 48px | 700 | Hero headlines (landing page) |
| `headline-lg` | Geist | 32px | 600 | Page titles (desktop) |
| `headline-lg-mobile` | Geist | 28px | 600 | Page titles (mobile) |
| `headline-md` | Geist | 24px | 600 | Section headers |
| `title-lg` | Inter | 20px | 600 | Card titles |
| `body-lg` | Inter | 18px | 400 | Body text |
| `body-md` | Inter | 16px | 400 | Default text |
| `label-md` | Inter | 14px | 500 | Labels, badges, nav items |
| `label-sm` | Inter | 12px | 500 | Meta text, timestamps |

Fonts: load `Geist` and `Inter` from `next/font/google`. Never load both at full weight range; only include the weights used.

### Border Radius

| Token | Value | Usage |
| --- | --- | --- |
| `--radius-sm` | 4px | Badges, chips |
| `--radius-md` | 8px | Buttons, inputs |
| `--radius-lg` | 12px | Cards |
| `--radius-xl` | 16px | Modal sheets, drawers |
| `--radius-full` | 9999px | Avatar circles, pill buttons |

## 14.3 Navigation

The dashboard layout uses a **left sidebar on desktop** and a **bottom navigation bar on mobile**.

### Sidebar (desktop, ≥ 1024px)

```
[Logo: Woles / Life Admin]

Dashboard
Reminders
Documents
Finances
AI Chat Hub       ← in nav with usage indicator
Family            ← Advanced plan only; hidden on free/premium
Settings

─────────────────
[User avatar, name, plan badge]
[Upgrade Now]     ← visible on free plan
```

Active state: filled dark green background (`--color-primary-container`) on the nav item.

### Bottom Navigation (mobile, < 1024px)

```
Dashboard | Reminders | Documents | Finances | More
```

"More" opens a bottom sheet with Chat, Family, Settings, and account actions.

### Topbar (mobile only)

Shown on mobile above page content: logo left, notification bell and avatar right.

## 14.4 Page Designs

### Dashboard

- **Stats row**: Active Reminders, Documents, Subscriptions (from `/api/v1/reminders?per_page=0`, `/api/v1/documents?per_page=0`, `/api/v1/subscriptions?per_page=0`).
- **Upcoming items**: chronological list of next 5-10 items from `/api/v1/timeline?range=30d&per_page=10`.
- **Quick actions**: New Reminder, New Document, Add Goal.
- **Recent activity**: last 5 events from audit log.
- Responsive: stats row collapses to 2 columns on mobile; upcoming list becomes full-width cards.

### Reminders

- Filter tabs: All Reminders | Upcoming (count) | Overdue (count) | Completed.
- Card grid (3 columns desktop → 2 columns tablet → 1 column mobile).
- Each card shows: category icon, status badge (DUE IN N DAYS / UPCOMING / OVERDUE), title, description, due date, 3-dot action menu.
- "DUE IN N DAYS" badge is shown in amber/red when `next_run_at` is within 3 days.
- FAB (Floating Action Button): `+ New Reminder`.
- API: `GET /api/v1/reminders?status={tab}&sort=next_run_at&order=asc`.

### Document Vault

- Topbar: page title, global search, notification bell, avatar.
- Stats row: `Expiring Soon` (count from documents where `expiry_date <= now + 30d`), `Security Status` (always Protected), `Cloud Storage` (from `/api/v1/documents/storage/usage`).
- Filter tabs: All Files | Vehicles | Identity | Insurance | Financials.
- Document grid (3 columns desktop → 2 columns tablet → 1 column mobile): thumbnail image or category icon, badge (VEHICLE / IDENTITY / INSURANCE), title, last updated, expiry status (`Expires in N days`, `Valid until YYYY`).
- Empty slot card: `Add New Document` with dashed border.
- Vault Health widget: completeness progress bar (from `/api/v1/documents/vault/health`).
- Recent Activity section (from audit log filtered by entity_type=document).
- Woles Tip card: contextual tip based on missing document types.
- File upload: clicking Upload New opens a modal with `multipart/form-data` POST.
- API: `GET /api/v1/documents?vault_category={tab}`.

### Finances (Goal Tracker)

URL: `/finances/goals`

- Page title: **Financial Goal Tracker** — subtitle: "Tracking your calm journey to financial freedom."
- Top-right CTA: `Update Progress` button → opens a drawer to record a new savings amount for the active goal.
- **FAB** bottom-left: `+ New Task` (adds a new goal).
- **Active Goal hero card** (left, wide):
  - `ACTIVE GOAL` badge (teal chip).
  - Goal name (e.g., Dana Nikah), Target amount (e.g., Rp 200.000.000).
  - Horizontal progress bar with percentage complete (e.g., 62.5% Completed).
  - Stats row (3 mini-cards): `Current Savings` | `Remaining` | `Monthly Target`.
  - Source: `GET /api/v1/goals?status=active&per_page=1` (first active goal displayed as hero).
- **Woles Pro promo card** (right, fixed width, dark green background):
  - Title: "Advanced Goal Analytics".
  - Description: "Predict exactly when you'll reach your goal based on spending patterns."
  - CTA: `Unlock Insights` button (white, opens billing/upgrade flow).
  - Visible on free and premium plans. Hidden on Advanced plan.
- **All Financial Goals** section (below hero, left column):
  - Header: "All Financial Goals" + `View History >` link → `GET /api/v1/goals/history`.
  - List of goals: goal icon (category color), goal name, progress bar (colour-coded: emerald for high progress, blue for mid, amber for low), percentage label.
  - Clicking a row expands inline or navigates to `/finances/goals/{id}`.
  - Source: `GET /api/v1/goals`.
- **Trajectory Outlook** widget (below hero, right column):
  - `PRO` badge in header.
  - Locked state (padlock icon, blurred chart): "Predictive Growth — Unlock the ability to see future projections based on inflation and savings rates."
  - CTA: `Upgrade to Pro` button.
  - This feature requires the Advanced plan and backend data from `/api/v1/finances/trend`.
- **Woles Finance Tip** card (full-width, bottom):
  - Light background card with a lightbulb icon.
  - Tip text is contextual based on the active goal's progress and `monthly_target` vs. current trajectory.
  - Backend generates tip text at `/api/v1/goals?status=active&include_tip=true`.
- **Responsive:**
  - Hero card + Pro promo card: side-by-side on desktop → stacked on mobile (hero first, promo below).
  - All Goals list + Trajectory Outlook: side-by-side on desktop (left 60% / right 40%) → stacked on mobile.
  - Finance Tip: full-width on all breakpoints.

### Finances (Financial Overview)

URL: `/finances/overview`

- Page title: **Financial Overview** — subtitle: "Manage your spending and monthly commitments effortlessly."
- Topbar (on this page only): global search input `Search finances...`, notification bell, avatar.
- Top-right actions: `Export Data` button → `GET /api/v1/finances/export?format=csv` and `Add Transaction` button → opens a form drawer to record a manual expense entry.
- **Budget Status** (sidebar widget, visible on desktop):
  - Small progress bar: "N% of monthly goal".
  - Source: `monthly_target` from the active goal vs. current total expenses.
- **Total Expenses hero card** (dark green, left):
  - Label: `TOTAL EXPENSES • THIS MONTH`.
  - Amount: formatted as Rp XX.XXX.XXX.
  - Change indicator: ↓ N% less / ↑ N% more than last month.
  - Sub-stats: Income (manual entry or V2 bank sync) and Savings.
  - Source: `GET /api/v1/finances/summary?period=monthly`.
- **Linked Accounts card** (right):
  - Lists connected bank accounts with masked card numbers.
  - `+ Connect Bank` CTA (V2 feature — Plaid or local open-banking).
  - **MVP**: Show empty state with `Connect Bank (Coming Soon)` placeholder. Do not block the page.
- **Category Breakdown card** (left, below hero):
  - Period selector: `Current Month` dropdown.
  - 2×2 grid of category tiles: each shows category icon, name, percentage, and IDR amount.
  - Categories: Household, Utilities, Transport, Others (derived from subscription categories + manual transactions).
  - Source: `GET /api/v1/finances/spending?period=monthly`.
- **Spending Trend card** (right, below hero):
  - Bar chart: weekly spending comparison for the current month.
  - Source: `GET /api/v1/finances/trend?period=weekly`.
- **Upcoming Bills table** (full-width, bottom):
  - Header: "Upcoming Bills" with count badge ("You have N payments due this week").
  - `View Calendar` link (V2).
  - Table columns: SERVICE (icon + name + billing type) | DUE DATE | AMOUNT | STATUS | ACTION.
  - STATUS badges: `PENDING` (amber), `URGENT` (red), `SCHEDULED` (grey).
  - ACTION icons: pay/mark-paid, view receipt, view detail.
  - Source: `GET /api/v1/finances/upcoming-bills` (combines subscriptions with `next_billing_at` within 30 days).
  - On mobile: table becomes stacked cards. Each card shows service name, due date, amount, and status badge.
- **Responsive:**
  - Hero + Linked Accounts: side-by-side on desktop → stacked on mobile.
  - Category Breakdown + Spending Trend: side-by-side on desktop → stacked on mobile.
  - Upcoming Bills table: horizontal scroll on tablet; card list on mobile (< 640px).

**Finance sub-navigation:**

Both Finances pages share a secondary tab bar directly below the page topbar:

```
[ Goal Tracker ]  [ Financial Overview ]
```

Active tab uses a bottom underline in primary color. This tab bar is rendered inside `finances/layout.tsx`.

### Notification History

- Page title: Notification History.
- Filter by category: All | Vehicle | Bills | Documents | Property.
- Date range selector: Last 7 Days | Last 30 Days | Last 90 Days | Custom.
- Table: Status | Date & Time | Category | Message Content | Actions (View Message / View Receipt / Renew Now).
- Status values: Sent, Read, Action Taken, Failed.
- Pagination: standard offset pagination matching Section 8.4.
- Stats cards below table: Delivery Rate (%), Top Category.
- Export PDF button → `GET /api/v1/notifications/export?format=pdf`.
- API: `GET /api/v1/notifications?category={filter}&range={range}`.

### Family Management (Advanced plan)

- Family Members grid (4 columns desktop → 2 columns tablet → 1 column mobile).
- Each member card: avatar, role badge (Primary / Spouse / Parent / Child), name, relation label, active reminder count, View Details button, Access management, delete button.
- Shared Reminders table: Task Name | Owner (avatar + name) | Due Date | Status badge | Action.
- Status badges: Urgent (red) | Upcoming (amber) | Safe (green).
- "Add Family Member" CTA → opens a form drawer.
- Plan gate: show upgrade prompt if user is not on Advanced plan.
- API: `GET /api/v1/family/members`, `GET /api/v1/family/reminders`.

### AI Chat Hub

- Three-panel layout on desktop: sidebar nav | message list | chat area.
- Message list panel: shows recent conversation threads.
- Intents panel (in message list, below): groups detected intents from recent chats (e.g., Saving Goals → Dana Nikah 200jt).
- Chat area: message bubbles (user right, assistant left), timestamps.
- Input: text field, attachment icon, Voice Command button (V2), Suggest Intent button.
- Usage monitor in sidebar: "N of 10 free messages used" progress bar; "Upgrade to Pro" CTA.
- Free plan quota notice banner at top: "You are on the Free Plan. N messages remaining this month. Go Unlimited."
- On mobile: shows only chat area; message list accessible via back navigation.
- API: `GET /api/v1/chat/messages`, `POST /api/v1/chat/messages`, `GET /api/v1/chat/usage`.

### Settings

Three sub-sections accessible via a secondary sidebar or tabs:

**Account Details:**

- Avatar upload (PATCH `/api/v1/account/avatar`).
- First Name, Last Name (split from `name` field), Email, plan badge.
- Save Changes button.

**Connectivity Integrations:**

- Google Workspace toggle (V2).
- Financial Plaid toggle (V2).
- Smart Home Hub toggle (V2).
- Each shows Connected / Disconnected status.

**Security & Privacy:**

- Password: last changed date + Update Password button.
- Two-Factor Authentication: toggle + RECOMMENDED badge.
- Active Sessions: list of devices with last active time; "View Sessions" / revoke per session.
- Danger Zone: Delete Account button (red outlined, requires confirmation modal).

### Landing Page

- Full Indonesian copy.
- Hero section: headline ("Semua urusan hidup, satu chat aja"), WhatsApp CTA, live AI assistant preview card.
- Problem section: "Lupa bayar tagihan? STNK mati?" with pain points.
- How it works: 3-step visual (Send via WA → Woles tracks → Get reminded).
- Categories grid: STNK, SIM, Pajak, Servis, Passport, Tagihan, Langganan, Cicilan.
- Trust section: Data Privacy, WhatsApp-First, Reliable & Human.
- FAQ accordion.
- CTA section: "Siap untuk hidup lebih santai?"
- Footer: links, social, legal.
- Server-rendered for SEO. FAQ uses JSON-LD schema markup.

## 14.5 Responsive Design

All pages must be fully responsive across three breakpoints:

| Breakpoint | Range | Layout |
| --- | --- | --- |
| Mobile | < 640px | Single-column, bottom nav, stacked cards |
| Tablet | 640px – 1023px | 2-column grids, sidebar collapses to top nav |
| Desktop | ≥ 1024px | 3-column grids, fixed sidebar |

### Responsive Rules

- **Sidebar**: fixed at 240px on desktop (≥ 1024px). Hidden on mobile; replaced by bottom nav.
- **Card grids**: `grid-cols-1 sm:grid-cols-2 lg:grid-cols-3` for Reminders and Documents.
- **Family Members grid**: `grid-cols-1 sm:grid-cols-2 xl:grid-cols-4`.
- **Tables** (Notification History, Family Reminders): become a card list on mobile. Each row renders as a stacked card.
- **Topbar**: visible only on mobile (< 1024px). Contains logo, notification bell, avatar.
- **Typography**: `headline-lg-mobile` (28px) replaces `headline-lg` (32px) on mobile.
- **Stats rows**: `grid-cols-1 sm:grid-cols-2 lg:grid-cols-3`.
- **AI Chat Hub**: three-panel on desktop. Single-panel chat on mobile with back navigation.
- **Settings**: horizontal tabs on desktop; stacked vertical sections on mobile.
- **Touch targets**: all interactive elements must be ≥ 44×44px on mobile.
- **Modals and drawers**: use full-screen bottom sheet on mobile; centered modal on desktop.

### Tailwind Configuration

```js
// tailwind.config.js
theme: {
  extend: {
    colors: {
      primary: '#003527',
      'primary-container': '#064e3b',
      'on-primary': '#ffffff',
      surface: '#f8faf6',
      'surface-container': '#eceeeb',
      'on-surface': '#191c1b',
      'on-surface-variant': '#404944',
      error: '#ba1a1a',
    },
    fontFamily: {
      display: ['Geist', 'sans-serif'],
      body: ['Inter', 'sans-serif'],
    },
    borderRadius: {
      sm: '4px',
      DEFAULT: '8px',
      md: '12px',
      lg: '16px',
      full: '9999px',
    },
  },
}
```

## 14.6 SEO Pages

Use static or server-rendered pages for SEO.

Each SEO page should include:

- Specific keyword-focused H1.
- Problem explanation.
- WhatsApp command example.
- Benefits.
- FAQ schema.
- CTA to start via WhatsApp.

Important pages:

- Reminder pajak mobil.
- Reminder STNK.
- Reminder SIM.
- Reminder servis mobil.
- Reminder passport.
- Reminder subscription.
- WhatsApp reminder Indonesia.

## 15. Observability

Required logs:

- Inbound WhatsApp message received.
- Intent extraction result.
- Reminder created.
- Notification scheduled.
- Notification send attempted.
- Notification sent or failed.
- Payment webhook received.

Metrics:

- Webhook latency.
- Intent extraction success rate.
- Notification due count.
- Notification sent count.
- Notification failure rate.
- Queue depth.
- Scheduler lag.
- API error rate.
- WhatsApp provider error rate.

Alerts:

- Notifications stuck in `scheduled` past due time.
- Queue depth above threshold.
- WhatsApp provider failure spike.
- Scheduler worker down.
- Database connection saturation.
- Payment webhook failures.

Recommended tools:

- OpenTelemetry for traces.
- Prometheus and Grafana for metrics.
- Loki or another structured log store.
- Sentry for frontend and backend errors.

## 16. Deployment Architecture

## 16.1 MVP Deployment

Simple VPS deployment:

```text
VPS
  - Backend API container
  - Worker container
  - Scheduler container
  - PostgreSQL container or managed Postgres
  - Redis container
  - RabbitMQ container
  - Next.js app container
  - Nginx or Caddy reverse proxy
  - MinIO container (S3-compatible object storage for document file uploads)
```

Recommended early provider:

- Hetzner CX22 or larger.

Important note:

For public launch, managed PostgreSQL is preferable once revenue or user trust justifies it. Self-hosting is cheaper early, but backups and recovery must be taken seriously.

## 16.2 Production Deployment

Scale path:

- Separate database from app servers.
- Use managed PostgreSQL or dedicated database VM.
- Run multiple backend API replicas.
- Run multiple notification workers.
- Add Temporal for durable workflows.
- Use managed object storage (AWS S3, Cloudflare R2, or Backblaze B2) for document file uploads.
- Add CDN for public SEO pages and document file signed URLs.

## 17. Backup and Disaster Recovery

Requirements:

- Daily PostgreSQL backups.
- Point-in-time recovery when possible.
- Backup restore test before public launch.
- Redis does not need to be primary backup source.
- Queue messages should be durable.
- Store provider webhook payloads for debugging.

Recovery objectives for MVP:

- RPO: less than 24 hours.
- RTO: less than 4 hours.

For public paid product:

- RPO: less than 1 hour.
- RTO: less than 1 hour.

## 18. Testing Strategy

## 18.1 Backend Tests

Required:

- Unit tests for recurrence calculation.
- Unit tests for document reminder offsets.
- Unit tests for subscription billing cycle calculation.
- Unit tests for timeline aggregation.
- Unit tests for plan limit enforcement.
- Idempotency tests for notification delivery.

Security tests required:

- argon2id hash and verify correctness.
- Account lockout triggers correctly after 5 failures.
- JWT access token validation: expired token, wrong signature, tampered payload, `none` algorithm rejection.
- Refresh token rotation: old token rejected after rotation, reuse triggers family revocation.
- CSRF token validation: missing header rejected, mismatched header rejected.
- Ownership checks: user A cannot access user B's resources (expect 404).
- Plan limit enforcement: requests beyond free plan limits are rejected with 403.
- Rate limit enforcement: 10th+ login attempt within 15 minutes is rejected with 429.
- Input validation: strings exceeding max length return 422, SQL injection payloads are harmless, unknown JSON fields are rejected.
- WhatsApp webhook signature: invalid signature returns 401.

## 18.2 Integration Tests

Required:

- PostgreSQL repository tests.
- WhatsApp webhook verification tests.
- RabbitMQ message publishing and consuming tests.
- Notification worker retry tests.
- Refresh token store: issue, rotate, revoke, detect reuse.
- CSRF middleware integration test.
- Rate limiter Redis integration test.
- Audit log writes on sensitive operations.

## 18.3 End-to-End Tests

Critical flows:

- User sends WhatsApp reminder command.
- Reminder appears in dashboard.
- Notification is scheduled.
- Notification worker sends message.
- Timeline shows upcoming item.

## 19. Development Milestones

## 19.1 Milestone 1: Foundation

- Prabogo project setup.
- PostgreSQL migrations.
- Basic user identity.
- WhatsApp webhook skeleton.
- Next.js dashboard shell.

## 19.2 Milestone 2: Reminder MVP

- Create reminder from WhatsApp.
- Create reminder from web.
- Recurrence calculation.
- Notification scheduling.
- Notification delivery worker.

## 19.3 Milestone 3: Documents and Subscriptions

- Document expiry tracking.
- Default document reminders.
- Subscription tracker.
- WhatsApp command for subscription list.

## 19.4 Milestone 4: Timeline

- Timeline aggregation service.
- Timeline web page.
- WhatsApp timeline query.
- 30 and 90 day filters.

## 19.5 Milestone 5: Public Beta

- Billing integration.
- Plan limit enforcement.
- SEO landing pages.
- Analytics events.
- Admin reliability dashboard.

## 20. Technical Risks and Recommendations

| Risk | Recommendation |
| --- | --- |
| Long-running reminders are hard to manage with only queues | Use PostgreSQL schedule table for MVP, then Temporal for durable workflows |
| WhatsApp provider costs become high | Track notification cost per active user and limit free plan usage |
| AI extraction errors create wrong reminders | Use confirmation and clarification for low-confidence parses |
| Self-hosted database failure | Add automated backups and restore tests early |
| Queue retry creates duplicate notifications | Use notification idempotency keys and unique constraints |
| SEO pages become thin content | Build practical Indonesian guides with real examples and FAQ schema |
| JWT signing key leaked | Use RS256 with asymmetric keys; store private key in secret manager; rotate every 90 days |
| Refresh token stolen via XSS | Use HttpOnly cookie for refresh token; enforce strict CSP; rotate tokens on each use |
| Credential stuffing attack on login | Enforce rate limiting and account lockout per IP and per account |
| CSRF attack on authenticated state-changing request | Use Double Submit Cookie + SameSite=Strict on session cookies |
| Stored XSS via user-supplied content | Sanitize HTML from all user inputs on the server; set nosniff and CSP headers |
| Password breach from database dump | argon2id hashing with per-user salt ensures leaked hashes cannot be reversed at scale |
| User enumeration via login error messages | Return identical response for wrong email and wrong password |
| IDOR: user accesses another user's data | Enforce ownership check at application service layer; return 404 not 403 |

## 21. Final Recommendation

Use this architecture for MVP:

- Backend: Go with Prabogo.
- Backend style: modular monolith using Hexagonal Architecture.
- HTTP server: Fiber through Prabogo adapter.
- Frontend: Next.js with TypeScript.
- Database: PostgreSQL.
- Cache: Redis.
- Queue: RabbitMQ.
- Scheduler: PostgreSQL-backed scheduled notification worker.
- AI: hybrid deterministic parser plus LLM fallback for intent extraction.
- WhatsApp: official WhatsApp Business API provider.

Prepare this architecture for scale:

- Move durable reminder workflows to Temporal when reminder volume and reliability needs grow.
- Keep RabbitMQ for event fan-out.
- Move PostgreSQL to managed or dedicated infrastructure.
- Add observability before paid public launch.

This approach keeps the MVP realistic while preserving a clean path to a larger, reliable public system.
