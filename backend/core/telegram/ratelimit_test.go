package telegram

import (
	"testing"
	"time"
)

func TestUserRateLimiter(t *testing.T) {
	limiter := NewUserRateLimiter(2, 50*time.Millisecond)
	userID := int64(112233)

	// 1st request allowed
	allow, _ := limiter.Allow(userID)
	if !allow {
		t.Errorf("1st request should be allowed")
	}

	// 2nd request allowed
	allow, _ = limiter.Allow(userID)
	if !allow {
		t.Errorf("2nd request should be allowed")
	}

	// 3rd request rejected
	allow, retryAfter := limiter.Allow(userID)
	if allow {
		t.Errorf("3rd request should be rejected")
	}
	if retryAfter <= 0 {
		t.Errorf("expected retryAfter > 0, got %v", retryAfter)
	}

	// Another user should still be allowed (isolation)
	otherUser := int64(445566)
	allowOther, _ := limiter.Allow(otherUser)
	if !allowOther {
		t.Errorf("other user should be allowed")
	}

	// Wait for window to slide
	time.Sleep(60 * time.Millisecond)
	allowAfterWait, _ := limiter.Allow(userID)
	if !allowAfterWait {
		t.Errorf("request after window expiry should be allowed")
	}
}
