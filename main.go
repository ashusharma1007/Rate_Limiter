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
	"rate-limiter/consumer"
	"rate-limiter/consumer/store"
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

	redisStore := store.NewRedisStore("localhost:6379")

	aggregator := consumer.NewAggregator(redisStore)

	c, err := consumer.NewConsumer(cfg.KafkaBrokerAddress, "rate-limiter-consumer", cfg.KafkaTopic, aggregator)
	if err != nil {
		log.Fatalf("error creating consumer: %v", err)
	}

	go func() {
		log.Println("consumer starting")
		if err := c.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Fatalf("consumer error: %v", err)
		}
	}()

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
