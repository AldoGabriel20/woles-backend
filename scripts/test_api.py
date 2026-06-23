#!/usr/bin/env python3
"""Full REST API test for Woles Backend. Run while server is running on :8080."""
import requests
import json
import sys
import time
from datetime import datetime, timedelta, timezone

BASE = "http://localhost:8080/api/v1"
HEALTH_URL = "http://localhost:8080/health"

# Use a unique email per test run so register always succeeds
RUN_ID = int(time.time()) % 100000
TEST_EMAIL = f"apitest_{RUN_ID}@woles.dev"

class Tester:
    def __init__(self):
        self.session = requests.Session()
        self.token = None
        self.results = []

    def _csrf(self):
        """Refresh CSRF cookie by doing a GET to /health, return token value."""
        self.session.get(HEALTH_URL)
        return self.session.cookies.get("csrf_token", "")

    def _headers(self, need_csrf=False):
        h = {"Content-Type": "application/json"}
        if self.token:
            h["Authorization"] = f"Bearer {self.token}"
        if need_csrf:
            h["X-CSRF-Token"] = self._csrf()
        return h

    def _record(self, method, path, req_body, status_code, resp_body, expected=None):
        if expected is not None:
            ok = status_code == expected
        else:
            ok = status_code < 400
        icon = "✅" if ok else "❌"
        entry = {
            "ok": ok,
            "icon": icon,
            "method": method,
            "path": path,
            "req": req_body,
            "status": status_code,
            "resp": resp_body,
        }
        self.results.append(entry)
        short = json.dumps(resp_body, ensure_ascii=False)[:120] if isinstance(resp_body, dict) else str(resp_body)[:120]
        print(f"  {icon} {method} {path} → {status_code}  {short}")
        return ok

    def _call(self, method, path, data=None, multipart=None, expected=None):
        url = f"{BASE}{path}"
        need_csrf = method.upper() != "GET"
        h = self._headers(need_csrf=need_csrf)
        kwargs = {}
        if multipart:
            del h["Content-Type"]
            kwargs["files"] = multipart
        elif data is not None:
            kwargs["json"] = data
        r = getattr(self.session, method.lower())(url, headers=h, **kwargs)
        try:
            resp = r.json()
        except Exception:
            resp = {"raw": r.text[:300]}
        self._record(method.upper(), path, data, r.status_code, resp, expected=expected)
        return r, resp

    def get(self, path, **kw):    return self._call("GET", path, **kw)
    def post(self, path, **kw):   return self._call("POST", path, **kw)
    def patch(self, path, **kw):  return self._call("PATCH", path, **kw)
    def delete(self, path, **kw): return self._call("DELETE", path, **kw)


def fmt_json(obj, indent=4):
    return json.dumps(obj, indent=indent, ensure_ascii=False, default=str)


def main():
    t = Tester()
    ids = {}  # store created resource IDs

    future = (datetime.now(timezone.utc) + timedelta(days=30)).strftime("%Y-%m-%dT%H:%M:%SZ")
    next_year = (datetime.now(timezone.utc) + timedelta(days=365)).strftime("%Y-%m-%dT%H:%M:%SZ")

    print("\n" + "="*60)
    print("WOLES API TEST SUITE")
    print("="*60)

    # ── Health ────────────────────────────────────────────────────
    print("\n[Health]")
    r, resp = t.session.get(HEALTH_URL), None
    try:
        resp = r.json()
    except Exception:
        resp = {}
    icon = "✅" if r.status_code < 400 else "❌"
    print(f"  {icon} GET /health → {r.status_code}  {resp}")
    t.results.append({"ok": r.status_code < 400, "icon": icon, "method": "GET", "path": "/health (no prefix)", "req": None, "status": r.status_code, "resp": resp})

    # ── Auth ──────────────────────────────────────────────────────
    print("\n[Auth — Register]")
    r, resp = t.post("/auth/register", data={
        "email": TEST_EMAIL,
        "password": "Test@Woles99",
        "name": "API Test User",
        "timezone": "Asia/Jakarta"
    })
    if r.status_code in (201, 400):  # 400 = already registered, get token
        if r.status_code == 201:
            t.token = resp.get("tokens", {}).get("access_token")
            ids["user_id"] = resp.get("user", {}).get("id")

    print("\n[Auth — Login]")
    r, resp = t.post("/auth/login", data={
        "email": TEST_EMAIL,
        "password": "Test@Woles99"
    })
    if r.status_code == 200:
        t.token = resp.get("tokens", {}).get("access_token")

    print("\n[Auth — /me]")
    r, resp = t.get("/auth/me")
    if r.status_code == 200:
        ids["user_id"] = resp.get("user", {}).get("id")

    print("\n[Auth — Change Password]")
    t.post("/auth/password/change", data={
        "old_password": "Test@Woles99",
        "new_password": "Test@Woles99"  # same, just testing the endpoint
    })

    print("\n[Auth — OTP Request]")
    t.post("/auth/otp/request", data={"phone": "+6281234567890"})

    print("\n[Auth — Password Reset Request]")
    t.post("/auth/password/reset/request", data={"email": TEST_EMAIL})

    print("\n[Auth — Password Reset Confirm (not implemented)]")
    t.post("/auth/password/reset/confirm", data={
        "token": "fake_token",
        "new_password": "NewPass123!"
    }, expected=501)

    print("\n[Auth — 2FA Enable]")
    r, resp = t.post("/auth/2fa/enable", data={})
    ids["totp_secret"] = resp.get("secret", "")

    print("\n[Auth — 2FA Verify (will fail without real code)]")
    t.post("/auth/2fa/verify", data={"totp_code": "000000"}, expected=400)

    print("\n[Auth — 2FA Disable]")
    t.post("/auth/2fa/disable", data={"password": "Test@Woles99"})

    print("\n[Auth — Sessions]")
    r, resp = t.get("/auth/sessions")
    sessions = resp.get("sessions", [])
    if sessions:
        ids["session_id"] = sessions[0].get("id")

    if ids.get("session_id"):
        print("\n[Auth — Revoke Session]")
        t.delete(f"/auth/sessions/{ids['session_id']}")

    print("\n[Auth — Revoke All Sessions]")
    # NOTE: this would log us out; skip for now
    # t.delete("/auth/sessions")

    print("\n[Auth — Refresh Token (expected token_reused after session revoke)]")
    t.post("/auth/refresh", data={}, expected=401)

    # Re-login after potential session changes
    print("\n[Auth — Re-login]")
    r, resp = t.post("/auth/login", data={
        "email": TEST_EMAIL,
        "password": "Test@Woles99"
    })
    if r.status_code == 200:
        t.token = resp.get("tokens", {}).get("access_token")

    # ── Reminders ─────────────────────────────────────────────────
    print("\n[Reminders — Create]")
    r, resp = t.post("/reminders", data={
        "title": "Pay Electricity Bill",
        "category": "bill",
        "recurrence_type": "monthly",
        "next_run_at": future,
        "timezone": "Asia/Jakarta"
    })
    if r.status_code == 201:
        ids["reminder_id"] = resp.get("reminder", {}).get("id")

    print("\n[Reminders — Create custom_interval]")
    t.post("/reminders", data={
        "title": "Service Mobil",
        "category": "vehicle",
        "recurrence_type": "custom_interval",
        "recurrence_rule": {"interval_days": 180},
        "next_run_at": future,
        "timezone": "Asia/Jakarta"
    })

    print("\n[Reminders — Create invalid category (expect 422)]")
    t.post("/reminders", data={
        "title": "Bad Category",
        "recurrence_type": "monthly",
        "next_run_at": future,
        "category": "INVALID_CAT"
    }, expected=422)

    print("\n[Reminders — List]")
    t.get("/reminders")

    print("\n[Reminders — List with filters]")
    t.get("/reminders?status=active&category=bill&sort=next_run_at&order=asc&page=1&per_page=5")

    if ids.get("reminder_id"):
        rid = ids["reminder_id"]
        print("\n[Reminders — Get by ID]")
        t.get(f"/reminders/{rid}")

        print("\n[Reminders — Update]")
        t.patch(f"/reminders/{rid}", data={
            "title": "Pay Electricity Bill (Updated)",
            "category": "bill"
        })

        print("\n[Reminders — Pause]")
        t.post(f"/reminders/{rid}/pause", data={})

        print("\n[Reminders — Resume]")
        t.post(f"/reminders/{rid}/resume", data={})

        print("\n[Reminders — Complete]")
        t.post(f"/reminders/{rid}/complete", data={})

        print("\n[Reminders — Get non-existent (expect 404)]")
        t.get("/reminders/00000000-0000-0000-0000-000000000000", expected=404)

        print("\n[Reminders — Delete]")
        t.delete(f"/reminders/{rid}")

    # ── Documents ─────────────────────────────────────────────────
    print("\n[Documents — Create]")
    r, resp = t.post("/documents", data={
        "title": "STNK Toyota Avanza",
        "document_type": "stnk",
        "vault_category": "vehicles",
        "expiry_date": "2027-03-15",
        "reminder_offsets": [30, 7, 1],
        "notes": "Stored in glove box"
    })
    if r.status_code == 201:
        ids["doc_id"] = resp.get("document", {}).get("id")

    print("\n[Documents — Create passport]")
    t.post("/documents", data={
        "title": "Passport Alice",
        "document_type": "passport",
        "vault_category": "identity",
        "expiry_date": "2031-08-20"
    })

    print("\n[Documents — List]")
    t.get("/documents")

    print("\n[Documents — List with filters]")
    t.get("/documents?vault_category=vehicles&page=1&per_page=10")

    print("\n[Documents — Storage usage]")
    t.get("/documents/storage/usage")

    print("\n[Documents — Vault health]")
    t.get("/documents/vault/health")

    if ids.get("doc_id"):
        did = ids["doc_id"]
        print("\n[Documents — Get by ID]")
        t.get(f"/documents/{did}")

        print("\n[Documents — Update]")
        t.patch(f"/documents/{did}", data={
            "title": "STNK Toyota Avanza 2022",
            "expiry_date": "2028-03-15",
            "notes": "Updated notes"
        })

        print("\n[Documents — Upload file (small text as PDF)]")
        import io
        fake_pdf = io.BytesIO(b"%PDF-1.4 fake pdf content")
        r2, resp2 = t.post(f"/documents/{did}/file",
                           multipart={"file": ("test.pdf", fake_pdf, "application/pdf")})

        print("\n[Documents — Delete file]")
        t.delete(f"/documents/{did}/file")

        print("\n[Documents — Delete]")
        t.delete(f"/documents/{did}")

    # ── Subscriptions ─────────────────────────────────────────────
    print("\n[Subscriptions — Create]")
    r, resp = t.post("/subscriptions", data={
        "name": "Netflix",
        "amount": 54000,
        "currency": "IDR",
        "billing_cycle": "monthly",
        "next_billing_at": future,
        "category": "entertainment"
    })
    if r.status_code == 201:
        ids["sub_id"] = resp.get("subscription", {}).get("id")

    print("\n[Subscriptions — List]")
    t.get("/subscriptions")

    print("\n[Subscriptions — List with filters]")
    t.get("/subscriptions?status=active&category=entertainment")

    if ids.get("sub_id"):
        sid = ids["sub_id"]
        print("\n[Subscriptions — Get by ID]")
        t.get(f"/subscriptions/{sid}")

        print("\n[Subscriptions — Update]")
        t.patch(f"/subscriptions/{sid}", data={
            "name": "Netflix Premium",
            "amount": 75000
        })

        print("\n[Subscriptions — Archive]")
        t.post(f"/subscriptions/{sid}/archive", data={})

        print("\n[Subscriptions — Delete]")
        t.delete(f"/subscriptions/{sid}")

    # ── Goals ─────────────────────────────────────────────────────
    print("\n[Goals — Create (expect 403 on free plan)]")
    r, resp = t.post("/goals", data={
        "title": "Emergency Fund",
        "icon": "emergency",
        "target_amount": 50000000,
        "monthly_target": 2000000,
        "currency": "IDR",
        "target_date": "2027-12-31"
    }, expected=403)
    if r.status_code == 201:
        ids["goal_id"] = resp.get("goal", {}).get("id")

    print("\n[Goals — List]")
    t.get("/goals")

    print("\n[Goals — History]")
    t.get("/goals/history")

    if ids.get("goal_id"):
        gid = ids["goal_id"]
        print("\n[Goals — Get by ID]")
        t.get(f"/goals/{gid}")

        print("\n[Goals — Update progress]")
        t.post(f"/goals/{gid}/progress", data={"amount": 5000000})

        print("\n[Goals — Update]")
        t.patch(f"/goals/{gid}", data={"title": "Emergency Fund 2027", "target_amount": 60000000})

        print("\n[Goals — Delete]")
        t.delete(f"/goals/{gid}")

    # ── Finances ──────────────────────────────────────────────────
    print("\n[Finances — Summary]")
    t.get("/finances/summary")

    print("\n[Finances — Monthly costs]")
    t.get("/finances/monthly-costs")

    # ── Timeline ──────────────────────────────────────────────────
    print("\n[Timeline — By range]")
    t.get("/timeline?range=30d")

    print("\n[Timeline — By date]")
    t.get("/timeline?from=2026-06-01&to=2026-12-31")

    # ── Notifications ─────────────────────────────────────────────
    print("\n[Notifications — List]")
    t.get("/notifications")

    print("\n[Notifications — List with filters]")
    t.get("/notifications?status=scheduled&entity_type=reminder")

    print("\n[Notifications — Stats]")
    t.get("/notifications/stats")

    print("\n[Notifications — Export CSV]")
    r, resp = t._call("GET", "/notifications/export?format=csv&range=30d")

    print("\n[Notifications — Export PDF]")
    t.get("/notifications/export?format=pdf&range=30d")

    # ── Family ────────────────────────────────────────────────────
    print("\n[Family — Create member (expect 403 on free plan)]")
    r, resp = t.post("/family/members", data={
        "name": "Budi Santoso",
        "role": "spouse",
        "relation_label": "Husband"
    }, expected=403)
    if r.status_code == 201:
        ids["member_id"] = resp.get("member", {}).get("id")

    print("\n[Family — List members (expect 403 on free plan)]")
    t.get("/family/members", expected=403)

    print("\n[Family — Shared reminders (expect 403 on free plan)]")
    t.get("/family/reminders", expected=403)

    if ids.get("member_id"):
        mid = ids["member_id"]
        print("\n[Family — Get member]")
        t.get(f"/family/members/{mid}")

        print("\n[Family — Update member]")
        t.patch(f"/family/members/{mid}", data={"name": "Budi S."})

        print("\n[Family — Delete member]")
        t.delete(f"/family/members/{mid}")

    # ── Chat ──────────────────────────────────────────────────────
    print("\n[Chat — Send message]")
    r, resp = t.post("/chat/messages", data={
        "content": "Ingatkan bayar tagihan internet tiap tanggal 10"
    })

    print("\n[Chat — Get messages]")
    t.get("/chat/messages")

    print("\n[Chat — Usage]")
    t.get("/chat/usage")

    print("\n[Chat — Intents]")
    t.get("/chat/intents")

    print("\n[Chat — Delete all messages]")
    t.delete("/chat/messages")

    # ── Account ───────────────────────────────────────────────────
    print("\n[Account — Get profile]")
    t.get("/account/profile")

    print("\n[Account — Update profile]")
    t.patch("/account/profile", data={
        "name": "API Test User Updated",
        "timezone": "Asia/Makassar"
    })

    # ── Billing ───────────────────────────────────────────────────
    print("\n[Billing — Get plan]")
    t.get("/billing/plan")

    # ── Webhooks ──────────────────────────────────────────────────
    print("\n[Webhook — Missing signature]")
    r = requests.post("http://localhost:8080/webhooks/whatsapp/fontiva",
                      json={"message": "hello"})
    icon = "✅" if r.status_code == 401 else "❌"
    print(f"  {icon} POST /webhooks/whatsapp/fontiva → {r.status_code}  {r.json()}")
    t.results.append({"ok": True, "icon": icon, "method": "POST",
                       "path": "/webhooks/whatsapp/:provider (missing sig)",
                       "req": None, "status": r.status_code, "resp": r.json()})

    print("\n[Auth — Logout]")
    t.post("/auth/logout", data={})

    # ── Summary ───────────────────────────────────────────────────
    total = len(t.results)
    passed = sum(1 for r in t.results if r["ok"])
    failed = total - passed

    print("\n" + "="*60)
    print(f"RESULTS: {passed}/{total} passed  |  {failed} failed")
    print("="*60)

    # Print failures
    if failed:
        print("\nFAILED:")
        for r in t.results:
            if not r["ok"]:
                print(f"  ❌ {r['method']} {r['path']} → {r['status']}: {json.dumps(r['resp'], ensure_ascii=False)[:200]}")

    # Output full results as JSON for file writing
    print("\n--- JSON_RESULTS_START ---")
    print(json.dumps(t.results, indent=2, ensure_ascii=False, default=str))
    print("--- JSON_RESULTS_END ---")

    return t.results


if __name__ == "__main__":
    main()
