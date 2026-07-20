package ratelimit

import (
	"sync"
	"time"
)

type Limiter interface {
	Allow(clientID string) bool
}

type TokenBucketLimiter struct {
	mu      sync.Mutex
	clients map[string]*bucket
	rps     int
	window  time.Duration
}

type bucket struct {
	tokens     int
	lastrefill time.Time
}

func NewLimiter(rps int, window time.Duration) *TokenBucketLimiter {
	return &TokenBucketLimiter{rps: rps, window: window, clients: make(map[string]*bucket)}
}

func (l *TokenBucketLimiter) Allow(clientID string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	//get bucket for clientID from l.clients
	//if it doesn't exist, create one with full tokens and current time
	b, exists := l.clients[clientID]
	if !exists {
		l.clients[clientID] = &bucket{
			tokens:     l.rps,
			lastrefill: time.Now(),
		}
		b = l.clients[clientID]
	}

	//chek if window has passed since b.lastRefill
	//if yes: reset b.tokens = l.rps, update b.lastRefill = now
	if time.Since(b.lastrefill) > l.window {
		b.tokens = l.rps
		b.lastrefill = time.Now()
	}

	// if b.tokens > 0: decrement tokens, return true
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false

}
