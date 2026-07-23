package server

import (
	"context"
	"net/http"
	"rate-limiter/auth"
	"rate-limiter/kafka"
	"rate-limiter/middleware"
	"rate-limiter/ratelimit"
)

type Server struct {
	http *http.Server
}

func New(port string, jwtSecret string, limiter *ratelimit.TokenBucketLimiter, producer *kafka.Producer) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.Handle("/api/", auth.Middleware(jwtSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})))

	return &Server{
		http: &http.Server{
			Addr:    ":" + port,
			Handler: middleware.RateLimit(limiter, producer)(mux),
		},
	}
}

func (s *Server) Start() error {
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
