package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/brk3/habits/internal/config"
	"github.com/brk3/habits/internal/storage"
	"github.com/brk3/habits/pkg/habit"
	"github.com/brk3/habits/pkg/versioninfo"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/securecookie"
	"golang.org/x/oauth2"
)

// TODO(pbourke): implement from token
const userID = "XXX"

type Server struct {
	store    storage.Store
	authConf *AuthConfig
	cfg      *config.Config
}

type AuthConfig struct {
	oauth2     *oauth2.Config
	oidcProv   *oidc.Provider
	idVerifier *oidc.IDTokenVerifier
	cookie     *securecookie.SecureCookie
	state      *StateStore
}

func New(cfg *config.Config, store storage.Store) (*Server, error) {
	srv := &Server{
		store: store,
		cfg:   cfg,
	}

	if cfg.AuthEnabled {
		clientID := cfg.OIDCProviders[0].ClientID
		clientSecret := cfg.OIDCProviders[0].ClientSecret
		issuerURL := cfg.OIDCProviders[0].IssuerURL
		redirectURL := cfg.OIDCProviders[0].RedirectURL
		scopes := cfg.OIDCProviders[0].Scopes

		prov, err := oidc.NewProvider(context.Background(), issuerURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
		}

		verifier := prov.Verifier(&oidc.Config{ClientID: clientID})
		oauth2Cfg := &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     prov.Endpoint(),
			RedirectURL:  redirectURL,
			Scopes:       scopes,
		}

		sc := securecookie.New([]byte("HMAC-KEY-32B-MIN"), []byte("ENC-KEY-32B-MIN"))
		sc.MaxAge(86400)
		srv.authConf = &AuthConfig{
			oauth2:     oauth2Cfg,
			oidcProv:   prov,
			idVerifier: verifier,
			cookie:     sc,
			state:      NewStateStore(5 * time.Minute),
		}
	}

	return srv, nil
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

	r.Handle("/metrics", promhttp.Handler())
	r.Get("/version", s.getVersionInfo)

	if s.cfg.AuthEnabled {
		r.Route("/auth", func(r chi.Router) {
			r.Get("/login", s.login)
			r.Get("/callback", s.callback)
			r.Post("/logout", s.logout)
		})
	}

	r.Route("/habits", func(r chi.Router) {
		if s.cfg.AuthEnabled {
			r.Use(s.authMiddleware)
		}
		r.Post("/", s.trackHabit)
		r.Get("/", s.listHabits)
		r.Get("/{habit_id}", s.getHabit)
		r.Get("/{habit_id}/summary", s.getHabitSummary)
		r.Delete("/{habit_id}", s.deleteHabit)
	})

	return r
}

func (s *Server) getHabitSummary(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "habit_id")
	if userID == "" || habitID == "" {
		http.Error(w, `{"error":"user id and habit id are required"}`, http.StatusBadRequest)
		return
	}

	currentStreak, longestSreak, err := s.computeStreaks(userID, habitID)
	if err != nil {
		http.Error(w, `{"error":"error computing streaks"}`, http.StatusInternalServerError)
		return
	}

	firstLogged, err := s.getFirstLogged(userID, habitID)
	if err != nil {
		http.Error(w, `{"error":"error retrieving first logged date"}`, http.StatusInternalServerError)
		return
	}

	totalDaysDone, err := s.computeTotalDaysDone(userID, habitID)
	if err != nil {
		http.Error(w, `{"error":"error computing total days done"}`, http.StatusInternalServerError)
		return
	}

	daysThisMonth, err := s.computeDaysThisMonth(userID, habitID)
	if err != nil {
		http.Error(w, `{"error":"error computing days this month"}`, http.StatusInternalServerError)
		return
	}

	bestMonth, err := s.computeBestMonth(userID, habitID)
	if err != nil {
		http.Error(w, `{"error":"error computing best month"}`, http.StatusInternalServerError)
		return
	}

	summary := habit.HabitSummary{
		Name:          habitID,
		CurrentStreak: currentStreak,
		LongestStreak: longestSreak,
		FirstLogged:   firstLogged,
		TotalDaysDone: totalDaysDone,
		BestMonth:     bestMonth,
		ThisMonth:     daysThisMonth,
		LastWrite:     time.Now().Unix(),
	}

	summaryResponse := HabitSummaryResponse{
		HabitID:      habitID,
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

func (s *Server) listHabits(w http.ResponseWriter, r *http.Request) {
	if userID == "" {
		http.Error(w, `{"error":"user id is required"}`, http.StatusBadRequest)
		return
	}
	names, err := s.store.ListHabitNames(userID)
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
	if userID == "" {
		http.Error(w, `{"error":"user id is required"}`, http.StatusBadRequest)
		return
	}
	var h habit.Habit
	if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if err := validateHabit(h); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	if err := s.store.PutHabit(userID, h); err != nil {
		http.Error(w, `{"error":"database write failed"}`, http.StatusInternalServerError)
		return
	}

	habits, _ := s.store.ListHabitNames(userID)
	activeHabits.Set(float64(len(habits)))

	if err := writeJSON(w, http.StatusCreated, h); err != nil {
		http.Error(w, `{"error":"failed to serialize response"}`, http.StatusInternalServerError)
		return
	}
}

func (s *Server) getHabit(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "habit_id")
	if userID == "" || habitID == "" {
		http.Error(w, `{"error":"user id and habit id are required"}`, http.StatusBadRequest)
		return
	}

	entries, err := s.store.GetHabit(userID, habitID)
	if err != nil {
		http.Error(w, `{"error":"storage error"}`, http.StatusInternalServerError)
		return
	}
	if len(entries) == 0 {
		http.Error(w, `{"error":"habit not found"}`, http.StatusNotFound)
		return
	}

	h := HabitGetResponse{
		HabitID: habitID,
		Entries: entries,
	}
	if err := writeJSON(w, http.StatusOK, h); err != nil {
		http.Error(w, `{"error":"failed to serialize response"}`, http.StatusInternalServerError)
		return
	}
}

func (s *Server) deleteHabit(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "habit_id")
	if userID == "" || habitID == "" {
		http.Error(w, `{"error":"user id and habit id are required"}`, http.StatusBadRequest)
		return
	}

	err := s.store.DeleteHabit(userID, habitID)
	if err != nil {
		http.Error(w, `{"error":"storage error"}`, http.StatusInternalServerError)
		return
	}

	habits, _ := s.store.ListHabitNames(userID)
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
