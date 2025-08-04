package main

import (
	"context"
	"encoding/json"
	"fmt"
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

	r := unsafehttp.NewRouter()
	r.Group("api/1.0", func(rr unsafehttp.Router) {
		rr.Group("/users", func(users unsafehttp.Router) {
			users.Post("/", postCreateUser)
		})
	})
	r.Get("health", health)

	r.Walk(func(s1, s2 string, hf unsafehttp.HandlerFunc) {
		fmt.Printf("pattern: %s, method: %s\n", s1, s2)
	})

	serverOpts := []unsafehttp.ServerOption{
		unsafehttp.WithAddr("localhost:8181"),
		unsafehttp.WithRouter(r),
	}

	s := unsafehttp.NewServer(serverOpts...)
	go http.ListenAndServe(":6060", nil)
	log.Fatal(s.Serve(ctx)) // nolint: errcheck,gosec
}
