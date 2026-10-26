package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

func main() {
	upstream, err := url.Parse(env("UPSTREAM_URL", "http://127.0.0.1:8081"))
	if err != nil {
		slog.Error("invalid UPSTREAM_URL", "error", err)
		os.Exit(1)
	}

	proxy := httputil.NewSingleHostReverseProxy(upstream)
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, err error) {
		slog.Warn("model server unavailable", "error", err)
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"error": map[string]string{"message": "local model server is unavailable", "type": "server_error"},
		})
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler(upstream))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"name":      "local-ai",
			"endpoints": []string{"/v1/models", "/v1/chat/completions", "/v1/completions", "/v1/embeddings"},
		})
	})
	mux.Handle("/v1/", proxy)

	server := &http.Server{
		Addr: env("LISTEN_ADDR", ":8080"), Handler: mux,
		ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second,
	}
	slog.Info("starting OpenAI-compatible local model gateway", "address", server.Addr, "upstream", upstream.String())
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func healthHandler(upstream *url.URL) http.HandlerFunc {
	client := &http.Client{Timeout: 2 * time.Second}
	return func(w http.ResponseWriter, r *http.Request) {
		healthURL := upstream.ResolveReference(&url.URL{Path: "/health"})
		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, healthURL.String(), nil)
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "loading"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func env(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
