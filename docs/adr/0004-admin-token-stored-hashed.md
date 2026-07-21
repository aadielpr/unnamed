# Admin token is stored hashed

The admin token is a bearer credential — one per event, shown to the organizer exactly once at creation, used to authenticate the organizer on the admin link and the `/portal` paste flow. It is stored as a SHA-256 digest in `events.admin_token_hash`, never as plaintext. The plaintext token is returned in the single `POST /api/events` 201 response and never again.

A DB leak should not grant admin power over every event. Hashing costs nothing: the server only ever compares a supplied token against the stored digest (on `/portal` and admin endpoints), never looks the token up to return it. The admin link carries the event id *and* the token; the server looks up the event by id, hashes the supplied token, and compares — so no unique index on the hash is required and the event id alone still grants nothing (DECISIONS.md #3).

## Considered Options

- **Plaintext with a unique index** — only needed if the admin link were looked up by token alone. Rejected: the link carries the event id, so the compare is scoped by id; plaintext buys nothing and a DB leak compromises every event.
- **Hashed (chosen)** — SHA-256 of 32 random bytes (base64url, 43 chars). Compare-by-event-id keeps it simple; the digest never needs to be unique across events.

## Consequences

- The `events.admin_token_hash` column stores a 64-char hex digest (or the base64 of it) — not the 43-char token.
- The plaintext token exists in exactly one place: the `POST /api/events` 201 response. Storing or logging it anywhere else is a bug.
- Losing the token is unrecoverable from the DB; recovery is out-of-band (DECISIONS.md #3: "delivered manually, WhatsApp/in person") or via the `/portal` page *only after* the organizer already has the token to paste. The server cannot re-display it.