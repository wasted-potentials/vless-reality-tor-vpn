
package xray

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Runner struct {
	Binary     string
	ConfigPath string

	cmd    *exec.Cmd
	cancel context.CancelFunc

	stdout io.ReadCloser
	stderr io.ReadCloser

	startedAt time.Time
}

var (
	xrayStarts = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "vpnd",
		Name:      "xray_starts_total",
		Help:      "Number of times xray process was started",
	})
	xrayRestarts = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "vpnd",
		Name:      "xray_restarts_total",
		Help:      "Number of process restarts",
	})
	xrayUptime = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "vpnd",
		Name:      "xray_uptime_seconds",
		Help:      "Uptime of xray process in seconds",
	})
	xrayUp = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "vpnd",
		Name:      "xray_up",
		Help:      "Is xray process running (1) or not (0)",
	})
)

func (r *Runner) Start(ctx context.Context, args ...string) error {
	if r.cmd != nil {
		return errors.New("xray already running")
	}
	ctx, cancel := context.WithCancel(ctx)
	r.cancel = cancel

	fullArgs := append([]string{"run", "-config", r.ConfigPath}, args...)
	cmd := exec.CommandContext(ctx, r.Binary, fullArgs...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	r.stdout = stdout
	r.stderr = stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting xray: %w", err)
	}
	r.cmd = cmd
	r.startedAt = time.Now()
	xrayStarts.Inc()
	xrayUp.Set(1)

	go pipe("xray-stdout", stdout)
	go pipe("xray-stderr", stderr)

	// Wait and update metrics
	go func() {
		err := cmd.Wait()
		if err != nil {
			xrayUp.Set(0)
		}
	}()

	// Handle SIGTERM/SIGINT to stop child properly
	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT)
		defer signal.Stop(sigc)
		select {
		case <-ctx.Done():
			return
		case <-sigc:
			_ = r.Stop()
		}
	}()

	// Update uptime gauge
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				xrayUptime.Set(time.Since(r.startedAt).Seconds())
			}
		}
	}()

	return nil
}

func (r *Runner) Stop() error {
	if r.cmd == nil {
		return nil
	}
	if r.cancel != nil {
		r.cancel()
	}
	// Give it a moment to exit
	done := make(chan error, 1)
	go func() { done <- r.cmd.Wait() }()
	select {
	case <-time.After(3 * time.Second):
		_ = r.cmd.Process.Kill()
	case <-done:
	}
	r.cmd = nil
	xrayUp.Set(0)
	return nil
}

func (r *Runner) Restart(ctx context.Context, args ...string) error {
	_ = r.Stop()
	xrayRestarts.Inc()
	return r.Start(ctx, args...)
}

func pipe(name string, r io.Reader) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		// Print raw lines; in production you might integrate with slog/zap
		fmt.Printf("[%s] %s\n", name, sc.Text())
	}
}
