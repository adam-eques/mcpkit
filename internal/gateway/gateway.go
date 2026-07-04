// Package gateway exposes an MCP server over HTTP. A client POSTs a single
// JSON-RPC request to /rpc and receives the JSON-RPC response synchronously.
// /healthz and /metrics support operational monitoring.
package gateway

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/adam-eques/mcpkit/internal/log"
	"github.com/adam-eques/mcpkit/server"
)

// Handler builds the HTTP handler for srv.
func Handler(srv *server.Server, logger *log.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc", rpcHandler(srv, logger))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, srv.Metrics().Snapshot())
	})
	return withLogging(mux, logger)
}

const maxBodyBytes = 8 << 20 // 8 MiB

func rpcHandler(srv *server.Server, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes))
		if err != nil {
			http.Error(w, "read error", http.StatusBadRequest)
			return
		}
		resp, err := srv.HandleMessage(r.Context(), body)
		if err != nil {
			logger.Error("gateway handle", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if resp == nil {
			// Notification: acknowledge with 204 and no body.
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(resp); err != nil {
			logger.Error("gateway write", "err", err)
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// statusRecorder captures the response status for access logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func withLogging(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logger.Debug("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"durationMs", time.Since(start).Milliseconds())
	})
}
