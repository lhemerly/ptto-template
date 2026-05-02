package app

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/lhemerly/ptto-template-go/internal/app/views"
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
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	return mux
}

func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()
	if err := views.Home(0).Render(context.Background(), &bytes.Buffer{}); err != nil {
		http.Error(w, "render failed", http.StatusInternalServerError)
		return
	}
	renderMicros := time.Since(start).Microseconds()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.Home(renderMicros).Render(r.Context(), w); err != nil {
		http.Error(w, "render failed", http.StatusInternalServerError)
		return
	}
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
