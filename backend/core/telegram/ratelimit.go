package telegram

import (
	"sync"
	"time"
)

// UserRateLimiter tracks per-user request counts over a sliding window.
type UserRateLimiter struct {
	mu          sync.Mutex
	limit       int
	window      time.Duration
	userHistory map[int64][]time.Time
}

// NewUserRateLimiter creates a new rate limiter allowing maxRequests per window per user.
// If maxRequests <= 0, defaults to 6 per minute.
func NewUserRateLimiter(maxRequests int, window time.Duration) *UserRateLimiter {
	if maxRequests <= 0 {
		maxRequests = 6
	}
	if window <= 0 {
		window = time.Minute
	}
	return &UserRateLimiter{
		limit:       maxRequests,
		window:      window,
		userHistory: make(map[int64][]time.Time),
	}
}

// Allow returns true if the user has not exceeded their quota, or false with retryAfter duration.
func (r *UserRateLimiter) Allow(userID int64) (bool, time.Duration) {
	if r == nil || r.limit <= 0 {
		return true, 0
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	// Filter timestamps older than cutoff
	timestamps := r.userHistory[userID]
	var valid []time.Time
	for _, ts := range timestamps {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}

	if len(valid) >= r.limit {
		oldest := valid[0]
		retryAfter := oldest.Add(r.window).Sub(now)
		if retryAfter <= 0 {
			retryAfter = time.Second
		}
		r.userHistory[userID] = valid
		return false, retryAfter
	}

	valid = append(valid, now)
	r.userHistory[userID] = valid
	return true, 0
}
