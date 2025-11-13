package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// rateLimiter is a simple in-memory rate limiter that tracks request counts per client IP.
// Note: This is suitable for single-instance deployments. For production environments
// with multiple instances, consider using a distributed rate limiter with Redis.
type rateLimiter struct {
	mu      sync.Mutex               // Protects concurrent access to clients map
	clients map[string]*clientInfo   // Maps client IP to their request info
}

// clientInfo stores the request count and reset time for a single client.
type clientInfo struct {
	count      int        // Number of requests made in current window
	lastReset  time.Time  // When the current window started
}

var limiter = &rateLimiter{
	clients: make(map[string]*clientInfo),
}

// RateLimit is a Gin middleware that limits the number of requests per client IP.
// This helps prevent abuse and ensures fair resource allocation across all users.
//
// Parameters:
//   - maxRequests: Maximum number of requests allowed per window
//   - window: Time window for rate limiting (e.g., 1 minute)
//
// Example usage:
//   router.Use(middleware.RateLimit(100, time.Minute)) // 100 req/min
//
// Returns 429 Too Many Requests if limit is exceeded, with a retry_after duration.
//
// Note: This uses in-memory storage and resets on server restart.
// For production with multiple instances, use Redis-backed rate limiting.
func RateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	// Start background cleanup goroutine to remove stale entries
	// This prevents memory leaks from inactive clients
	go func() {
		for {
			time.Sleep(window)
			limiter.mu.Lock()
			for ip, info := range limiter.clients {
				if time.Since(info.lastReset) > window {
					delete(limiter.clients, ip)
				}
			}
			limiter.mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		limiter.mu.Lock()
		defer limiter.mu.Unlock()

		info, exists := limiter.clients[clientIP]
		if !exists {
			limiter.clients[clientIP] = &clientInfo{
				count:     1,
				lastReset: time.Now(),
			}
			c.Next()
			return
		}

		// Reset counter if window has passed
		if time.Since(info.lastReset) > window {
			info.count = 1
			info.lastReset = time.Now()
			c.Next()
			return
		}

		// Check if limit exceeded
		if info.count >= maxRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"retry_after": window - time.Since(info.lastReset),
			})
			c.Abort()
			return
		}

		info.count++
		c.Next()
	}
}
