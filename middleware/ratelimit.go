package middleware

import (
	"net"
	"net/http"
	"rate-limiter/auth"
	"rate-limiter/kafka"
	"rate-limiter/models"
	"rate-limiter/ratelimit"
	"time"
)

func RateLimit(limiter *ratelimit.TokenBucketLimiter, producer *kafka.Producer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID, _, _ := net.SplitHostPort(r.RemoteAddr)
			if claims, ok := auth.GetClaims(r.Context()); ok {
				clientID = claims.UserID
			}
			allowed := limiter.Allow(clientID)
			statusCode := http.StatusOK
			if !allowed {
				statusCode = http.StatusTooManyRequests
			}

			event := models.RateLimitEvent{
				ClientIP:   r.RemoteAddr,
				Endpoint:   r.URL.Path,
				UserID:     clientID,
				Allowed:    allowed,
				Timestamp:  time.Now(),
				StatusCode: statusCode,
			}
			producer.Publish(event)
			if !allowed {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}