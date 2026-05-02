package app

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/lhemerly/ptto-template-go/internal/app/views"
	"github.com/lhemerly/ptto-template-go/internal/db"
)

const (
	sessionCookieName     = "ptto_session"
	registerCookieName    = "ptto_webauthn_reg"
	registerSessionMaxAge = 10 * time.Minute
)

type App struct {
	db       *sql.DB
	webauthn *webauthn.WebAuthn
}
type webauthnUser struct {
	id          int64
	displayName string
	webauthnID  []byte
}

func (u webauthnUser) WebAuthnID() []byte                         { return u.webauthnID }
func (u webauthnUser) WebAuthnName() string                       { return fmt.Sprintf("user-%d", u.id) }
func (u webauthnUser) WebAuthnDisplayName() string                { return u.displayName }
func (u webauthnUser) WebAuthnCredentials() []webauthn.Credential { return nil }

func New() (*App, error) { return NewWithDBPath("data.sqlite") }
func NewWithDBPath(path string) (*App, error) {
	database, err := db.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	w, err := webauthn.New(&webauthn.Config{RPDisplayName: "ptto-template", RPID: "localhost", RPOrigins: []string{"http://localhost:8080"}})
	if err != nil {
		return nil, fmt.Errorf("create webauthn: %w", err)
	}
	return &App{db: database, webauthn: w}, nil
}
func (a *App) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.handleHome)
	mux.HandleFunc("/latency-ping", a.handleLatencyPing)
	mux.HandleFunc("/resource-monitor", a.handleResourceMonitor)
	mux.HandleFunc("/webauthn/register/start", a.handleWebAuthnRegisterStart)
	mux.HandleFunc("/webauthn/register/finish", a.handleWebAuthnRegisterFinish)
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	return mux
}

func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	start := time.Now()
	if err := views.Home(0, false, "", "").Render(context.Background(), &bytes.Buffer{}); err != nil {
		http.Error(w, "render failed", http.StatusInternalServerError)
		return
	}
	renderMicros := time.Since(start).Microseconds()
	active, sessionID, credentialID := a.lookupSessionForView(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.Home(renderMicros, active, sessionID, credentialID).Render(r.Context(), w); err != nil {
		http.Error(w, "render failed", http.StatusInternalServerError)
	}
}

func (a *App) handleWebAuthnRegisterStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, err := a.createUser()
	if err != nil {
		http.Error(w, "create user failed", http.StatusInternalServerError)
		return
	}
	opts, sessionData, err := a.webauthn.BeginRegistration(user)
	if err != nil {
		http.Error(w, "begin registration failed", http.StatusInternalServerError)
		return
	}
	registrationID, err := randomToken(32)
	if err != nil {
		http.Error(w, "registration setup failed", http.StatusInternalServerError)
		return
	}
	sessionBytes, err := json.Marshal(sessionData)
	if err != nil {
		http.Error(w, "registration setup failed", http.StatusInternalServerError)
		return
	}
	if err := a.storeRegistrationState(registrationID, user.id, sessionBytes, time.Now().Add(registerSessionMaxAge)); err != nil {
		http.Error(w, "registration setup failed", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, a.makeCookie(r, registerCookieName, registrationID, registerSessionMaxAge))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(struct {
		Status    string                                      `json:"status"`
		PublicKey protocol.PublicKeyCredentialCreationOptions `json:"publicKey"`
	}{Status: "ok", PublicKey: opts.Response}); err != nil {
		http.Error(w, "encode failed", http.StatusInternalServerError)
		return
	}
}

func (a *App) handleWebAuthnRegisterFinish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Credential json.RawMessage `json:"credential"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Credential) == 0 {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	regCookie, err := r.Cookie(registerCookieName)
	if err != nil {
		http.Error(w, "missing registration state", http.StatusBadRequest)
		return
	}
	userID, sessionData, err := a.loadRegistrationState(regCookie.Value)
	if err != nil {
		http.Error(w, "invalid registration state", http.StatusBadRequest)
		return
	}
	user, err := a.userByID(userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusBadRequest)
		return
	}
	parsedCredential, err := protocol.ParseCredentialCreationResponseBytes(body.Credential)
	if err != nil {
		http.Error(w, "parse credential failed", http.StatusBadRequest)
		return
	}
	credential, err := a.webauthn.CreateCredential(user, *sessionData, parsedCredential)
	if err != nil {
		http.Error(w, "finish registration failed", http.StatusBadRequest)
		return
	}
	if err := a.storeCredential(user.id, credential); err != nil {
		http.Error(w, "store credential failed", http.StatusInternalServerError)
		return
	}
	sessionID, err := a.createSession(user.id, credential.ID)
	if err != nil {
		http.Error(w, "create session failed", http.StatusInternalServerError)
		return
	}
	_ = a.deleteRegistrationState(regCookie.Value)
	http.SetCookie(w, a.makeCookie(r, registerCookieName, "", -1))
	http.SetCookie(w, a.makeCookie(r, sessionCookieName, sessionID, 24*time.Hour))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.AuthSessionCard(sessionID, hex.EncodeToString(credential.ID)).Render(r.Context(), w); err != nil {
		http.Error(w, "render failed", http.StatusInternalServerError)
	}
}

func (a *App) createUser() (webauthnUser, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return webauthnUser{}, fmt.Errorf("random webauthn id: %w", err)
	}
	res, err := a.db.Exec(`INSERT INTO users(display_name,webauthn_id) VALUES(?,?)`, fmt.Sprintf("Potato %d", time.Now().UnixNano()), b)
	if err != nil {
		return webauthnUser{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return webauthnUser{}, err
	}
	return webauthnUser{id: id, displayName: "Potato", webauthnID: b}, nil
}
func (a *App) userByID(id int64) (webauthnUser, error) {
	var u webauthnUser
	err := a.db.QueryRow(`SELECT id,display_name,webauthn_id FROM users WHERE id=?`, id).Scan(&u.id, &u.displayName, &u.webauthnID)
	return u, err
}
func (a *App) storeCredential(userID int64, c *webauthn.Credential) error {
	tr := make([]string, 0, len(c.Transport))
	for _, t := range c.Transport {
		tr = append(tr, string(t))
	}
	_, err := a.db.Exec(`INSERT INTO credentials(user_id,credential_id,public_key,aaguid,sign_count,transports) VALUES(?,?,?,?,?,?)`, userID, c.ID, c.PublicKey, c.Authenticator.AAGUID, c.Authenticator.SignCount, strings.Join(tr, ","))
	return err
}
func (a *App) createSession(userID int64, credentialID []byte) (string, error) {
	sid, err := randomToken(32)
	if err != nil {
		return "", err
	}
	_, err = a.db.Exec(`INSERT INTO sessions(id,user_id,credential_id,expires_at) VALUES(?,?,?,?)`, sid, userID, credentialID, time.Now().Add(24*time.Hour).UTC().Format(time.RFC3339Nano))
	return sid, err
}
func randomToken(size int) (string, error) {
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
func (a *App) makeCookie(r *http.Request, name, value string, maxAge time.Duration) *http.Cookie {
	secure := r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
	c := &http.Cookie{Name: name, Value: value, HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, Path: "/"}
	if maxAge < 0 {
		c.MaxAge = -1
		c.Expires = time.Unix(0, 0)
		return c
	}
	c.Expires = time.Now().Add(maxAge)
	return c
}
func (a *App) storeRegistrationState(id string, userID int64, sessionData []byte, expires time.Time) error {
	_, err := a.db.Exec(`INSERT INTO registration_states(id,user_id,session_data,expires_at) VALUES(?,?,?,?)`, id, userID, sessionData, expires.UTC().Format(time.RFC3339Nano))
	return err
}
func (a *App) loadRegistrationState(id string) (int64, *webauthn.SessionData, error) {
	var userID int64
	var payload []byte
	var expires string
	err := a.db.QueryRow(`SELECT user_id,session_data,expires_at FROM registration_states WHERE id=?`, id).Scan(&userID, &payload, &expires)
	if err != nil {
		return 0, nil, err
	}
	exp, err := time.Parse(time.RFC3339Nano, expires)
	if err != nil || time.Now().After(exp) {
		return 0, nil, errors.New("expired")
	}
	var sd webauthn.SessionData
	if err := json.Unmarshal(payload, &sd); err != nil {
		return 0, nil, err
	}
	return userID, &sd, nil
}
func (a *App) deleteRegistrationState(id string) error {
	_, err := a.db.Exec(`DELETE FROM registration_states WHERE id=?`, id)
	return err
}
func (a *App) lookupSessionForView(r *http.Request) (bool, string, string) {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return false, "", ""
	}
	var credentialID []byte
	var expiresAt string
	if err := a.db.QueryRow(`SELECT credential_id,expires_at FROM sessions WHERE id=?`, c.Value).Scan(&credentialID, &expiresAt); err != nil {
		return false, "", ""
	}
	exp, err := time.Parse(time.RFC3339Nano, expiresAt)
	if err != nil || time.Now().After(exp) {
		return false, "", ""
	}
	return true, c.Value, hex.EncodeToString(credentialID)
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
func (a *App) Close() error { return a.db.Close() }
