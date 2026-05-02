# Implementation Roadmap (Showcase Todo List)

This is the task list to complete the `ptto-template-go` showcase. Once these are checked off, this repo will be embedded into the `ptto new` CLI command.

### Phase 1: Core Engine & Build Pipeline
- [x] Initialize `go.mod` and project structure (`/cmd`, `/internal`).
- [x] Integrate `modernc.org/sqlite` and auto-create a WAL-mode `data.sqlite` on startup.
- [x] Create `Makefile` with a `dev` command to concurrently watch/build `templ`, Tailwind, and Go.

### Phase 2: The Interactive Showcase (Views)
- [ ] **The Hero Section**: Minimalist landing page displaying the current server render time (e.g., "Rendered in 42µs").
- [ ] **The Latency Ping**: A button leveraging `hx-post` that hits the SQLite DB and returns the server's timestamp, proving network-speed UI updates without JSON parsing.
- [ ] **The Resource Monitor**: A tiny footer widget using `hx-get` with `hx-trigger="every 5s"` to stream the Go binary's current RAM usage (proving the sub-20MB footprint).

### Phase 3: The Potato Auth (Passkeys)
- [ ] Create SQLite schema for `users`, `credentials` (passkeys), and HTTP-only `sessions`.
- [ ] Implement Go WebAuthn library (`github.com/go-webauthn/webauthn`).
- [ ] **WebAuthn Sandbox UI**: A section on the landing page where a user can click "Register Device" to trigger the browser's native biometric prompt.
- [ ] Implement HTMX swapping upon successful auth to replace the "Register" button with the user's active session data and credential ID.
