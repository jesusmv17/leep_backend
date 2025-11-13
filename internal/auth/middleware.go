package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// UserClaims represents the JWT claims structure from Supabase Auth.
// These claims are embedded in every JWT token issued by Supabase.
type UserClaims struct {
	Sub   string `json:"sub"`   // User ID (UUID from Supabase Auth)
	Email string `json:"email"` // User email address
	Role  string `json:"role"`  // Token role (anon or authenticated)
	jwt.RegisteredClaims          // Standard JWT claims (iat, exp, iss, etc.)
}

// ContextKey is a custom type for storing user data in Gin context.
// Using a custom type prevents key collisions with other middleware.
type ContextKey string

// Context keys for storing authenticated user information.
// These values are set by the authentication middleware and can be
// retrieved in route handlers using the helper functions below.
const (
	UserIDKey  ContextKey = "user_id"  // Stores the authenticated user's UUID
	UserEmail  ContextKey = "user_email" // Stores the authenticated user's email
	UserRole   ContextKey = "user_role"  // Stores the user's role (for RBAC)
	UserToken  ContextKey = "user_token" // Stores the original JWT token
)

// RequireAuth is a Gin middleware that validates JWT tokens from Supabase.
// This middleware REQUIRES authentication - requests without valid tokens are rejected.
//
// Flow:
//   1. Extracts "Authorization: Bearer <token>" header
//   2. Validates JWT signature using Supabase JWT secret
//   3. Checks token expiration and claims
//   4. Stores user info (ID, email, token) in Gin context for use in handlers
//
// Usage:
//   Protected routes should use this middleware:
//   router.POST("/songs", auth.RequireAuth(), createSongHandler)
//
// Returns 401 Unauthorized if:
//   - Authorization header is missing
//   - Token format is invalid
//   - Token signature is invalid
//   - Token is expired
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if Authorization header is present
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		// Header should be in format: "Authorization: Bearer eyJhbGc..."
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate the JWT token
		// This checks signature, expiration, and decodes claims
		claims, err := ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": fmt.Sprintf("invalid token: %v", err),
			})
			c.Abort()
			return
		}

		// Store user information in Gin context for handler access
		// Handlers can retrieve these using GetUserID() or GetUserToken()
		c.Set(string(UserIDKey), claims.Sub)
		c.Set(string(UserEmail), claims.Email)
		c.Set(string(UserToken), tokenString)

		// Continue to the next middleware/handler
		c.Next()
	}
}

// OptionalAuth is a Gin middleware that validates JWT tokens if present,
// but allows the request to proceed even without authentication.
//
// This is useful for endpoints that have different behavior for authenticated
// vs. anonymous users (e.g., public song listing that shows published songs
// for anonymous users, but also shows unpublished songs for the song owner).
//
// Flow:
//   1. If no Authorization header, continue without setting user context
//   2. If header present, attempt to validate token
//   3. If valid, store user info in context
//   4. If invalid, silently ignore and continue (no error returned)
//
// Usage:
//   Public endpoints with optional auth:
//   router.GET("/songs", auth.OptionalAuth(), listSongsHandler)
//
// Handlers can check if user is authenticated using GetUserID():
//   userID, err := auth.GetUserID(c)
//   if err != nil {
//     // User is not authenticated, show public content only
//   } else {
//     // User is authenticated, can show personalized content
//   }
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if Authorization header exists
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header, proceed as anonymous user
			c.Next()
			return
		}

		// Attempt to parse and validate token
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString := parts[1]
			claims, err := ValidateToken(tokenString)
			if err == nil {
				// Valid token, store user info in context
				c.Set(string(UserIDKey), claims.Sub)
				c.Set(string(UserEmail), claims.Email)
				c.Set(string(UserToken), tokenString)
			}
			// If token is invalid, we silently ignore it and proceed
			// This allows the request to continue for public access
		}

		c.Next()
	}
}

// ValidateToken validates a Supabase JWT token and extracts the claims.
// This function performs complete JWT validation including:
//   - Signature verification using Supabase JWT secret
//   - Expiration check
//   - Claims structure validation
//
// Parameters:
//   - tokenString: The JWT token string (without "Bearer " prefix)
//
// Returns:
//   - *UserClaims: Decoded claims containing user ID, email, and role
//   - error: If validation fails (expired, invalid signature, malformed, etc.)
//
// Security notes:
//   - The JWT secret is read from SUPABASE_JWT_SECRET environment variable
//   - Secret can be base64-encoded or plain text
//   - Only HS256/HS384/HS512 signing methods are accepted
func ValidateToken(tokenString string) (*UserClaims, error) {
	// Load JWT secret from environment
	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("SUPABASE_JWT_SECRET not configured")
	}

	// Try to decode the secret as base64 (Supabase uses base64-encoded secrets)
	// If decoding fails, use the secret as-is (plain text)
	secretKey, err := base64.StdEncoding.DecodeString(jwtSecret)
	if err != nil {
		// Not base64-encoded, use as plain text
		secretKey = []byte(jwtSecret)
	}

	// Parse and validate the JWT token
	// The validation callback verifies the signing method and provides the secret key
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate that the token uses HMAC signing (HS256/HS384/HS512)
		// Reject tokens with unexpected signing methods to prevent attacks
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Check if token is valid (signature verified, not expired)
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract and type-assert the claims
	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// GetUserID extracts the authenticated user's ID from Gin context.
// This function should be called in handlers after RequireAuth() or OptionalAuth().
//
// Returns:
//   - string: The user's UUID from Supabase Auth
//   - error: If user is not authenticated (context key not set)
//
// Usage in handlers:
//   userID, err := auth.GetUserID(c)
//   if err != nil {
//     // User is not authenticated
//     return
//   }
//   // Use userID to query user's data
func GetUserID(c *gin.Context) (string, error) {
	userID, exists := c.Get(string(UserIDKey))
	if !exists {
		return "", fmt.Errorf("user not authenticated")
	}
	return userID.(string), nil
}

// GetUserToken extracts the JWT token from Gin context.
// This is useful when you need to forward the user's token to Supabase
// for requests that enforce Row Level Security (RLS).
//
// Returns:
//   - string: The original JWT token string
//   - error: If user token is not found in context
//
// Usage:
//   token, err := auth.GetUserToken(c)
//   if err != nil {
//     // User is not authenticated or token not stored
//     return
//   }
//   // Use token for Supabase requests
func GetUserToken(c *gin.Context) (string, error) {
	token, exists := c.Get(string(UserToken))
	if !exists {
		return "", fmt.Errorf("user token not found")
	}
	return token.(string), nil
}

// MustGetUserID extracts the user ID or panics if not authenticated.
// WARNING: Only use this in handlers that are protected by RequireAuth() middleware.
// This function will panic if the user is not authenticated.
//
// Returns:
//   - string: The user's UUID
//
// Panics if user is not authenticated.
//
// Usage (only after RequireAuth):
//   userID := auth.MustGetUserID(c)
//   // userID is guaranteed to be valid here
func MustGetUserID(c *gin.Context) string {
	userID, err := GetUserID(c)
	if err != nil {
		panic(err)
	}
	return userID
}

// GetUserRole fetches the user's role from Supabase profiles table
func GetUserRole(ctx context.Context, userID string, supabaseClient interface{}) (string, error) {
	// This will be implemented to query the profiles table
	// For now, return a default
	return "fan", nil
}

// RequireRole middleware checks if user has required role
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			c.Abort()
			return
		}

		// TODO: Fetch user role from Supabase profiles table
		// For now, we'll skip role validation
		// In production, query: SELECT role FROM profiles WHERE id = userID

		_ = userID // Use userID when implementing role check

		c.Next()
	}
}

// ProfileResponse represents a user profile from Supabase
type ProfileResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	CreatedAt   string `json:"created_at"`
}

// ParseProfile parses a profile JSON response
func ParseProfile(data []byte) (*ProfileResponse, error) {
	var profiles []ProfileResponse
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return nil, fmt.Errorf("profile not found")
	}
	return &profiles[0], nil
}
