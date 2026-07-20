# EventLens (working name) — MVP PRD

> Published spec for the MVP build. Synthesizes `DECISIONS.md` (the locked grilling decisions) and the closed wayfinder research tickets (#3, #4, #5). The wayfinder map (#1) is the decision artifact; this is the build spec.

## Problem Statement

Organizers of an event (a birthday, a wedding, a casual gathering) want to collect the photos their guests take into one shared gallery, without forcing guests to download an app, sign up, or hand photos to a person.

Today the alternatives are: a shared cloud folder nobody uploads to on the night, a WhatsApp group that buries photos in scrollback, or one tired organizer collecting phones at the end. None of them produce a single gallery that everyone can relive the event through.

## Solution

A web app, centered on the **event** (not the user). The core loop:

1. An organizer creates an event on their phone and gets two links.
2. They share the **guest link** (a readable slug, what the QR codes) at the event.
3. Guests scan the QR, land on the event page, and upload photos directly from their phone camera roll via the native file picker (`capture` attribute).
4. Photos appear in a shared gallery, polling every few seconds.
5. The organizer gets an **admin link** (carrying a secret token) that opens a dashboard: see total uploads, delete any photo, feature (pin) photos to the top, review reported photos, close the event, and download a ZIP of all photos.
6. When the upload window closes, uploading stops. When the gallery expires, the gallery stops being viewable. There is no permanent storage of anyone's photos, and no social graph around any of it.

No accounts. No email. No follow/DM/profile. The event is the center.

## User Stories

### Event creation

1. As an organizer, I want to create an event from my phone by entering a title, date/time, and one-line description, so that guests have something recognizable to land on.
2. As an organizer, I want the event to give me two links — a readable **guest link** (e.g. `/e/sarah-bday`) and a secret **admin link** — so that I can share one publicly and keep one private.
3. As an organizer, I want the admin token shown to me once on-screen at creation, with a warning that it must be saved, so that I understand it won't be shown again.
4. As an organizer, if I lose the admin link, I want to be able to paste a token into a `/portal` page to get back to my dashboard, so that I'm not locked out forever.
5. As an organizer, I want the guest link encoded as a QR I can scan / project, so that guests at a physical event can join with one tap.

### Guest identity & upload

6. As a guest, I want to land on the event page by scanning the QR, with no sign-up or login, so that the barrier to upload is as low as possible.
7. As a guest, I want my phone's native photo picker (with capture) to open when I tap upload, so that I can snap a new photo or pick existing ones.
8. As a guest, I want a **confirmation screen** before my upload goes live, so that I can change my mind or remove a photo I didn't mean to add.
9. As a guest, I want my three-photo limit enforced, so that one guest can't flood the gallery.
10. As a guest, I want to see how many uploads I have left (3 − used), so that I know whether I can add more.
11. As a guest, I want my upload count to persist even if I clear my cookies or re-scan the QR (accepted as bypassable — the limit is a blast-radius cap, not a hard wall), so that… well, it doesn't; I understand clearing cookies resets me and that's an accepted tradeoff.

### Gallery

12. As a guest, I want to open the gallery and see the event's photos load automatically, so that I can relive the event as it happens.
13. As a guest, I want the gallery to refresh on its own every few seconds, so that photos guests just uploaded appear without me reloading.
14. As a guest, I want photos the organizer **featured** to appear at the top of the gallery, so that I see the best shots first.
15. As a guest, I want to see the event title, date/time, one-line description, and a countdown to start (if the event hasn't started), so that the page feels like an event, not a folder.
16. As a guest, I want to download a photo I like, so that I can keep one.
17. As a guest, I want to see a "my uploads" filter showing only the photos I uploaded (driven by my per-event guest session), so that I can find my own shots in a big gallery.

### Organizer dashboard

18. As an organizer, I want to see a total upload counter, so that I know how much the event is producing.
19. As an organizer, I want to delete any photo from the dashboard, so that I can remove a bad shot, a duplicate, or something inappropriate — instantly.
20. As an organizer, I want to feature (pin) and unfeature photos, so that I can curate the top of the gallery.
21. As an organizer, I want a queue of reported photos, so that I can act on guest reports in one place.
22. As an organizer, I want to close the event manually (kill switch), so that uploads stop on command even before the window clock says so.
23. As an organizer, I want to download all photos as a ZIP of the display images, so that I can take the event home with me.
24. As an organizer, I want the dashboard to work on my phone, so that I can curate from the event floor.

### Safety & moderation

25. As a guest, I want to report a photo, so that the organizer can review it.
26. As an organizer, I want deletion to be instant and to remove the photo from the gallery immediately, so that something inappropriate disappears now, not later.
27. As an organizer, I want the 3-photo limit and the confirmation screen to be in place from day one, so that I'm not fighting a flood.

### Lifecycle

28. As an organizer, I want to set how long uploads stay open (`upload_closes_at`, default ~1 day), so that the gallery stops accumulating new photos after the moment has passed.
29. As an organizer, I want to set how long the gallery stays viewable (`gallery_expires_at`, longer than uploads), so that the event disappears from the internet on a schedule, not never.
30. As a guest, I do not want to be able to upload after `upload_closes_at` or after the organizer closed the event — I want a clear "uploading has closed" state, so that I'm not left guessing.
31. As a guest, when the gallery has expired I do not want to see any photos (the event 404s / shows an expired state), so that the event genuinely ends.

### Operational

32. As an organizer, I want the whole thing running live on the internet, so that a real event can use it.
33. As a developer, I want the box reprovisioned from `.env` + git (Neon holds the DB, R2 holds images), so that a redeploy is safe and stateless.
34. As a developer, I want expired-asset physical deletion to happen lazily / via a periodic DB-driven sweep with one R2 lifecycle rule as backstop, so that storage doesn't quietly fill up.

## Implementation Decisions

### Stack & shape

- **Backend:** Go. HTTP server (stdlib or a thin router). JSON-over-HTTP API; no GraphQL.
- **Frontend:** SolidJS SPA, built with Vite. Served as static assets by the Go server (or Caddy in front — see hosting).
- **Database:** Postgres on Neon (managed, Singapore region, co-located with the droplet). Migrations via a tool chosen during the build (likely `goose` or `golang-migrate`).
- **Object storage:** Cloudflare R2, exposed through Cloudflare CDN.
- **Hosting:** DigitalOcean Droplet (Singapore, stateless) via Docker Compose: `go-app` + `caddy` + `redis` (provisional). All persistent state is in Neon (DB) and R2 (images); the droplet holds no state, so redeploys via `.env` + git are safe. See closed ticket #3.

### Data model (high-level, not names as final)

- `events`: id, slug, title, description, `starts_at`, `upload_closes_at`, `gallery_expires_at`, `is_closed`, admin token (secret), created_at.
- `photos`: id, event_id, uploader guest session id, storage key (thumbnail), storage key (display), `featured_at` (nullable), `uploaded_at`, deleted_at (soft or hard — decided during the build).
- `guest_sessions`: per-event server-side cookie session → an opaque id; upload count is derived from `photos` grouped by session, not stored separately (or a small counter table — resolved during build).
- `reports`: id, photo_id, reported_at, resolved_at.

### Auth

- **Two links per event** (decision #3):
  - Guest link = `/e/<slug>` — public, the readable slug the QR encodes.
  - Admin link = carries the event id + a secret token. One secret per event. The event id alone grants nothing.
- Admin link shown **once** at creation. Lost token is delivered manually (WhatsApp/in person) — out of band.
- `/portal` accepts a pasted token → sets a cookie that grants admin powers for that event.
- Guest session = per-event server-side cookie (decision #4). Upload count lives in the DB keyed by guest session id. Clearing cookies resets you (accepted).

### Upload mechanism

- `<input type="file" accept="image/*" capture>` (decision #1). No custom in-app camera UI.
- 3-photo limit per guest, enforced server-side keyed to the guest session.
- **Confirmation screen** before the upload goes live (decision #2 safety).
- **Instant publish** — no approve-before-publish in the MVP.

### Image processing (decision #6 + ticket #4)

- On upload, generate two versions server-side:
  - **thumbnail** (~400px wide, JPEG q~75),
  - **display image** (max ~2048px long edge, JPEG q~80).
- Both stored in R2. **No raw originals in the MVP** (full-res returns as a premium-tier feature, post-MVP).
- Wire `srcset` + `sizes` on the gallery `<img>` so browsers pick the right version (free, ticket #4 recommendation). Serve both versions through the CDN.
- Pre-generation (on upload) vs on-the-fly transform location: decided during the build. Hosting is a single droplet (ticket #3), so an on-the-fly transform would run alongside the Go app on the box; the on-the-fly vs pre-gen decision is a small one left to the implementer of the upload slice.

### Gallery

- Polling every ~3–5s (decision #5). WebSocket deferred (the big-screen live slideshow / TV mode is a later feature).
- Sort: **featured-first** (by `featured_at`), then `uploaded_at` desc.
- Guest page shows title, date, one-line description, countdown to start (client-side, before start only).
- "My uploads" filter for guests uses the per-event guest session id already on the photo rows.

### Lifecycle (decision #7 + ticket #5)

- Two independent windows: `upload_closes_at` (short, default ~1 day) and `gallery_expires_at` (longer). Plus a manual `is_closed` kill switch.
- Enforced **lazily on access — no scheduler daemon in the MVP**:
  - On any guest upload attempt: reject if `now > upload_closes_at` OR `is_closed = true`.
  - On any gallery view: if `now > gallery_expires_at`, return an expired/404 state and do not serve photos.
- Physical deletion of expired assets: a periodic DB-driven **batch sweep** that lists expired events and deletes their R2 objects with free `DeleteObject` calls; plus one wide **R2 lifecycle rule** as a backstop. No scheduler daemon.
- Reserve the design seam of **"deletion trigger as DB state"** (close flag / expiry timestamps are the source of truth; the sweep is the effect) so the USB/physical-handoff flow and tier durations can attach as later branches without a rewrite.

### Organizer dashboard

- Total uploads counter.
- Delete any photo (instant; removes from gallery).
- Feature / unfeature (sets/clears `featured_at`).
- Reported-photos queue (from `reports` table).
- Close-event button (sets `is_closed = true`).
- Download-all ZIP (display images). Streaming approach decided during the build slice.
- Works on mobile (responsive, not a separate app).

### Hosting / deploy

- Docker Compose on the DO droplet: `go-app` + `caddy` + `redis` (provisional — only if a session store is needed; Neon may carry it in the DB).
- Caddy handles TLS and reverse-proxies to the Go app; serves the SPA static assets.
- `.env` + git reprovision the box. All state in Neon + R2.

### Out of scope — cross-reference

See `DECISIONS.md` "Out of scope" and the map's "Out of scope" section: cover photo; likes/comments/guestbook/hashtags; live slideshow/TV (WebSocket); analytics; multi-event organizer accounts; all AI features (best-shot, blur/dup detection, face grouping, highlight album); storage tiers/billing; raw original full-res; permanent social-graph boundary (no followers/DM/profiles).

## Testing Decisions

- **One seam: the HTTP API.** Handler-level integration tests in Go, using `httptest` against a real test Postgres (a test DB on Neon or a local instance) and a real test object storage (MinIO in place of R2) plus the real image pipeline. This is the highest seam that still covers the whole core loop as external behavior.
  - Good test = asserts on the HTTP response (status, body, headers, cookies) and on observable side effects in the DB / object storage. Doesn't assert on internal function calls or struct shapes.
  - Each test scenario reads like a user story (create event → claim admin → guest session → upload 3 → confirm → gallery shows them → feature → gallery reorders → close → upload rejected → expire → gallery 404s).
- **MinIO stands in for R2** via the S3-compatible API; the app uses the S3 SDK against an interface so R2 and MinIO are swappable by config.
- **A real test Postgres**, not a mock, so SQL and migrations are actually exercised.
- **Frontend tests: out of scope for now.** The SPA is exercised indirectly via the API contract. We add Playwright later when the API seam can't catch client wiring. (User decision, this session.)
- TDD at this seam: write a failing integration test for the next behavior in the slice, then make it pass.

## Out of Scope

- All items in `DECISIONS.md` "Out of scope" — cover photo, likes/comments/guestbook/hashtags, live slideshow/TV (WebSocket), analytics, multi-event organizer accounts, all AI features, storage tiers/billing, raw original full-res storage.
- **Permanent boundary:** this is not a social network — no followers, following, DMs, or influencer-style profiles.
- Frontend automated tests (deferred to when the API seam stops being enough).
- Proving at a real ~200-guest event (that's the step *after* MVP live).
- Post-MVP fog tracked in the wayfinder map's "Not yet specified": app naming/"Kala", business model/pricing, which AI features come first, multi-event organizer accounts, quick-event product concept, on-the-fly vs pre-gen image processing as a standalone decision (it'll be folded into the upload slice).

## Further Notes

- **Source of truth for locked decisions:** `DECISIONS.md`. This PRD restates them as user stories and build decisions; if the two ever conflict, `DECISIONS.md` wins.
- **Wayfinder map:** issue #1. All decision tickets (#2–#5) closed; #1 retires when `/to-tickets` produces the build tracking issue.
- **Hosting topology detail:** closed ticket #3.
- **Image tiers detail:** closed ticket #4 (research branch `research/image-resolution-tiers`).
- **Lifecycle detail:** closed ticket #5 (research branch `research/two-window-lifecycle`).
- The MVP is **lean on purpose**: build the core loop end-to-end first, harden (upload retry, WebSocket, scale, auto-deletion) when a real event approaches.