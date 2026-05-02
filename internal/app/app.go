package app

import (
	"database/sql"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/lhemerly/ptto-template-go/internal/db"
)

type App struct {
	db *sql.DB
}

func New() (*App, error) {
	return NewWithDBPath("data.sqlite")
}

func NewWithDBPath(path string) (*App, error) {
	database, err := db.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	return &App{db: database}, nil
}

func (a *App) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.handleHome)
	mux.HandleFunc("/latency-ping", a.handleLatencyPing)
	mux.HandleFunc("/resource-monitor", a.handleResourceMonitor)
	return mux
}

func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()
	renderTime := time.Since(start).Microseconds()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprintf(w, `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>ptto-template showcase</title>
  <script src="https://unpkg.com/htmx.org@1.9.12"></script>
</head>
<body style="font-family: sans-serif; margin: 2rem; line-height:1.5;">
  <main>
    <h1>ptto-template: Interactive Showcase</h1>
    <p>Rendered in %dµs</p>

    <section>
      <h2>Latency Ping</h2>
      <button hx-post="/latency-ping" hx-target="#latency-result" hx-swap="innerHTML">Ping SQLite</button>
      <div id="latency-result" style="margin-top: 0.5rem; color: #333;"></div>
    </section>
  </main>

  <footer style="margin-top: 2rem; font-size: 0.9rem; color: #555;">
    RAM usage: <span id="resource-monitor" hx-get="/resource-monitor" hx-trigger="load, every 5s" hx-swap="innerHTML">loading...</span>
  </footer>
</body>
</html>`, renderTime)
}

func (a *App) handleLatencyPing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var ts string
	if err := a.db.QueryRow(`SELECT strftime('%Y-%m-%dT%H:%M:%fZ', 'now');`).Scan(&ts); err != nil {
		http.Error(w, "database query failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("SQLite time: " + ts))
}

func (a *App) handleResourceMonitor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	mb := float64(mem.Alloc) / (1024 * 1024)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(strconv.FormatFloat(mb, 'f', 2, 64) + " MB"))
}

func (a *App) Close() error {
	return a.db.Close()
}
