// Package auth provides authentication and authorization functionality for the Leep Audio backend.
// This package handles user authentication via Supabase Auth, including:
//   - User registration (signup)
//   - User login (authentication)
//   - JWT token validation
//   - User profile retrieval
//   - Session management (logout)
//
// All authentication is handled through Supabase Auth API, with JWT tokens
// being issued and validated using Supabase's built-in authentication system.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jesusmv17/leep_backend/internal/supabase"
)

// Handler manages auth endpoints
type Handler struct {
	supabaseClient *supabase.Client
}

// NewHandler creates a new auth handler
func NewHandler(supabaseClient *supabase.Client) *Handler {
	return &Handler{
		supabaseClient: supabaseClient,
	}
}

// SignupRequest represents the signup request body
type SignupRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	DisplayName string `json:"display_name"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents the auth response from Supabase
type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	TokenType    string      `json:"token_type"`
	ExpiresIn    int         `json:"expires_in"`
	RefreshToken string      `json:"refresh_token"`
	User         interface{} `json:"user"`
}

// Signup handles user registration
// POST /auth/signup
func (h *Handler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
			"details": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Call Supabase Auth API for signup
	resp, err := h.supabaseClient.Post(ctx, "/auth/v1/signup", map[string]interface{}{
		"email":    req.Email,
		"password": req.Password,
		"data": map[string]interface{}{
			"display_name": req.DisplayName,
		},
	}, "")

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create account",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "signup failed",
			"details": string(body),
		})
		return
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse response",
		})
		return
	}

	c.JSON(http.StatusCreated, authResp)
}

// Login handles user authentication
// POST /auth/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
			"details": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Call Supabase Auth API for login with password
	resp, err := h.supabaseClient.Post(ctx, "/auth/v1/token?grant_type=password", map[string]interface{}{
		"email":    req.Email,
		"password": req.Password,
	}, "")

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "login failed",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "invalid credentials",
			"details": string(body),
		})
		return
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse response",
		})
		return
	}

	c.JSON(http.StatusOK, authResp)
}

// GetMe returns the current user's profile
// GET /auth/me
func (h *Handler) GetMe(c *gin.Context) {
	token, err := GetUserToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "not authenticated",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Get user from Supabase Auth
	resp, err := h.supabaseClient.Get(ctx, "/auth/v1/user", token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch user",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to fetch user",
			"details": string(body),
		})
		return
	}

	var user map[string]interface{}
	if err := json.Unmarshal(body, &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse user",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Logout handles user logout
// POST /auth/logout
func (h *Handler) Logout(c *gin.Context) {
	token, err := GetUserToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "not authenticated",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Call Supabase logout endpoint
	resp, err := h.supabaseClient.Post(ctx, "/auth/v1/logout", nil, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "logout failed",
		})
		return
	}
	defer resp.Body.Close()

	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}

// GetUserProfile fetches user profile from profiles table
func (h *Handler) GetUserProfile(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "not authenticated",
		})
		return
	}

	token, _ := GetUserToken(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Query profiles table
	path := fmt.Sprintf("/rest/v1/profiles?id=eq.%s&select=*", userID)
	resp, err := h.supabaseClient.Get(ctx, path, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch profile",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to fetch profile",
			"details": string(body),
		})
		return
	}

	var profiles []ProfileResponse
	if err := json.Unmarshal(body, &profiles); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse profile",
		})
		return
	}

	if len(profiles) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "profile not found",
		})
		return
	}

	c.JSON(http.StatusOK, profiles[0])
}
