
package metrics

import (
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Start(addr string, log *slog.Logger) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		log.Info("metrics http starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Error("metrics http error", "err", err)
		}
	}()
	return srv
}
