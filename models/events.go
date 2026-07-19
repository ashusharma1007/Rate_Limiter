package models

import "time"

type RateLimitEvent struct {
	ClientIP   string    `json:"client_ip"`
	Endpoint   string    `json:"endpoint"`
	UserID     string    `json:"user_id"`
	Allowed    bool      `json:"allowed"`
	Timestamp  time.Time `json:"timestamp"`
	StatusCode int       `json:"status_code"`
}
