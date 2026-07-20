# Use Echo as the web framework

The PRD and issue #7 originally said "HTTP server (+ thin router)." During the T1 scaffold grilling we chose **Echo** (a web framework) over stdlib `net/http` or `chi`, for the developer experience of built-in JSON binding, middleware, and error handling on a solo MVP. This over-delivers on "thin router" — issue #7's wording is updated to "web framework."

Handlers are tied to `echo.Context`, so reversing later means rewriting handlers. We accepted that lock-in for faster DX now.

## Considered Options

- **stdlib `net/http` (Go 1.22+ `ServeMux`)** — zero deps, handlers stay plain `http.HandlerFunc`. Rejected: more boilerplate for binding/middleware on a solo build.
- **`chi`** — thin router, `http.HandlerFunc`-compatible. Rejected: still no built-in binding/validation; if we go beyond stdlib we wanted the full framework DX.
- **Gin** — has a bigger ecosystem but more magic and baggage. Echo has a smaller surface and felt easier to reason about for our `httptest` seam.

## Consequences

- Handler signatures are `func(c echo.Context) error`, not `http.HandlerFunc`. Switching routers later is a rewrite.
- Integration tests use `e.ServeHTTP(w, r)` against the real `echo.Echo` — stays compatible with the PRD's "one seam: the HTTP API" testing decision.