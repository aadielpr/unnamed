# EventLens (working name) — MVP Decisions

> Source of truth for the MVP build spec. Locked via a grilling session on 2025-07-15. The wayfinder map (`wayfinder:map` issue on GitHub) is an index of open/resolved tickets; the full locked decisions live here, not in the map.

## Destination

A **deployed MVP, live on the internet, with the core loop working end-to-end**: an organizer creates an event, guests scan a QR and upload photos, and the photos appear in a shared gallery.

Proving it at a real ~200-guest event is the **next step after the MVP is live** — not part of the MVP. Build simple now; harden (upload retry, WebSocket, scale, auto-deletion) when that event approaches.

## Context

Portfolio project **and** a potential business. Tech: **Go backend + SolidJS SPA (Vite). Postgres. Object storage (Cloudflare R2 or S3) + CDN** for images. Hosting topology is undecided (tracked as a ticket); the code is hosting-agnostic.

## Locked decisions

1. **Upload mechanism** — capture-primary + library option via `<input type="file" accept="image/*" capture>`. No custom in-app camera UI in the MVP.
2. **Safety** — 3-photo limit per guest; a confirmation screen before upload goes live; **instant publish (delete-after-publish)**; organizer can delete any photo; guests can report. Approve-before-publish is deferred.
3. **Organizer auth** — token URL, no accounts, no email. Two links per event: the **guest link** is a readable slug (what the QR encodes, e.g. `/e/sarah-bday`); the **admin link** carries the event id + a secret token. The admin link is shown once on-screen at creation; if lost, the token is delivered manually (WhatsApp/in person). A `/portal` page accepts a pasted token to log back in. The event id alone never grants admin powers.
4. **Guest identity** — server-side guest session scoped per event (cookie). Upload count lives in the DB. Bypassable by clearing cookies / re-scanning — acceptable; the 3-photo limit is a blast-radius cap, not a hard wall. The real backstops are organizer delete + guest report.
5. **Gallery** — polling every ~3–5s for the MVP. WebSocket deferred (for the big-screen live slideshow / TV mode later).
6. **Image processing** — on upload, generate a **thumbnail** (~400px wide, JPEG q~75) and a **display image** (max ~2048px long edge, JPEG q~80). Store both in object storage; **skip raw originals in the MVP** (full-res storage returns as a premium-tier feature once billing exists).
7. **Lifecycle** — two independent windows: `upload_closes_at` (short, default ~1 day) and `gallery_expires_at` (longer, gallery stays viewable/downloadable). Plus a manual `is_closed` kill switch. All enforced **lazily on access — no scheduler in the MVP**. Actual physical deletion of expired assets is deferred.
8. **Event page (guest-facing)** — title, date / `starts_at`, one-line description, countdown to start (client-side, before start only), gallery, upload button.
9. **Featured photos** — organizer can feature (pin) selected photos to the top of the gallery. `featured_at` timestamp on the photo (null = not featured); gallery sorts featured-first. Manual curation only; AI highlights deferred.
10. **Organizer dashboard** — total uploads counter, delete any photo, reported-photos queue, feature/unfeature, close-event button, download-all ZIP (display images). Guests get a "my uploads" filter (data already exists per uploader).

## Out of scope for the MVP (deferred / future)

Cover photo; likes / comments / guestbook / hashtags; live slideshow / TV display mode (the WebSocket upgrade); analytics; multi-event organizer accounts; all AI features (best-shot, blur/duplicate detection, face grouping, highlight album); storage tiers / billing; raw original full-res storage.

**Permanent scope boundary:** this is **not a social network** — no followers, following, DMs, or influencer-style profiles. The event, not the user, is the center.

## Open items (tracked as GitHub tickets under the wayfinder map)

- **Name the app** — deferred (resolved 2025-07-18). Stick with "EventLens" (working name) and `unnamed` repo. Frontrunner "Kala" (KBBI: moment); revisit post-MVP.
- **Hosting topology** (`wayfinder:grilling`) — where the Go server, Postgres, and object storage + CDN live. Doesn't block building code; blocks going live.
- **Image resolution tiers** (`wayfinder:research`) — how professionals serve responsive multi-level images (`/low` `/med` `/high`); refine beyond the MVP's 2-version approach. Doesn't block the MVP.
- **Two-window lifecycle solid implementation** (`wayfinder:research`) — incl. the **cron/scheduler** question, auto-deletion mechanics, the USB/physical-handoff business flow, storage-tier durations, and quick-event validation. Doesn't block the MVP.
