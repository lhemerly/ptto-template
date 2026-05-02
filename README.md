# 🥔 ptto-template: The Indestructible Go Stack

> The official "Patient Zero" application template for the `ptto` deployment engine.

This repository is a fully functional, production-ready web application template. It does not display a generic "Hello World" or a Todo list. Out of the box, it serves as an interactive **Performance and Authentication Showcase**, proving exactly what a minimalist stack can do.

If you are looking for microservices, React components, JSON REST APIs, or third-party Auth providers (Google/Auth0), you are in the wrong place.

## The Stack

* **Logic & Server**: Go
* **Database**: Embedded SQLite (via `modernc.org/sqlite` - **No CGO required**)
* **Views**: `templ` (Type-safe HTML components compiled to Go)
* **Reactivity**: HTMX (Zero client-side state)
* **Styling**: Tailwind CSS
* **Authentication**: Native WebAuthn (Passkeys) via SQLite sessions.

## Philosophy

1. **One File to Rule Them All**: This entire application compiles into a single, statically linked Linux binary.
2. **Zero-Config Security**: Passkeys are built-in. Users authenticate using their device's biometric hardware. We do not store passwords, and we do not send "magic link" emails.
3. **Data Proximity**: The database lives in the same memory space as the application. Queries execute in microseconds. 

## Local Development

```bash
# Clone the template
git clone [https://github.com/lhemerly/ptto-template-go](https://github.com/lhemerly/ptto-template-go) my-app
cd my-app

# Required CLIs
go install github.com/a-h/templ/cmd/templ@latest
go install github.com/air-verse/air@latest
# install tailwindcss standalone binary and ensure it's on PATH

# Run the dev watcher (auto-compiles templ, tailwind, and go)
make dev

# App is live at http://localhost:8080
```

## Showcase Endpoints

* **`POST /latency-ping`**: Returns total server response time, SQLite query time, and SQLite clock timestamp.
* **`GET /resource-monitor`**: Returns live memory allocation in MB (used by the footer poller).

## Deployment

This template is designed specifically to be deployed by `ptto`.

```bash
# Connect to your VPS
ptto init root@your-server-ip

# Compile and deploy
ptto deploy --domain your-app.com
```
