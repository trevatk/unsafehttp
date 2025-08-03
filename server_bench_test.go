package unsafehttp

import (
	"context"
	"crypto/sha256"
	"net/http"
	"testing"
	"time"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

func BenchmarkServerAllocationsGetUnsafeHTTP(b *testing.B) {
	mux := NewServeMux()
	mux.Get("/bench", func(w ResponseWriter, r *Request) {
		for i := 0; i < 1000; i++ {
			sha256.Sum256([]byte("hello world"))
		}
		data := make([]byte, 1024)
		w.SetStatus(StatusOK)
		w.Write(data)
	})
	opts := []ServerOption{
		WithMux(mux),
		WithMaxHeaderSize(16),
		WithAddr("localhost:8082"),
	}
	s := NewServer(opts...)

	go s.Serve(context.Background())
	time.Sleep(time.Second)
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			rr, err := client.Get("http://localhost:8082/bench")
			if err != nil {
				b.Fatal(err)
			}
			rr.Body.Close()
		}
	})
}

type x struct {
	Name string `json:"name"`
}

// func BenchmarkServerAllocationsPostUnsafeHTTP(b *testing.B) {
// 	mux := NewServeMux()
// 	mux.Post("/api/v1/user", func(w ResponseWriter, r *Request) {

// 		var newUser x
// 		if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
// 			w.SetStatus(StatusBadRequest)
// 			return
// 		}

// 		if newUser.Name == "" {
// 			w.SetStatus(StatusBadRequest)
// 			return
// 		}

// 		user := &x{Name: newUser.Name}

// 		w.SetStatus(StatusCreated)
// 		w.SetHeader("Content-Type", "application/json")

// 		json.NewEncoder(w).Encode(user) // nolint: errcheck,gosec
// 	})
// 	opts := []ServerOption{
// 		WithMux(mux),
// 		WithMaxHeaderSize(16),
// 		WithAddr("localhost:8082"),
// 	}
// 	s := NewServer(opts...)

// 	go s.Serve(context.Background())
// 	time.Sleep(time.Second)
// 	b.ResetTimer()

// 	b.RunParallel(func(p *testing.PB) {
// 		for p.Next() {

// 			body, err := json.Marshal(&x{Name: "benchmark"})
// 			if err != nil {
// 				b.Fatal(err)
// 			}

// 			rr, err := client.Post("http://localhost:8082/api/v1/user", "application/json", bytes.NewBuffer(body))
// 			if err != nil {
// 				b.Fatal(err)
// 			}
// 			rr.Body.Close()
// 		}
// 	})
// }

func BenchmarkServerAllocationsGetHTTP(b *testing.B) {
	mux := http.NewServeMux()
	mux.HandleFunc("/bench", func(w http.ResponseWriter, r *http.Request) {
		data := make([]byte, 1024)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})
	s := &http.Server{Addr: ":8083", Handler: mux}
	go s.ListenAndServe()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			rr, err := client.Get("http://localhost:8082/bench")
			if err != nil {
				b.Fatal(err)
			}
			rr.Body.Close()
		}
	})
}

// func BenchmarkServerAllocationsPostHTTP(b *testing.B) {
// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/api/v1/user", func(w http.ResponseWriter, r *http.Request) {

// 		var newUser x
// 		if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
// 			w.WriteHeader(http.StatusBadRequest)
// 			return
// 		}

// 		if newUser.Name == "" {
// 			w.WriteHeader(http.StatusBadRequest)
// 			return
// 		}

// 		user := &x{Name: newUser.Name}

// 		w.WriteHeader(http.StatusCreated)
// 		w.Header().Add("Content-Type", "application/json")

// 		json.NewEncoder(w).Encode(user) // nolint: errcheck,gosec
// 	})

// 	s := &http.Server{Addr: ":8083", Handler: mux}
// 	go s.ListenAndServe()
// 	b.ResetTimer()

// 	time.Sleep(time.Second)
// 	b.ResetTimer()

// 	b.RunParallel(func(p *testing.PB) {
// 		for p.Next() {

// 			body, err := json.Marshal(&x{Name: "benchmark"})
// 			if err != nil {
// 				b.Fatal(err)
// 			}

// 			rr, err := client.Post("http://localhost:8083/api/v1/user", "application/json", bytes.NewBuffer(body))
// 			if err != nil {
// 				b.Fatal(err)
// 			}
// 			rr.Body.Close()
// 		}
// 	})
// }
