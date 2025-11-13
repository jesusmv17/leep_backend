package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORS is a Gin middleware that enables Cross-Origin Resource Sharing.
// This allows the frontend (hosted on Vercel) to make API requests to this backend.
//
// Current configuration:
//   - Allows all origins (*) - should be restricted to specific domain in production
//   - Allows credentials (cookies, authorization headers)
//   - Allows all common headers and methods
//
// Security note: In production, replace "*" with your actual frontend domain:
//   c.Writer.Header().Set("Access-Control-Allow-Origin", "https://leepaudio.vercel.app")
//
// This middleware handles both actual requests and preflight OPTIONS requests.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // In production, set specific origin
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
