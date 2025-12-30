package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimitConfig struct {
	Enabled  bool
	Requests int
	Window   int
}

type rateLimitEntry struct {
	count     int
	expiresAt time.Time
}

type rateLimiter struct {
	entries map[string]*rateLimitEntry
	mu      sync.Mutex
	config  RateLimitConfig
}

func newRateLimiter(config RateLimitConfig) *rateLimiter {
	rl := &rateLimiter{
		entries: make(map[string]*rateLimitEntry),
		config:  config,
	}

	go rl.cleanup()

	return rl
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, entry := range rl.entries {
			if now.After(entry.expiresAt) {
				delete(rl.entries, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[key]

	if !exists || now.After(entry.expiresAt) {
		rl.entries[key] = &rateLimitEntry{
			count:     1,
			expiresAt: now.Add(time.Duration(rl.config.Window) * time.Second),
		}
		return true
	}

	entry.count++
	return entry.count <= rl.config.Requests
}

func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	if !config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	limiter := newRateLimiter(config)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !limiter.allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "リクエスト数が制限を超えました。しばらくしてからお試しください。",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
