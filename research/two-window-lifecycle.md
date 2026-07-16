# Research: two-window lifecycle — solid implementation (incl. scheduler)

> Wayfinder ticket [#5](https://github.com/aadielpr/unnamed/issues/5), child of map
> [#1](https://github.com/aadielpr/unnamed/issues/1). Branch: `research/two-window-lifecycle`.
>
> **Scope:** hardens the MVP decision recorded in `DECISIONS.md` §7 — two windows
> (`upload_closes_at`, `gallery_expires_at`) + manual `is_closed`, enforced lazily on access with
> **no scheduler in the MVP**. This file gathers facts a post-MVP hardening needs, and answers the
> five sub-questions on the ticket. It does **not** change the MVP spec; it captures findings +
> a recommended hardening direction.
>
> Generated against **primary sources** (Cloudflare R2 docs, AWS S3 docs). Where a section adds
> design reasoning beyond a cited fact, it is marked **[analysis]** so fact and recommendation are
> not confused.

---

## TL;DR (the decision the ticket waits on)

- **A long-running scheduler daemon is not required.** The clean hardening is
  **lazy-on-access enforcement (correctness) + a periodic batch sweep (cleanup)**. "Scheduler" in
  the ticket really meant "a cron to auto-revoke upload access and physically delete expired
  assets" — both can be done without a 24/7 scheduling process.
  - *Auto-revoke upload access* = a timestamp check on the upload request against
    `upload_closes_at`. No timer needed; the check happens for free on the request you'd serve
    anyway.
  - *Physical deletion* of expired assets = a **periodic batch sweep** that deletes objects via the
    S3 API then their DB rows, OR R2/S3 native **object-lifecycle rules** as a coarse backstop.
- **R2 native lifecycle rules** can auto-delete objects by prefix (and expire multipart uploads),
  but are a **coarse global TTL**, not a per-event custom-expiry clock: filtering is prefix-based
  (no per-event tag-driven duration on R2 as documented; S3 lifecycle does support tag filters),
  and deletion lags up to ~24h. Best used as a **safety-net backstop**, not the primary clock.
- **Deletion on R2 is free** (`DeleteObject` is a free operation), and **egress is free**, so a
  sweep job costs nothing in operations/egress — only the Class A `PutObject`/`LifecycleStorageTierTransition`
  you already pay for on upload.
- The **per-event** variable `gallery_expires_at` clock is best kept in **Postgres** (single source
  of truth), with the sweep job driving object-storage cleanup off the DB rows — not by trying to
  express every event's expiry as an object-storage rule.

---

## 1. Scheduler vs lazy + batch — the key implementation question

### 1a. Upload-access revocation: lazy only, no clock

The upload endpoint already reads the event row from Postgres to validate the guest session and
enforce the 3-photo limit (`DECISIONS.md` §2, §4). Comparing `now` against `upload_closes_at`
(and the manual `is_closed` flag) on that same read adds ~zero cost. When closed, return `410` /
"Uploads closed."

**No scheduler is needed to revoke upload access**, because access control already rides on the
request handler. A timer that flips a boolean at the stroke of `upload_closes_at` would add an
infrastructure dependency for no correctness gain — guests uploading a second late still get
rejected by the lazy check.

> Fact (primary source): a presigned / direct-storage upload model is *not* how the MVP works —
> uploads are processed server-side to generate the thumbnail + display image (`DECISIONS.md` §6).
> So upload revocation is a backend-gate decision, never delegated to object storage. (Even in a
> future direct-to-R2 model, R2 presigned URLs expire in **1s–7 days**
> ([R2 Presigned URLs](https://developers.cloudflare.com/r2/api/s3/presigned-urls/)), which is far
> coarser than a per-event `upload_closes_at`; the backend gate remains the real control.)

### 1b. Physical deletion: periodic batch sweep (recommended) or native lifecycle (backstop)

Two real options, which **compose**:

**Option A — periodic batch sweep (recommended primary mechanism).**
A job runs on a fixed interval (e.g. hourly) and:
1. `SELECT` events (and their photos) where `gallery_expires_at < now()` AND `assets_deleted_at IS NULL`.
2. `DeleteObject` each object's thumbnail + display key via the S3 API (R2 is S3-compatible).
3. On success, delete / soft-soft-delete the DB rows; mark `assets_deleted_at = now()`.

Run *it* three ways depending on hosting (kept independent of the undecided hosting ticket #3):
- a `time.Ticker` goroutine in the Go server (simplest; dies with the process — fine, next start
  resumes by scanning), or
- an external scheduler: a Cloudflare Cron Trigger / GitHub Actions `schedule` / systemd timer /
  `robfig/cron` if a cron library is preferred. Treated as a small, replaceable seam.

Properties: exact per-event expiry (driven by the DB `gallery_expires_at`), immediate on next
sweep, idempotent (re-sweep is safe), no extra storage-class cost. **Deletion itself is free on
R2** ([R2 Pricing](https://developers.cloudflare.com/r2/pricing/) — "Free operations include
`DeleteObject`, `DeleteBucket` and `AbortMultipartUpload`").

**Option B — R2 / S3 native object-lifecycle rules (backstop).**
Set a bucket lifecycle rule that expires objects under a prefix (e.g. `events/`) after N days — the
longest tier you ever want to keep. This is the platform doing the deleting for you.

Constraint facts (primary sources):
- R2: lifecycle rules expire objects "typically within 24 hours of the x-amz-expiration value";
  existing objects may lag longer on first rollout; max **1,000 rules** per bucket; filter shown is
  **prefix**-based ([R2 Object lifecycles](https://developers.cloudflare.com/r2/buckets/object-lifecycles/)).
  Rules can express "delete after N days" (`Expiration.Days`) or "delete on a specific date"
  (`Expiration.Date`). A default rule also aborts multipart uploads 7 days after initiation.
- S3 lifecycle: same 1,000-rule cap per bucket, but filters support **prefix, object size, object
  tags, or combinations** ([S3 Lifecycle configuration elements](https://docs.aws.amazon.com/AmazonS3/latest/userguide/intro-lifecycle-rules.html));
  S3 deletes expired objects on your behalf with no extra charge beyond the storage you stop
  paying for ([S3 managing lifecycle](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lifecycle-mgmt.html)).
- Both enforce a **fixed-duration TTL by prefix/tag**, not a free-form per-event timestamp. Good for
  "nothing outlives 90 days, ever"; **not** good as the *per-event* `gallery_expires_at` clock.

**[analysis] Recommended composition:** DB-driven batch sweep = the precise per-event clock +
deleter; a wide R2 lifecycle rule (e.g. expire anything under `events/` older than the max paid
retention) = a backstop that guarantees no object outlives the longest tier even if the sweep job is
down. The sweep handles correctness and exact expiry; the lifecycle rule handles the
"never leaks past the cap" guarantee.

### 1c. Gallery "expires" appearance: lazy, like upload

Showing the gallery as expired / read-only vs. gone can also be lazy: on the gallery fetch, compare
`now` vs `gallery_expires_at`. Before expiry → full gallery. After expiry → either hide
(`DECISIONS.md` already notes post-expiry behavior is deferred/undecided) or show a "this event's
gallery has ended" state. No clock needed for the *appearance* transition; only physical asset
deletion wants the sweep.

---

## 2. Auto-deletion mechanics — safely, irreversibly removing assets + DB rows

Facts (primary sources):
- **R2 delete is irreversible.** "Deleting objects from a bucket is irreversible"
  ([R2 Delete objects](https://developers.cloudflare.com/r2/objects/delete-objects/)); object
  lifecycles also permanently expire. R2 provides strong global consistency for reads/writes/deletes
  and lists ([R2 consistency model](https://developers.cloudflare.com/r2/reference/consistency/)).
- R2 has a **Bucket locks** feature (retention policies) that *prevents* deletion for compliance
  ([R2 Bucket locks](https://developers.cloudflare.com/r2/buckets/bucket-locks/)) — the opposite of
  what auto-deletion needs; do **not** enable it on the photo bucket.
- R2 **Event notifications** emit `object-delete` events (incl. lifecycle deletions) to a
  Cloudflare Queue
  ([R2 Event notifications](https://developers.cloudflare.com/r2/buckets/event-notifications/)) — a
  way to reconcile "object is gone" back into the DB without the sweep being the only writer.

**[analysis] Safe-and-irreversible deletion procedure (post-MVP hardening):**

1. **Order: delete from storage first, then commit the DB row delete.** If the process dies between
   the two, the orphan object gets reaped on the next sweep by the lifecycle backstop (it's under the
   prefix and old); an orphaned DB row pointing at no object is the less-bad failure (gallery 404/410
   rather than serving a phantom file).
2. **Idempotency:** tag/scope by `assets_deleted_at IS NULL`; retrying the sweep re-selects only
   not-yet-cleaned events. Each `DeleteObject` of a missing key is a no-op (S3/R2 delete is
   idempotent — deleting a non-existent object returns success).
3. **DB row handling:** either hard `DELETE` the photo rows, or keep them with `deleted_at` for an
   "event archive" view. Hard-delete keeps the schema dead-simple, matching the MVP's lean stance;
   keep soft-delete as a tunable if a future archive/restore feature appears.
4. **Multi-object efficiency / rate:** a single event holds at most a few hundred photos (the ~200-guest
   event * 3-photo cap is ~600 objects, two sizes each = ~1,200 keys). Sequential `DeleteObject` is
   fine at MVP scale; if it grows, S3/R2 `DeleteObjects` (batch, up to 1,000 keys per request) cuts
   requests. Watch Class A op volume only at much larger scale.
5. **Reclaim space / stop billing:** an object stops being billable once deleted; lifecycle-expired
   objects are billed only up to the eligibility time even if actual deletion lags
  ([R2 Object lifecycles](https://developers.cloudflare.com/r2/buckets/object-lifecycles/),
   [R2 Pricing](https://developers.cloudflare.com/r2/pricing/)).

Note on the CDN: images are served through a CDN (`DECISIONS.md` Context). A deleted object may stay
in the CDN cache until its `Cache-TTL` expires — set a modest `s-maxage` and/or purge on delete if a
hard "instantly gone everywhere" requirement ever appears. Not an MVP concern.

---

## 3. The "offload to USB / hand to customer" business flow

This is **[analysis]** all the way down — it's a product/ops flow, not something a primary source
specifies. Recorded so the lifecycle design doesn't accidentally foreclose it.

The two-window model supports a handoff instead of pure deletion cleanly, because **deletion is a
post-expiry action driven off the DB**, not wired into the storage. Concretely:

- Add an event-level state, e.g. `retention_end_state ∈ {pending, auto_delete, handed_off}` (or just
  `is_archived` / `handoff_at`). The sweep job only auto-deletes events that reach expiry in
  `auto_delete` state. An organizer who wants a handoff marks it `handed_off` (or it's set by
  downloading all), and the sweep **skips** it — assets stay until the physical handoff, then a
  manual "purge after handoff" removes them.
- The hardening already buys this for free: `download-all ZIP` (`DECISIONS.md` §10) is the customer
  handoff artifact. A "handoff" event is just "the download happened + the organizer confirms" →
  then run delete. No new storage primitive required; one extra enum column + one branch in the
  sweep.
- Concrete recommendation: do **not** build this in the MVP. Reserve the seam by keeping the
  *deletion trigger* a DB-column decision (not a hardcoded "expired → delete"), so the handoff
  branch is a later `if` rather than a refactor. This is the single lifecycle design choice with
  real forward leverage.

Open question for later (not researched here, not graduable yet): whether a handoff is a *paid*
feature (business model is still fog on the map). That decides whether the seam is just sketched or
built.

---

## 4. Storage-tier-based durations — how durations map to future free/paid tiers

Facts (primary sources) that constrain the design:
- R2 storage classes: **Standard** (no min duration, no retrieval fee) and **Infrequent Access**
  (30-day **minimum storage duration** + data-retrieval fees; billed for the full 30 days even if
  deleted earlier). Lifecycle can transition Standard → IA but **not back**
  ([R2 Storage classes](https://developers.cloudflare.com/r2/buckets/storage-classes/),
   [R2 Pricing](https://developers.cloudflare.com/r2/pricing/)).
- The free tier (10 GB-month storage, 1M Class A ops, 10M Class B ops/month, free egress) applies
  to Standard only — **not** to Infrequent Access ([R2 Pricing](https://developers.cloudflare.com/r2/pricing/)).

**[analysis] Mapping durations to tiers:**
- The clean mental model: **durations are a pricing-tier property, not a storage-class property.**
  Free tier = short retention (`gallery_expires_at` ~7 days, matching `photo.md`'s "7-day storage"
  free line); paid tiers = longer retention. The storage *class* choice (Standard vs IA) is a
  separate cost optimization that doesn't have to track the tier boundary.
- **Do not** use Infrequent Access for short-retention free photos: the 30-day minimum makes early
  deletion *more* expensive than Standard, and the free tier doesn't apply to IA. IA only pays off
  for photos kept >30 days and rarely read — i.e. a *long-retention paid tier* storage optimization,
  applied late (after the upload window), not a per-event knob.
- Implementation: store `tier_retention_days` per event/tier in Postgres; the sweep reads it to
  decide `gallery_expires_at` enforcement. Object-storage lifecycle rules can encode the *global
  cap* (e.g. "expire everything under `events/` at the longest paid retention") as the backstop
  described in §1b. S3 (if chosen over R2) additionally allows **tag-based** lifecycle
  (`tier=free` → expire at 7d) for a storage-side enforcement of tier durations; R2's documented
  filters at time of writing are prefix-based, so on R2 the DB sweep remains the primary clock.

This stays in MVP-friendly shape: durations are Postgres data, and the only storage-side hardening
reserved is a single backstop lifecycle rule.

---

## 5. Quick-event validation — does the two-window model support spontaneous 2–3h events?

**Yes — cleanly, with no schema change.** **[analysis] + small fact note:**

- The two-window model already has two independent timestamps: `upload_closes_at` (short) and
  `gallery_expires_at` (longer). For a 2–3h karaoke, set `upload_closes_at = starts_at + 2–3h`
  while `gallery_expires_at` stays at its normal multi-day/week value. The model doesn't care that
  the upload window is *hours* instead of *a day* — it's the same two comparisons. No new concept,
  no special case.
- The only MVP-imposed ceiling to check: the ~200-guest/3-photo assumptions. A short, dense event
  concentrates uploads in time (rate peak), not volume — a *throughput* concern (image processing +
  R2 `PutObject` rate), **not** a lifecycle concern. The lifecycle model is unaffected; a rate
  concern is a separate hardening axis (relevant to the ~200-guest event prep, which `DECISIONS.md`
  Destination already scopes as *after* the MVP).
- Recommendation: no change. Quick-events are listed as "Not yet specified" fog on the map
  (map §Not yet specified); this validation resharpens it slightly — **the lifecycle model does
  not block quick-events**, so any future quick-event ticket should focus on the *UI / event
  creation flow* (faster setup) and *upload-rate* hardening, not on lifecycle schema.

---

## Summary table — what to harden, and what to leave

| Sub-question | MVP (§7) stays | Post-MVP hardening |
|---|---|---|
| 1 Scheduler vs lazy+batch | Lazy on-access rejections (no scheduler) | Periodic DB-driven batch sweep (free `DeleteObject`) + wide R2 lifecycle backstop |
| 2 Auto-deletion mechanics | n/a (deferred) | Delete objects first, then rows; `assets_deleted_at` idempotent guard; `DeleteObjects` batch later |
| 3 USB / handoff flow | n/a | Reserve seam: keep the deletion trigger a DB state (`auto_delete` vs `handed_off`) so handoff is one branch |
| 4 Storage-tier durations | Single flat window | `tier_retention_days` in Postgres; IA only for long-retention paid tier; lifecycle backstop = max-tier days |
| 5 Quick-event validation | Already supported | None — confirm short `upload_closes_at`; rate hardening, not lifecycle, is the real axis |

**Headline for the map:** the two-window lifecycle is sound as specified in §7; its hardening needs
**no scheduler daemon**, only (a) the lazy rejections already planned, (b) a periodic batch sweep
driven off Postgres for exact per-event physical deletion, and (c) one storage-side lifecycle rule
as a backstop. The single forward-looking design choice worth making early is **keeping the
deletion trigger a DB-decided state** so the USB/handoff and tier-duration flows attach as later
branches instead of a rewrite.

---

## Sources (primary)

- Cloudflare R2 — Object lifecycles: https://developers.cloudflare.com/r2/buckets/object-lifecycles/
- Cloudflare R2 — Delete objects: https://developers.cloudflare.com/r2/objects/delete-objects/
- Cloudflare R2 — Storage classes: https://developers.cloudflare.com/r2/buckets/storage-classes/
- Cloudflare R2 — Pricing (free ops incl. DeleteObject; IA 30-day min; free tier): https://developers.cloudflare.com/r2/pricing/
- Cloudflare R2 — Event notifications (`object-delete` incl. LifecycleDeletion): https://developers.cloudflare.com/r2/buckets/event-notifications/
- Cloudflare R2 — Bucket locks (do not enable on photo bucket): https://developers.cloudflare.com/r2/buckets/bucket-locks/
- Cloudflare R2 — Consistency model (strong global): https://developers.cloudflare.com/r2/reference/consistency/
- Cloudflare R2 — Presigned URLs (1s–7d expiry; not used for MVP server-gated uploads): https://developers.cloudflare.com/r2/api/s3/presigned-urls/
- AWS S3 — Managing object lifecycle (deletes expired objects on your behalf): https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lifecycle-mgmt.html
- AWS S3 — Lifecycle configuration elements (filter by prefix/size/tags; 1,000-rule cap): https://docs.aws.amazon.com/AmazonS3/latest/userguide/intro-lifecycle-rules.html
- Internal: `DECISIONS.md` §7 (lifecycle), §2/§4 (upload gate), §6 (server-side image processing), §10 (download ZIP), Destination (~200-guest event is post-MVP). Map: GitHub issue #1.