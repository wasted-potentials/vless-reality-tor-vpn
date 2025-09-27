package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/google/uuid"

	"github.com/vasyza/vless-reality-tor-vpn/internal/api"
	cfg "github.com/vasyza/vless-reality-tor-vpn/internal/config"
	"github.com/vasyza/vless-reality-tor-vpn/internal/users"
	"github.com/vasyza/vless-reality-tor-vpn/internal/xray"
)

var (
	flagConfig = flag.String("config", "", "Path to YAML config (required)")
)

func main() {
	flag.Parse()
	if *flagConfig == "" {
		fmt.Println("Usage: vpnd -config /etc/vpnd/config.yaml")
		os.Exit(2)
	}

	conf, err := cfg.FromFile(*flagConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "reading config: %v\n", err)
		os.Exit(1)
	}
	if err := conf.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "invalid config: %v\n", err)
		os.Exit(1)
	}

	// logger
	var h slog.Handler
	if conf.Log.JSON {
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level(conf.Log.Level)})
	} else {
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level(conf.Log.Level)})
	}
	log := slog.New(h)

	// Prepare users store
	_ = os.MkdirAll(filepath.Dir(conf.Users.Path), 0755)
	userStore, err := users.NewStore(conf.Users.Path)
	if err != nil {
		log.Error("user store", "err", err)
		os.Exit(1)
	}
	// Ensure at least one user exists (opinionated)
	if len(userStore.List()) == 0 {
		_, _ = userStore.Add(uuid.New())
		log.Info("no users found, created one", "users", len(userStore.List()))
	}

	// Initial config write
	uus := userStore.List()
	var ids []uuid.UUID
	for _, u := range uus {
		id, _ := uuid.Parse(u.UUID)
		ids = append(ids, id)
	}
	if err := xray.ValidateRealityFields(conf); err != nil {
		log.Error("reality fields", "err", err)
		os.Exit(1)
	}
	root, err := xray.MakeConfig(conf, ids)
	if err != nil {
		log.Error("make xray config", "err", err)
		os.Exit(1)
	}
	if err := xray.Write(conf.Xray.ConfigPath, root); err != nil {
		log.Error("write xray config", "err", err)
		os.Exit(1)
	}

	// Start xray
	run := &xray.Runner{Binary: conf.Xray.Binary, ConfigPath: conf.Xray.ConfigPath}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := run.Start(ctx); err != nil {
		log.Error("xray start", "err", err)
		os.Exit(1)
	}

	// Start admin/api server
	srv := api.NewServer(conf, log, userStore, run)
	if err := srv.Start(ctx); err != nil {
		log.Error("admin start", "err", err)
		os.Exit(1)
	}

	// Wait for signal
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT)
	<-sigc
	log.Info("shutting down")
	_ = srv.Stop(context.Background())
	_ = run.Stop()
}

func level(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
