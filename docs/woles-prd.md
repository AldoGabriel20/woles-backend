# Woles Product Requirements Document

## 1. Product Summary

**Product name:** Woles

**Tagline:** Semua urusan hidup, satu chat aja

**Product type:** WhatsApp-first personal life administration assistant with a companion web dashboard.

Woles helps users remember, track, and manage recurring life responsibilities such as SIM, STNK, vehicle tax, car service, subscriptions, installments, bills, passport renewals, and financial goals. Users can create and query reminders through WhatsApp, receive notifications through WhatsApp, and manage everything from a web dashboard.

The initial product should feel simple, reliable, and calm. The core promise is not an AI chatbot. The core promise is that users no longer forget important personal admin tasks.

## 2. Problem Statement

Many users manage important personal responsibilities across scattered places: calendar apps, WhatsApp chats, notes apps, banking apps, spreadsheets, paper documents, and memory. This creates a high chance of missed renewals, late payments, expired documents, and stress around routine admin tasks.

Examples of fragmented responsibilities:

- SIM expiry
- STNK expiry
- Vehicle tax
- Car service
- Home service schedules
- Internet bills
- Subscriptions
- Installments
- Passport expiry
- Financial goals

Existing tools are either too generic, too manual, or not integrated into the channel users already check every day: WhatsApp.

## 3. Target Users

### Primary Segment

Urban Indonesian adults who manage personal and household responsibilities and already use WhatsApp as their main communication channel.

Examples:

- Young professionals managing subscriptions, vehicle documents, and bills
- Newly married couples managing household tasks and financial goals
- Car or motorcycle owners tracking tax, STNK, service, and insurance
- Frequent travelers tracking passport and visa expiry
- Family administrators who track responsibilities for parents, spouse, or children

### Early Adopter Profile

The first ideal users are people who already search for specific reminder needs such as:

- Reminder pajak mobil
- Reminder STNK
- Reminder SIM
- Reminder servis mobil
- Reminder passport

These users have high intent because they are already trying to solve a specific admin problem.

## 4. Product Goals

### MVP Goals

- Help users create important reminders through natural WhatsApp messages.
- Send reliable WhatsApp reminders before due dates.
- Provide a simple web dashboard for viewing and managing reminders, documents, subscriptions, and timeline items.
- Validate whether users will maintain at least 5 active reminders in Woles.
- Validate retention from real reminder usefulness, not AI novelty.

### Business Goals

- Reach the first 50 active users within 30 days of MVP launch.
- Achieve more than 40% user retention after the first reminder cycle.
- Reach an average of at least 5 active reminders per active user.
- Identify which life-admin category drives the highest recurring usage.
- Validate willingness to pay for unlimited reminders, timeline, and goal tracking.

### Non-Goals for MVP

- Full AI proactive assistant behavior.
- OCR document scanning.
- Family account collaboration.
- Native mobile apps.
- Complex finance planning.
- Deep integrations with banks, government systems, or vehicle service providers.

## 5. Product Positioning

Woles is a personal life admin assistant for Indonesian users who want one reliable place to track recurring life responsibilities.

The product should be positioned around outcomes:

- Never forget STNK, SIM, passport, vehicle tax, subscriptions, or service schedules.
- Add reminders naturally through WhatsApp.
- See what needs attention this month.
- Get notified before things become urgent.

Woles should avoid positioning itself as a generic productivity app or open-ended chatbot. The strongest wedge is concrete life-admin reminders.

## 6. Core User Journeys

### 6.1 WhatsApp Onboarding

**User:** Halo

**Bot:**
Halo Aldo. I can help you manage reminders, important documents, subscriptions, and upcoming life admin tasks. What would you like to save first?

Expected result:

- User starts from WhatsApp without installing an app.
- Bot explains the core value in one message.
- User is guided to create the first useful reminder.

### 6.2 Create Reminder

**User:** ingatkan bayar internet tiap tanggal 10

System extracts:

```json
{
  "intent": "create_reminder",
  "title": "Bayar Internet",
  "category": "bill",
  "recurrence_type": "monthly",
  "day_of_month": 10
}
```

**Bot:** Saved. I will remind you every month on the 10th to pay internet.

Acceptance criteria:

- User can create one-time and recurring reminders from WhatsApp.
- Bot asks a clarification question if the date or recurrence is ambiguous.
- Reminder appears in the web dashboard.
- Reminder is scheduled for future notification delivery.

### 6.3 Create Document Expiry Reminder

**User:** SIM saya habis 15 Mei 2027

**Bot:** Saved. I will remind you 30 days, 7 days, and 1 day before your SIM expires.

Acceptance criteria:

- System stores the document type and expiry date.
- System creates default reminder schedules.
- User can view the document on the web dashboard.
- User can modify reminder timing later.

### 6.4 Track Subscription

**User:** Netflix 120rb per bulan

System extracts:

```json
{
  "intent": "create_subscription",
  "name": "Netflix",
  "amount": 120000,
  "currency": "IDR",
  "billing_cycle": "monthly"
}
```

**User:** langganan saya

**Bot:**
Your active subscriptions:

- Netflix - Rp120,000/month
- ChatGPT - USD20/month

Acceptance criteria:

- User can create a subscription from WhatsApp.
- User can list active subscriptions from WhatsApp.
- Subscription appears in dashboard statistics and timeline.

### 6.5 Track Goal

**User:** target dana nikah 200 juta

**Bot:** Saved. Your goal is Dana Nikah with a target amount of Rp200,000,000.

Acceptance criteria:

- User can save a financial goal.
- User can optionally update current progress later.
- Goal tracking can be premium-gated after MVP validation.

### 6.6 Life Timeline

**User:** apa yang harus saya urus bulan ini?

**Bot:**
This month:

- Netflix
- Pajak Mobil
- Servis Mobil

**User:** apa yang akan jatuh tempo 90 hari ke depan?

**Bot:**
In the next 90 days:

- STNK renewal
- Passport renewal reminder
- Car service

Acceptance criteria:

- Timeline combines reminders, documents, subscriptions, and goals.
- User can query by this month, next month, next 7 days, next 30 days, and next 90 days.
- Timeline is available in WhatsApp and web.

## 7. MVP Feature Scope

### 7.1 Reminder Management

Capabilities:

- Create reminder from WhatsApp.
- Create reminder from web dashboard.
- Support one-time reminders.
- Support recurring reminders: daily, weekly, monthly, yearly, and custom interval such as every 6 months.
- View reminders.
- Edit reminders.
- Delete reminders.
- Mark reminders as done.
- Pause and resume reminders.

Priority: P0

### 7.2 Important Documents

Capabilities:

- Save document type, title, expiry date, and optional notes.
- Default reminders: 30 days, 7 days, and 1 day before expiry.
- Support document categories such as SIM, STNK, passport, vehicle insurance, visa, and custom.
- View documents by expiry date.
- Show soon-to-expire documents in timeline.

Priority: P0

### 7.3 Subscription Tracker

Capabilities:

- Save subscription name, amount, currency, billing cycle, and next billing date.
- View active subscriptions.
- Ask WhatsApp bot for active subscriptions.
- Include subscriptions in monthly timeline.
- Show monthly estimated subscription cost.

Priority: P0

### 7.4 Life Timeline

Capabilities:

- Show all upcoming responsibilities in one chronological view.
- Support monthly view on web.
- Support WhatsApp queries for upcoming tasks.
- Filter by category: reminder, document, subscription, goal.

Priority: P0

### 7.5 Web Dashboard

Capabilities:

- Dashboard overview with upcoming items.
- Statistics: active reminders, documents, subscriptions.
- Quick actions for reminder, document, and subscription.
- Reminder list with search.
- Timeline page.
- Mobile-friendly responsive UI.

Priority: P0

### 7.6 WhatsApp Notification Delivery

Capabilities:

- Send reminders through WhatsApp.
- Record delivery status.
- Retry failed notifications.
- Prevent duplicate notifications.
- Include direct action links where possible, such as mark as done or open dashboard.

Priority: P0

### 7.7 AI Intent Extraction

Capabilities:

- Convert natural language WhatsApp messages into structured actions.
- Ask clarification when needed.
- Keep AI scope limited to extraction and classification in MVP.
- Avoid open-ended free chatbot behavior in MVP to control cost and reliability.

Priority: P1 for MVP, P0 if natural language creation is considered essential.

## 8. V2 Feature Scope

### 8.1 Proactive AI Assistant

Example:

**User:** saya beli mobil baru

**Bot:**
Do you want me to create reminders for:

- Regular service
- Annual vehicle tax
- STNK renewal
- Insurance renewal

V2 behavior:

- Detect life events.
- Suggest relevant reminder templates.
- Let users accept or reject suggestions.
- Learn from user categories and previous behavior.

### 8.2 OCR

Use OCR to extract expiry dates and document information from uploaded images, such as SIM, STNK, passport, or insurance documents.

### 8.3 Family Account

Allow one user to manage reminders for spouse, parents, children, or household members.

### 8.4 Advanced Financial Goals

Add goal progress updates, contribution reminders, and target date estimation.

## 9. User Experience Principles

- WhatsApp should be the fastest input channel.
- The web dashboard should be the clearest management channel.
- Every reminder must feel trustworthy.
- Bot responses should be concise, calm, and useful.
- The UI should be elegant, unique, and public-launch ready, but not decorative at the expense of clarity.
- The product should feel like a quiet personal operations center, not a generic productivity board.

## 10. Web Information Architecture

### 10.1 Dashboard

Sections:

- Welcome message
- Upcoming items
- Key statistics
- Quick action buttons
- Recent activity

Example statistics:

- Active Reminders: 15
- Documents: 7
- Subscriptions: 4

### 10.2 Reminders Page

Features:

- Search reminders.
- Filter by recurrence, category, and status.
- Show next due date.
- Edit, pause, delete, and mark as done.

### 10.3 Documents Page

Features:

- List documents by expiry date.
- Show expiry risk states: safe, upcoming, urgent, expired.
- Add or edit document metadata.

### 10.4 Subscriptions Page

Features:

- List subscriptions.
- Show monthly total.
- Filter by currency and billing cycle.
- Mark canceled subscriptions.

### 10.5 Timeline Page

Features:

- Chronological monthly timeline.
- Upcoming 7, 30, and 90 day views.
- Category filters.
- Item detail drawer.

### 10.6 Settings Page

Features:

- User profile.
- WhatsApp number connection status.
- Notification preferences.
- Plan and billing.
- Data export and account deletion.

## 11. Success Metrics

### Activation Metrics

- Percentage of users who create at least 1 reminder during onboarding.
- Percentage of users who create at least 3 reminders in first 7 days.
- WhatsApp-to-dashboard account connection completion rate.

### Engagement Metrics

- Average active reminders per user.
- Timeline queries per user per month.
- Reminder completion rate.
- Subscription list query usage.
- Monthly active users.

### Retention Metrics

- Day 7 retention.
- Day 30 retention.
- Reminder cycle retention: user returns after receiving a useful reminder.
- Percentage of users with at least 5 active reminders after 30 days.

### Revenue Metrics

- Free-to-premium conversion rate.
- Monthly recurring revenue.
- Average revenue per paying user.
- Churn rate by plan.
- WhatsApp notification cost per active user.

## 12. Subscription Model

### Free Plan - Rp0

Limits:

- 20 reminders
- 5 documents
- 5 subscriptions
- Basic WhatsApp reminders
- Basic web dashboard

Purpose:

- Reduce adoption friction.
- Let users experience the value before paying.
- Encourage users to build enough data that the product becomes useful.

### Premium Plan - Rp39,000/month

Includes:

- Unlimited reminders
- Unlimited documents
- Unlimited subscriptions
- Goal tracker
- Life timeline
- Advanced notification settings
- Priority reminder delivery rules

Purpose:

- Main consumer monetization plan.
- Best fit for individuals who already rely on Woles for daily life admin.

### Advanced Plan - Rp99,000/month

Includes:

- Family account
- AI assistant suggestions
- OCR document extraction
- Shared household timeline
- Multiple WhatsApp numbers, subject to provider policy and cost

Purpose:

- Higher willingness-to-pay users.
- Families and household managers.
- Users with many documents, vehicles, subscriptions, and responsibilities.

## 13. Monetization Strategy

### Primary Revenue

Monthly subscriptions from Premium and Advanced users.

Initial target assumptions:

- 1,000 users with 5% Premium conversion = 50 paying users.
- 50 x Rp39,000 = Rp1,950,000 monthly revenue.
- 10,000 users with 5% Premium conversion = 500 paying users.
- 500 x Rp39,000 = Rp19,500,000 monthly revenue.
- 100,000 users with 5% Premium conversion = 5,000 paying users.
- 5,000 x Rp39,000 = Rp195,000,000 monthly revenue.

### Expansion Revenue

Potential future revenue streams:

- Annual plan with discount.
- Family add-ons.
- OCR packs for heavy document users.
- Business or team plan for small offices that manage asset renewals.
- Partner offers for insurance, vehicle service, travel documents, or financial products.

Partner monetization should only be introduced after trust is established. Woles handles sensitive life-admin data, so aggressive ads would damage the brand.

## 14. Cost Model

### Early Stage Infrastructure

Recommended early setup:

- VPS: Hetzner CX22 or similar, around Rp250,000/month.
- PostgreSQL: self-hosted initially.
- Redis: self-hosted initially.
- Queue: self-hosted RabbitMQ or Temporal depending on workflow needs.
- Object storage: optional for future OCR uploads.

### WhatsApp Cost

For 1,000 active users, expected cost may range from Rp200,000 to Rp1,000,000/month depending on notification volume, provider pricing, and conversation category.

Cost controls:

- Batch non-urgent reminders where acceptable.
- Avoid unnecessary bot chatter.
- Let users configure notification frequency.
- Track cost per active user and cost per successful reminder.

### AI Cost

For MVP intent extraction only, expected cost for 1,000 users may stay around Rp100,000 to Rp500,000/month depending on volume and model choice.

Cost controls:

- Use deterministic parsing for common patterns.
- Use AI only when rules fail.
- Cache repeated extraction patterns.
- Keep responses short.
- Avoid open-ended conversation in MVP.

## 15. Marketing and SEO Strategy

### 15.1 SEO Positioning

Woles should start from high-intent, problem-specific SEO instead of generic productivity keywords.

Primary keyword themes:

- Reminder pajak mobil
- Reminder STNK
- Reminder SIM
- Reminder servis mobil
- Reminder passport
- Reminder langganan bulanan
- Aplikasi pengingat pajak kendaraan
- Aplikasi pengingat dokumen penting
- WhatsApp reminder Indonesia

The SEO strategy should target users who already know the pain and are searching for a direct solution.

### 15.2 Landing Pages

Create focused landing pages for each high-intent use case:

- `/reminder-pajak-mobil`
- `/reminder-stnk`
- `/reminder-sim`
- `/reminder-servis-mobil`
- `/reminder-passport`
- `/reminder-subscription`
- `/whatsapp-reminder`

Each page should include:

- Clear use case headline.
- Explanation of the problem.
- How Woles solves it through WhatsApp.
- Example WhatsApp message.
- Reminder schedule example.
- Call to action to start through WhatsApp.
- FAQ with schema markup.

### 15.3 Content Plan

Initial content categories:

- Practical guides: how to remember STNK renewal, SIM expiry, passport expiry.
- Cost guides: vehicle tax reminder, subscription tracking, household bills.
- Templates: WhatsApp reminder examples for common life admin tasks.
- Comparison articles: calendar reminders vs WhatsApp reminders.
- Checklist articles: monthly household admin checklist.

### 15.4 Programmatic SEO

Create scalable pages from structured templates:

- Reminder for `{document_type}` expiry
- Reminder for `{vehicle_task}`
- Reminder for `{subscription_category}`
- Monthly checklist for `{life_context}` such as car owner, new couple, traveler, freelancer

Programmatic SEO must still be useful and locally relevant. Avoid thin pages.

### 15.5 Distribution Channels

Organic channels:

- SEO landing pages.
- TikTok and Instagram Reels showing before/after admin workflows.
- Twitter/X and LinkedIn founder-building updates.
- Indonesian personal finance and car-owner communities.
- WhatsApp group referrals.

Paid channels after validation:

- Google Search Ads for high-intent keywords.
- Retargeting for landing page visitors.
- Creator partnerships in personal finance, automotive, and household management niches.

### 15.6 Referral Strategy

Referral loop:

- User invites spouse or family member to shared reminders.
- User gets temporary Premium limit upgrade.
- Invited user gets free onboarding quota.

This should be introduced after the core reminder experience is reliable.

## 16. Launch Plan

### Phase 1: Private MVP

Goal: 50 first users.

Scope:

- WhatsApp onboarding.
- Reminder creation.
- Document expiry reminders.
- Subscription tracker.
- Timeline query.
- Basic dashboard.

Validation:

- At least 40% retention.
- At least 5 active reminders per active user.
- Manual support allowed.

### Phase 2: Public Beta

Goal: 1,000 users.

Scope:

- Improved AI extraction.
- Better dashboard UX.
- Billing integration.
- SEO landing pages.
- Notification analytics.

Validation:

- Free-to-premium conversion.
- WhatsApp cost per active user.
- Reminder delivery reliability.

### Phase 3: Public Launch

Goal: scalable acquisition and paid conversion.

Scope:

- Premium and Advanced plans.
- Family account beta.
- OCR beta.
- More SEO pages.
- Operational dashboards.

## 17. Risks and Mitigations

### WhatsApp Provider Risk

Risk: WhatsApp API pricing, template approval, or policy changes may affect cost and UX.

Mitigation:

- Keep web and email fallback options open.
- Track provider cost carefully.
- Avoid unnecessary outbound messages.

### Reminder Reliability Risk

Risk: Missed reminders damage trust immediately.

Mitigation:

- Use durable scheduling.
- Store notification attempts.
- Add retry and dead-letter handling.
- Monitor upcoming reminder execution.

### AI Misinterpretation Risk

Risk: Incorrect extraction creates wrong reminders.

Mitigation:

- Confirm important parsed values before saving.
- Ask clarification for ambiguous dates.
- Use deterministic parsers for common Indonesian patterns.

### Low Willingness to Pay

Risk: Users like reminders but do not upgrade.

Mitigation:

- Make Free useful but capped.
- Put timeline, goals, unlimited quota, family account, and OCR behind paid plans.
- Test annual pricing and family pricing.

## 18. Open Questions

- Which WhatsApp provider should be used for MVP?
- Should email login be required, or can WhatsApp number be the primary identity?
- Should AI extraction be included in the first MVP or added after rule-based parsing?
- Which payment provider should be used first: Midtrans, Xendit, or another provider?
- Should documents allow image upload in MVP, or only metadata?
- Should the web dashboard be available only after WhatsApp onboarding, or also support standalone registration?

## 19. MVP Definition of Done

The MVP is ready when:

- A user can onboard from WhatsApp.
- A user can create reminders, document expiry records, and subscriptions.
- Woles can send WhatsApp reminder notifications reliably.
- The dashboard shows upcoming items and statistics.
- The timeline can answer what is due this month and in the next 90 days.
- Basic plan limits are enforced.
- Notification attempts are logged.
- Admin can inspect failed notifications.
- The first SEO landing pages are live.
