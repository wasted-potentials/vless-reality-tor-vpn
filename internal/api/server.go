package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	cfg "github.com/vasyza/vless-reality-tor-vpn/internal/config"
	"github.com/vasyza/vless-reality-tor-vpn/internal/users"
	"github.com/vasyza/vless-reality-tor-vpn/internal/xray"
)

type Server struct {
	cfg     cfg.Config
	log     *slog.Logger
	users   *users.Store
	runner  *xray.Runner
	authTok string

	srv     *http.Server
	metrics *http.Server
}

func NewServer(c cfg.Config, log *slog.Logger, userStore *users.Store, run *xray.Runner) *Server {
	return &Server{
		cfg:     c,
		log:     log,
		users:   userStore,
		runner:  run,
		authTok: c.Admin.AuthToken,
	}
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /readyz", s.handleReady)
	mux.HandleFunc("POST /api/users", s.auth(s.handleAddUser))
	mux.HandleFunc("DELETE /api/users/{uuid}", s.auth(s.handleDelUser))
	mux.HandleFunc("GET /api/users", s.auth(s.handleListUsers))
	mux.HandleFunc("POST /api/reload", s.auth(s.handleReload))

	s.srv = &http.Server{
		Addr:              s.cfg.Admin.Listen,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		s.log.Info("admin http starting", "addr", s.cfg.Admin.Listen)
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Error("admin http error", "err", err)
		}
	}()

	// metrics
	mm := http.NewServeMux()
	mm.Handle("/metrics", promhttp.Handler())
	s.metrics = &http.Server{Addr: s.cfg.Admin.MetricsListen, Handler: mm}
	go func() {
		s.log.Info("metrics http starting", "addr", s.cfg.Admin.MetricsListen)
		if err := s.metrics.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Error("metrics http error", "err", err)
		}
	}()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if s.metrics != nil {
		_ = s.metrics.Shutdown(ctx)
	}
	if s.srv != nil {
		_ = s.srv.Shutdown(ctx)
	}
	return nil
}

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.authTok == "" {
			next.ServeHTTP(w, r)
			return
		}
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}
		if strings.TrimPrefix(auth, "Bearer ") != s.authTok {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	// naive: consider ready if xray process is alive
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ready":true}`))
}

func (s *Server) handleAddUser(w http.ResponseWriter, r *http.Request) {
	type req struct {
		UUID string `json:"uuid,omitempty"`
	}
	var rr req
	_ = json.NewDecoder(r.Body).Decode(&rr)
	var id uuid.UUID
	var err error
	if rr.UUID == "" {
		id = uuid.New()
	} else {
		id, err = uuid.Parse(rr.UUID)
		if err != nil {
			http.Error(w, "bad uuid", http.StatusBadRequest)
			return
		}
	}
	u, err := s.users.Add(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Write config and restart xray
	if err := s.rebuildAndReload(r.Context()); err != nil {
		http.Error(w, fmt.Sprintf("added but reload failed: %v", err), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(u)
}

func (s *Server) handleDelUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("uuid")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "bad uuid", http.StatusBadRequest)
		return
	}
	if err := s.users.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.rebuildAndReload(r.Context()); err != nil {
		http.Error(w, fmt.Sprintf("deleted but reload failed: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	list := s.users.List()
	_ = json.NewEncoder(w).Encode(list)
}

func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	if err := s.rebuildAndReload(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) rebuildAndReload(ctx context.Context) error {
	// Build config from users
	uu := s.users.List()
	userIDs := make([]uuid.UUID, 0, len(uu))
	for _, u := range uu {
		id, _ := uuid.Parse(u.UUID)
		userIDs = append(userIDs, id)
	}
	if err := xray.ValidateRealityFields(s.cfg); err != nil {
		return err
	}
	root, err := xray.MakeConfig(s.cfg, userIDs)
	if err != nil {
		return err
	}
	if err := xray.Write(s.cfg.Xray.ConfigPath, root); err != nil {
		return err
	}
	// restart xray
	return s.runner.Restart(ctx)
}
