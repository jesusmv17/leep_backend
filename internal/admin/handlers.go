// Package admin provides administrative and moderation functionality.
// This package handles platform moderation tasks that require elevated privileges:
//   - Taking down inappropriate songs (unpublishing)
//   - Deleting violating comments
//   - Deleting violating reviews
//   - Listing all users (admin dashboard)
//   - Updating user roles (promoting users to artist, producer, admin)
//
// Security: All operations in this package use the Supabase service role key,
// which bypasses Row Level Security (RLS) policies. These endpoints should be
// protected by admin-only middleware (currently using RequireAuth, but should
// be enhanced with role-based access control).
//
// WARNING: These operations are powerful and should only be accessible to
// trusted administrators. Ensure proper RBAC is implemented before production.
package admin

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

// Handler manages admin endpoints
type Handler struct {
	supabaseClient *supabase.Client
}

// NewHandler creates a new admin handler
func NewHandler(supabaseClient *supabase.Client) *Handler {
	return &Handler{
		supabaseClient: supabaseClient,
	}
}

// TakedownSong unpublishes a song (admin moderation)
// POST /admin/songs/:id/takedown
func (h *Handler) TakedownSong(c *gin.Context) {
	songID := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Call Supabase RPC function with service role key (admin action)
	rpcData := map[string]interface{}{
		"song_id": songID,
	}

	resp, err := h.supabaseClient.ServiceRolePost(ctx, "/rest/v1/rpc/admin_takedown_song", rpcData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to takedown song",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to takedown song",
			"details": string(body),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "song taken down successfully",
		"song_id": songID,
	})
}

// DeleteComment deletes a comment (admin moderation)
// DELETE /admin/comments/:id
func (h *Handler) DeleteComment(c *gin.Context) {
	commentID := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Call Supabase RPC function with service role key
	rpcData := map[string]interface{}{
		"comment_id": commentID,
	}

	resp, err := h.supabaseClient.ServiceRolePost(ctx, "/rest/v1/rpc/admin_delete_comment", rpcData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete comment",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to delete comment",
			"details": string(body),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "comment deleted successfully",
		"comment_id": commentID,
	})
}

// DeleteReview deletes a review (admin moderation)
// DELETE /admin/reviews/:id
func (h *Handler) DeleteReview(c *gin.Context) {
	reviewID := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Delete review using service role key
	path := fmt.Sprintf("/rest/v1/reviews?id=eq.%s", reviewID)
	resp, err := h.supabaseClient.ServiceRoleDelete(ctx, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete review",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to delete review",
			"details": string(body),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "review deleted successfully",
		"review_id": reviewID,
	})
}

// GetAllUsers returns all users (admin only)
// GET /admin/users
func (h *Handler) GetAllUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get all profiles using service role key
	resp, err := h.supabaseClient.Request(ctx, http.MethodGet, "/rest/v1/profiles?select=*&order=created_at.desc", nil, "", true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch users",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to fetch users",
			"details": string(body),
		})
		return
	}

	var users []map[string]interface{}
	if err := json.Unmarshal(body, &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse users",
		})
		return
	}

	c.JSON(http.StatusOK, users)
}

// UpdateUserRole updates a user's role (admin only)
// PATCH /admin/users/:id/role
func (h *Handler) UpdateUserRole(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Update user role using service role key
	path := fmt.Sprintf("/rest/v1/profiles?id=eq.%s", userID)
	updateData := map[string]interface{}{
		"role": req.Role,
	}

	resp, err := h.supabaseClient.ServiceRolePatch(ctx, path, updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update user role",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to update user role",
			"details": string(body),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user role updated successfully",
		"user_id": userID,
		"role":    req.Role,
	})
}
