package main

import (
	"context"
	"log"
	"rate-limiter/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = ctx
	err := config.LoadConfig("./config.yml")
	if err != nil {
		log.Fatalf("error while loading configuration", "error", err.Error())
	}
	cfg := config.GetConfig()
	_ = cfg
}
