package unsafehttp

import (
	"testing"
)

func TestMux(t *testing.T) {
	mux := NewServeMux()
	health := func(w ResponseWriter, r *Request) {
		w.SetStatus(StatusOK)
		w.Write([]byte("OK")) // nolint: errcheck,gosec
	}
	t.Run("GET", func(t *testing.T) {
		mux.Get("/health", health)
		if _, ok := mux.matchRoute(&Request{
			Version: HTTP1,
			Method:  MethodGet,
			Path:    "/health",
		}); !ok {
			t.Fatal("unable to find health handler")
		}
	})
	t.Run("POST", func(t *testing.T) {
		mux.Post("/health", health)
		if _, ok := mux.matchRoute(&Request{
			Version: HTTP1,
			Method:  MethodPost,
			Path:    "/health",
		}); !ok {
			t.Fatal("unable to find health handler")
		}
	})
	t.Run("PATCH", func(t *testing.T) {
		mux.Patch("/health", health)
		if _, ok := mux.matchRoute(&Request{
			Version: HTTP1,
			Method:  MethodPatch,
			Path:    "/health",
		}); !ok {
			t.Fatal("unable to find health handler")
		}
	})
	t.Run("PUT", func(t *testing.T) {
		mux.Put("/health", health)
		if _, ok := mux.matchRoute(&Request{
			Version: HTTP1,
			Method:  MethodPut,
			Path:    "/health",
		}); !ok {
			t.Fatal("unable to find health handler")
		}
	})
	t.Run("DELETE", func(t *testing.T) {
		mux.Delete("/health", health)
		if _, ok := mux.matchRoute(&Request{
			Version: HTTP1,
			Method:  MethodDelete,
			Path:    "/health",
		}); !ok {
			t.Fatal("unable to find health handler")
		}
	})
}
