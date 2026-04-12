package main

import (
	"log/slog"
	"net/http"
	"os"

	"api-gateway/internal/config"

	"github.com/go-chi/chi/v5"
)

func main() {
	app, err := config.LoadApp()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = app.Port
	}
	if port == "" {
		port = "8080"
	}

	slog.Info("gateway config",
		"routes", len(app.Routes),
		"default_timeout", app.DefaultTimeout,
	)

	r := chi.NewRouter()
	r.Get("/health", health)
	r.Get("/", home)

	slog.Info("listening", "addr", ":"+port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func home(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("api-gateway"))
}
