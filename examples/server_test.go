package examples

import (
	"context"
	"encoding/json"
	"log/slog"
	"os/signal"
	"syscall"

	http "github.com/trevatk/unsafehttp"
)

func health(w http.ResponseWriter, r *http.Request) {
	slog.Default().InfoContext(r.Context(), "health check received")
	w.SetStatus(http.StatusOK)
}

func postCreateUser(w http.ResponseWriter, r *http.Request) {
	slog.Default().InfoContext(r.Context(), "create user")

	type x struct {
		Name string `json:"name"`
	}

	var newUser x
	if err := json.Unmarshal(r.Body, &newUser); err != nil {
		w.SetStatus(http.StatusBadRequest)
		return
	}

	if newUser.Name == "" {
		w.SetStatus(http.StatusBadRequest)
		return
	}

	w.SetStatus(http.StatusCreated)
}

func Example_server() {

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		cancel()
		if r := recover(); r != nil {
			slog.Default().Error("panic", slog.Any("recovery", r))
		}
	}()

	mux := http.NewServeMux()

	mux.Post("/api/v1/users", postCreateUser)
	mux.Get("/health", health)

	serverOpts := []http.ServerOption{
		http.WithAddr("localhost:8181"),
		http.WithMux(mux),
	}

	s := http.NewServer(serverOpts...)

	s.Serve(ctx)
}
