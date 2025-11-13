// Package middleware provides HTTP middleware for the Leep Audio backend.
// This package includes:
//   - Logging: Structured request/response logging with user tracking
//   - Rate limiting: IP-based rate limiting to prevent abuse
//   - CORS: Cross-Origin Resource Sharing configuration for frontend integration
//
// These middleware components work together to provide observability,
// security, and proper API access control.
package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jesusmv17/leep_backend/internal/auth"
)

// Logger is a Gin middleware that logs all HTTP requests with detailed information.
// Log format includes:
//   - Timestamp (RFC3339)
//   - HTTP method (GET, POST, etc.)
//   - Request path
//   - Status code
//   - Request latency (duration)
//   - User ID (if authenticated)
//
// This provides complete audit trail and helps with debugging and monitoring.
// In production, consider sending logs to a centralized logging service.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Get user ID if authenticated
		userID, _ := auth.GetUserID(c)

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// Log format: [timestamp] method path status_code latency user_id
		if userID != "" {
			log.Printf("[%s] %s %s %d %v user=%s",
				time.Now().Format(time.RFC3339),
				method,
				path,
				statusCode,
				latency,
				userID,
			)
		} else {
			log.Printf("[%s] %s %s %d %v",
				time.Now().Format(time.RFC3339),
				method,
				path,
				statusCode,
				latency,
			)
		}
	}
}

