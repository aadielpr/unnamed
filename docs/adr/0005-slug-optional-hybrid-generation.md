# Slug is optional; user-or-server generated with dual collision handling

`POST /api/events` accepts an optional `slug` in the request body. When the organizer supplies one, the server validates it (`^[a-z0-9-]+$`, lowercased server-side, 1–100 chars) and rejects a taken slug with 409 `slug_taken`. When the organizer omits it, the server generates a slug from the title (NFKD → lowercase → a-z0-9 → dashes, cap ~40; empty result falls back to 8-char base62), and on collision auto-suffixes `-2`, `-3`, … until it inserts.

This departs from issue #9's wording ("generates a unique readable slug") and DECISIONS.md #3's framing (server-derived). The organizer can now name their own slug, with the server as a fallback. The two paths deliberately diverge on collision handling: a user-supplied slug is the organizer's choice, so a conflict is *their* conflict to fix (409); a server-generated slug is our suggestion, so a conflict is *ours* to resolve silently (auto-suffix).

## Considered Options

- **Server-generated only (issue #9 literal)** — simplest form, lowest barrier, but removes the organizer's ability to choose `/e/sarahs-wedding` deliberately. Rejected: the organizer often has a string in mind.
- **User-supplied required** — full control, but adds a form field and a validation-conversation to a flow meant to be a quick phone action, and forces every organizer to learn the slug rules. Rejected: breaks the "minimal three fields" spirit.
- **Optional, hybrid (chosen)** — default path stays three fields; organizers who care can name a slug. Two collision rules, one per source, keep the right actor responsible for the conflict.

## Consequences

- The create form has no required slug field; an optional one lives behind the same "Advanced" toggle as the lifecycle overrides.
- Two error codes brunch off the slug: 400 `slug_invalid` (charset/length) and 409 `slug_taken` (unique). The frontend only branches on these two.
- The slug generation + collision logic lives in the `POST /api/events` handler (or a helper it calls), not in the DB: the `slug TEXT UNIQUE` constraint is the race-free backstop for the server-generated path, and the a priori validation + collision SELECT is the user-supplied path's front door.