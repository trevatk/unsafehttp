package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os/signal"
	"syscall"

	"net/http"
	_ "net/http/pprof"

	"github.com/trevatk/unsafehttp"
)

func health(w unsafehttp.ResponseWriter, r *unsafehttp.Request) {
	slog.Default().InfoContext(r.Context(), "health check received")
	w.SetStatus(unsafehttp.StatusOK)
}

func postCreateUser(w unsafehttp.ResponseWriter, r *unsafehttp.Request) {
	slog.Default().InfoContext(r.Context(), "create user")

	type x struct {
		Name string `json:"name"`
	}

	var newUser x
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		w.SetStatus(unsafehttp.StatusBadRequest)
		return
	}

	if newUser.Name == "" {
		w.SetStatus(unsafehttp.StatusBadRequest)
		return
	}

	type y struct {
		Name string `json:"name"`
	}

	user := &y{Name: newUser.Name}

	w.SetStatus(unsafehttp.StatusCreated)
	w.SetHeader("Content-Type", "application/json")

	json.NewEncoder(w).Encode(user) // nolint: errcheck,gosec
}

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		cancel()
		if r := recover(); r != nil {
			slog.Default().Error("panic", slog.Any("recovery", r))
		}
	}()

	mux := unsafehttp.NewServeMux()

	mux.Post("/api/v1/users", postCreateUser)
	mux.Get("/health", health)

	serverOpts := []unsafehttp.ServerOption{
		unsafehttp.WithAddr("localhost:8181"),
		unsafehttp.WithMux(mux),
	}

	s := unsafehttp.NewServer(serverOpts...)
	go http.ListenAndServe(":6060", nil)
	log.Fatal(s.Serve(ctx)) // nolint: errcheck,gosec
}
