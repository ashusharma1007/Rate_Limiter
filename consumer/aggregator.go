package consumer

import (
	"context"
	"fmt"
	"net"
	"rate-limiter/consumer/store"
	"rate-limiter/models"
)

type Aggregator struct {
	store *store.RedisStore
}

func NewAggregator(store *store.RedisStore) *Aggregator {
	return &Aggregator{store: store}
}

// Process aggregates a RateLimitEvent into Redis — increments counters,
// tracks userIDs per IP, records endpoint hits, and resets the 24h TTL
func (a *Aggregator) Process(ctx context.Context, event models.RateLimitEvent) error {
	ip, _, _ := net.SplitHostPort(event.ClientIP)
	userID := event.UserID

	if err := a.store.IncrTotal(ctx, ip); err != nil {
		return fmt.Errorf("failed to increment total: %v", err)
	}

	if !event.Allowed {
		if err := a.store.IncrBlocked(ctx, ip); err != nil {
			return fmt.Errorf("failed to increment blocked: %v", err)
		}
	}

	if err := a.store.AddUserID(ctx, ip, userID); err != nil {
		return fmt.Errorf("failed to add userID: %v", err)
	}

	if err := a.store.IncrEndpoint(ctx, ip, event.Endpoint); err != nil {
		return fmt.Errorf("failed to increment endpoint: %v", err)
	}

	if err := a.store.SetTTL(ctx, ip); err != nil {
		return fmt.Errorf("failed to set ttl: %v", err)
	}

	return nil
}
