package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"rate-limiter/config"
	"rate-limiter/kafka"
	"rate-limiter/ratelimit"
	"rate-limiter/server"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	err := config.LoadConfig("./config.yml")
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	cfg := config.GetConfig()
	producer, er := kafka.NewProducer(cfg.KafkaBrokerAddress, cfg.KafkaTopic)
	if er != nil {
		log.Fatalf("error creating kafka producer: %v", er)
	}

	limiter := ratelimit.NewLimiter(cfg.ReqPerSec, cfg.RateLimitWindow)

	server := server.New(cfg.Port, cfg.JWTSecret, limiter, producer)

	go func() {
		log.Printf("server starting on port %s", cfg.Port)
		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	producer.Flush(3000)
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
