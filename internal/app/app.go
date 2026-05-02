package app

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/lhemerly/ptto-template-go/internal/db"
)

type App struct {
	db *sql.DB
}

func New() (*App, error) {
	database, err := db.Open("data.sqlite")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	return &App{db: database}, nil
}

func (a *App) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ptto-template core engine is up\n"))
	})
	return mux
}

func (a *App) Close() error {
	return a.db.Close()
}
