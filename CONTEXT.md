# EventLens

A web app centered on the **event** (not the user) that lets an organizer collect guests' photos into one shared gallery. No accounts, no email, no social graph. The event is the center.

## Language

**Event**:
A single gathering (a birthday, a wedding, a casual get-together) that owns a gallery, an upload window, and a gallery-expiry window. Created by an organizer. Identified to guests by a readable slug; identified to the organizer by a secret admin token.
_Avoid_: party, occasion, meetup

**Slug**:
A unique, human-readable string used in the guest link (e.g. `sarah-bday`). What the QR encodes. Lowercase alphanumeric + dashes, max 100 chars, globally unique per event. Optional at creation: if the organizer supplies one, it's validated and a taken slug is rejected (409) — not auto-suffixed; if absent, the server generates one from the title (random fallback when the title yields nothing), auto-suffixed on collision.
_Avoid_: handle, permalink, short code

**Guest link**:
The public URL `/e/<slug>`. Sharable, what guests scan or click, and what the QR encodes. Carries no secret, so it can be re-displayed anywhere — including on the admin dashboard — at any time. Distinct from the admin link, which is secret and shown once.
_Avoid_: public link, share link

**Admin token**:
A secret bearer credential, one per event, generated at creation and shown to the organizer exactly once. Grants admin powers for that event only. Stored hashed; the event id alone grants nothing.
_Avoid_: password, key, admin code

**Admin link**:
A URL that carries the event id and the admin token, shown once at creation so the organizer can save it. Losing it means recovering via the `/portal` paste flow (a later slice).
_Avoid_: dashboard link, management link

**Organizer**:
The person who creates an event and holds its admin token. Curates the gallery, closes the event, downloads the ZIP. Has no account.
_Avoid_: host, admin, owner

**Guest**:
A person who opens the guest link and uploads photos. Has no account; is tracked only by a per-event server-side session cookie.
_Avoid_: visitor, attendee, participant