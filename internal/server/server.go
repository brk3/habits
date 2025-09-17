package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/brk3/habits/internal/storage"
	"github.com/brk3/habits/pkg/habit"
	"github.com/brk3/habits/pkg/versioninfo"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/securecookie"
	"golang.org/x/oauth2"
)

// TODO(pbourke): implement from headers or path
const userID = "XXX"

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "habits_http_requests_total",
			Help: "Total number of HTTP requests by endpoint and method",
		},
		[]string{"endpoint", "method"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "habits_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint", "method"},
	)

	activeHabits = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "habits_active_habits_total",
			Help: "Total number of active habits",
		},
	)
)

type Server struct {
	// ⚡DB
	Store storage.Store

	// auth stuff (that I still dont know how to make optional)
	oauth2     *oauth2.Config
	oidcProv   *oidc.Provider
	idVerifier *oidc.IDTokenVerifier
	cookie     *securecookie.SecureCookie

	// secure in memory state storage
	state *StateStore
}

func New(store storage.Store, issuer, clientID, clientSecret, redirectURL string) (*Server, error) {
	ctx := context.Background()
	prov, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}

	verifier := prov.Verifier(&oidc.Config{ClientID: clientID})
	oauth2Cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     prov.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile"}, // minimal scope; "email", "groups"
	}

	sc := securecookie.New([]byte("HMAC-KEY-32B-MIN"), []byte("ENC-KEY-32B-MIN"))
	sc.MaxAge(86400)
	state := NewStateStore(5 * time.Minute)

	return &Server{
		Store:      store,
		oauth2:     oauth2Cfg,
		oidcProv:   prov,
		idVerifier: verifier,
		cookie:     sc,
		state:      state,
	}, nil
}

func writeJSON(w http.ResponseWriter, code int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(metricsMiddleware)

	// Public routes (not behind auth)
	r.Handle("/metrics", promhttp.Handler())
	r.Get("/version", s.getVersionInfo)

	// Routes for OIDC auth
	r.Route("/auth", func(r chi.Router) {
		r.Get("/login", s.login)
		r.Get("/callback", s.callback)
		r.Post("/logout", s.logout)
	})

	// I do not know how to make auth configurable.
	r.Route("/habits", func(r chi.Router) {
		//r.Use(s.authMiddleware)
		r.Post("/", s.trackHabit)
		r.Get("/", s.listHabits)
		r.Get("/{habit_id}", s.getHabit)
		r.Get("/{habit_id}/summary", s.getHabitSummary)
		r.Delete("/{habit_id}", s.deleteHabit)
	})

	return r
}

func (s *Server) getHabitSummary(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "habit_id")
	if id == "" {
		http.Error(w, `{"error":"habit id is required"}`, http.StatusBadRequest)
		return
	}

	currentStreak, longestSreak, err := s.computeStreaks(userID, id)
	if err != nil {
		http.Error(w, `{"error":"error computing streaks"}`, http.StatusInternalServerError)
		return
	}

	firstLogged, err := s.getFirstLogged(userID, id)
	if err != nil {
		http.Error(w, `{"error":"error retrieving first logged date"}`, http.StatusInternalServerError)
		return
	}

	totalDaysDone, err := s.computeTotalDaysDone(userID, id)
	if err != nil {
		http.Error(w, `{"error":"error computing total days done"}`, http.StatusInternalServerError)
		return
	}

	daysThisMonth, err := s.computeDaysThisMonth(userID, id)
	if err != nil {
		http.Error(w, `{"error":"error computing days this month"}`, http.StatusInternalServerError)
		return
	}

	bestMonth, err := s.computeBestMonth(userID, id)
	if err != nil {
		http.Error(w, `{"error":"error computing best month"}`, http.StatusInternalServerError)
		return
	}

	summary := habit.HabitSummary{
		Name:          id,
		CurrentStreak: currentStreak,
		LongestStreak: longestSreak,
		FirstLogged:   firstLogged,
		TotalDaysDone: totalDaysDone,
		BestMonth:     bestMonth,
		ThisMonth:     daysThisMonth,
		LastWrite:     time.Now().Unix(),
	}

	summaryResponse := HabitSummaryResponse{
		HabitID:      id,
		HabitSummary: summary,
	}
	if err := writeJSON(w, http.StatusOK, summaryResponse); err != nil {
		http.Error(w, `{"error":"failed to serialize response"}`, http.StatusInternalServerError)
		return
	}
}

func (s *Server) getVersionInfo(w http.ResponseWriter, _ *http.Request) {
	info := versioninfo.VersionInfo{
		Version:   versioninfo.Version,
		BuildDate: versioninfo.BuildDate,
	}
	if err := writeJSON(w, http.StatusOK, info); err != nil {
		http.Error(w, `{"error":"failed to serialize version info"}`, http.StatusInternalServerError)
		return
	}
}

func (s *Server) listHabits(w http.ResponseWriter, _ *http.Request) {
	names, err := s.Store.ListHabitNames(userID)
	if err != nil {
		http.Error(w, `{"error":"storage error"}`, http.StatusInternalServerError)
		return
	}
	if err := writeJSON(w, http.StatusOK, HabitListResponse{Habits: names}); err != nil {
		http.Error(w, `{"error":"failed to serialize response"}`, http.StatusInternalServerError)
		return
	}
}

func (s *Server) trackHabit(w http.ResponseWriter, r *http.Request) {
	var h habit.Habit
	if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if err := validateHabit(h); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	if err := s.Store.PutHabit(userID, h); err != nil {
		http.Error(w, `{"error":"database write failed"}`, http.StatusInternalServerError)
		return
	}

	habits, _ := s.Store.ListHabitNames(userID)
	activeHabits.Set(float64(len(habits)))

	if err := writeJSON(w, http.StatusCreated, h); err != nil {
		http.Error(w, `{"error":"failed to serialize response"}`, http.StatusInternalServerError)
		return
	}
}

func (s *Server) getHabit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "habit_id")
	if id == "" {
		http.Error(w, `{"error":"habit id is required"}`, http.StatusBadRequest)
		return
	}

	entries, err := s.Store.GetHabit(userID, id)
	if err != nil {
		http.Error(w, `{"error":"storage error"}`, http.StatusInternalServerError)
		return
	}
	if len(entries) == 0 {
		http.Error(w, `{"error":"habit not found"}`, http.StatusNotFound)
		return
	}

	h := HabitGetResponse{
		HabitID: id,
		Entries: entries,
	}
	if err := writeJSON(w, http.StatusOK, h); err != nil {
		http.Error(w, `{"error":"failed to serialize response"}`, http.StatusInternalServerError)
		return
	}
}

func (s *Server) deleteHabit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "habit_id")
	if id == "" {
		http.Error(w, `{"error":"habit id is required"}`, http.StatusBadRequest)
		return
	}

	err := s.Store.DeleteHabit(userID, id)
	if err != nil {
		http.Error(w, `{"error":"storage error"}`, http.StatusInternalServerError)
		return
	}

	habits, _ := s.Store.ListHabitNames(userID)
	activeHabits.Set(float64(len(habits)))

	w.WriteHeader(http.StatusNoContent)
}

func validateHabit(h habit.Habit) error {
	const maxNameLength = 20
	const maxNoteLength = 1024
	const minTS = 946684800
	const maxTS = 4102444800

	if len(h.Name) == 0 || len(h.Name) > maxNameLength {
		return fmt.Errorf("bad habit name: must be 1-%d characters", maxNameLength)
	}
	if len(h.Note) > maxNoteLength {
		return fmt.Errorf("bad habit note: must be 0-%d characters", maxNoteLength)
	}
	if h.TimeStamp < minTS || h.TimeStamp > maxTS {
		return fmt.Errorf("invalid timestamp")
	}

	return nil
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start).Seconds()
		httpRequestsTotal.WithLabelValues(r.URL.Path, r.Method).Inc()
		httpRequestDuration.WithLabelValues(r.URL.Path, r.Method).Observe(duration)
	})
}

// I reviewed the code, but it should be rewritten once this is working as desired. Or at least
// refactored and commented.
//
// BELOW HERE BE AI DRAGONS

type userCtxKey struct{}

type User struct {
	Subject string
	Email   string
	Claims  map[string]any
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		/// ???
		// I dont think this is needed anymore since I changed the routing in the chi server part above.
		p := r.URL.Path
		if strings.HasPrefix(p, "/auth/") || p == "/version" || strings.HasPrefix(p, "/metrics") {
			next.ServeHTTP(w, r)
			return
		}
		/// ???

		// 1) Try cookie, then Bearer
		var rawIDToken string
		if c, err := r.Cookie("id_token"); err == nil {
			if err := s.cookie.Decode("id_token", c.Value, &rawIDToken); err != nil {
				// bad cookie value; wipe it!
				http.SetCookie(w, &http.Cookie{Name: "id_token", Value: "", Path: "/", MaxAge: -1})
				rawIDToken = ""
			}
		}
		if rawIDToken == "" {
			if ah := r.Header.Get("Authorization"); strings.HasPrefix(ah, "Bearer ") {
				rawIDToken = strings.TrimPrefix(ah, "Bearer ")
			}
		}

		// 2) No token → redirect only for GET+HTML, else 401
		if rawIDToken == "" {
			if r.Method == http.MethodGet && acceptsHTML(r.Header.Get("Accept")) {
				http.Redirect(w, r, "/auth/login?return="+url.QueryEscape(r.URL.RequestURI()), http.StatusFound)
			} else {
				w.Header().Set("WWW-Authenticate", `Bearer realm="habits"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
			}
			return
		}

		// 3) Verify token; on failure, wipe cookie and redirect/401
		idTok, err := s.idVerifier.Verify(r.Context(), rawIDToken)
		if err != nil {
			http.SetCookie(w, &http.Cookie{Name: "id_token", Value: "", Path: "/", MaxAge: -1})
			if r.Method == http.MethodGet && acceptsHTML(r.Header.Get("Accept")) {
				http.Redirect(w, r, "/auth/login?return="+url.QueryEscape(r.URL.RequestURI()), http.StatusFound)
			} else {
				w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
			}
			return
		}

		// 4) Attach user to context
		var claims map[string]any
		_ = idTok.Claims(&claims)
		u := &User{
			Subject: idTok.Subject,
			Email:   strClaim(claims, "email"),
			Claims:  claims,
		}

		// Here we inject the User into the context for use in other areas after this middleware.
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userCtxKey{}, u)))
	})
}

// middleware helpers
func acceptsHTML(accept string) bool {
	return strings.Contains(accept, "text/html") || accept == ""
}

func strClaim(m map[string]any, k string) string {
	if v, ok := m[k].(string); ok {
		return v
	}
	return ""
}

// OIDC auth flow
// Need /login, /logout, and /callback endpoints. They don't have to be called exactly that, but
// they need to exist for those purposes.
//
// This should probably be a library dep but I had fun putting it together.

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	st := randState()
	verifier, err := genCodeVerifier(48)
	if err != nil {
		http.Error(w, "pkce gen failed", http.StatusInternalServerError)
		return
	}
	challenge := codeChallengeS256(verifier)

	/// ???
	// I dont know what this is doing, or rather WHY its doing it.
	// capture an optional return path (sanitize to keep it relative)
	ret := r.URL.Query().Get("return")
	if ret == "" {
		ret = "/"
	} else {
		// keep it relative to avoid open redirects
		if u, err := url.Parse(ret); err != nil || u.IsAbs() || u.Host != "" {
			ret = "/"
		}
	}
	/// ???

	s.state.Put(st, authState{
		Verifier: verifier,
		Return:   ret,
		ExpireAt: time.Now().Add(5 * time.Minute),
	})

	authURL := s.oauth2.AuthCodeURL(
		st,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (s *Server) callback(w http.ResponseWriter, r *http.Request) {
	st := param(r, "state")
	if st == "" {
		http.Error(w, "missing state", http.StatusBadRequest)
		return
	}
	code := param(r, "code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	saved, ok := s.state.GetAndDelete(st)
	if !ok || saved.Verifier == "" {
		http.Error(w, "invalid or expired state", http.StatusBadRequest)
		return
	}

	tok, err := s.oauth2.Exchange(
		r.Context(),
		code,
		oauth2.SetAuthURLParam("code_verifier", saved.Verifier),
	)
	if err != nil {
		http.Error(w, "code exchange failed", http.StatusBadGateway)
		return
	}
	rawIDToken, _ := tok.Extra("id_token").(string)
	if rawIDToken == "" {
		http.Error(w, "no id_token", http.StatusBadGateway)
		return
	}
	if _, err := s.idVerifier.Verify(r.Context(), rawIDToken); err != nil {
		http.Error(w, "id_token invalid", http.StatusUnauthorized)
		return
	}

	val, _ := s.cookie.Encode("id_token", rawIDToken)
	http.SetCookie(w, &http.Cookie{
		Name:     "id_token",
		Value:    val,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((8 * time.Hour).Seconds()),
	})

	http.Redirect(w, r, saved.Return, http.StatusFound)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "id_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

// Secure cookie storage for tokens
type authState struct {
	Verifier string
	Return   string
	ExpireAt time.Time
}

type StateStore struct {
	ttl time.Duration
	mu  sync.Mutex
	m   map[string]authState
}

func NewStateStore(ttl time.Duration) *StateStore {
	s := &StateStore{ttl: ttl, m: make(map[string]authState)}
	go func() { // janitor
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			s.mu.Lock()
			for k, v := range s.m {
				if now.After(v.ExpireAt) {
					delete(s.m, k)
				}
			}
			s.mu.Unlock()
		}
	}()
	return s
}

func (s *StateStore) Put(key string, v authState) {
	s.mu.Lock()
	s.m[key] = v
	s.mu.Unlock()
}

func (s *StateStore) GetAndDelete(key string) (authState, bool) {
	s.mu.Lock()
	v, ok := s.m[key]
	if ok {
		delete(s.m, key)
	}
	s.mu.Unlock()
	if ok && time.Now().After(v.ExpireAt) {
		return authState{}, false
	}
	return v, ok
}

// helpers
func genCodeVerifier(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func codeChallengeS256(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
func randState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
func param(r *http.Request, name string) string {
	if v := r.URL.Query().Get(name); v != "" {
		return v
	}
	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		if v := r.PostFormValue(name); v != "" {
			return v
		}
	}
	return ""
}
