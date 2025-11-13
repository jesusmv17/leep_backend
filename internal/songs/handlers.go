// Package songs provides handlers for song management in the Leep Audio platform.
// This package handles all CRUD operations for songs, including:
//   - Creating songs (artists only)
//   - Listing songs (public published songs, or user's own songs)
//   - Getting individual songs
//   - Updating songs (ownership required, enforced by RLS)
//   - Deleting songs (ownership required, enforced by RLS)
//   - Publishing/unpublishing songs (via Supabase RPC)
//
// All data access is controlled by Supabase Row Level Security (RLS) policies,
// which ensure users can only modify their own content.
package songs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jesusmv17/leep_backend/internal/auth"
	"github.com/jesusmv17/leep_backend/internal/supabase"
)

// Handler manages song endpoints
type Handler struct {
	supabaseClient *supabase.Client
}

// NewHandler creates a new songs handler
func NewHandler(supabaseClient *supabase.Client) *Handler {
	return &Handler{
		supabaseClient: supabaseClient,
	}
}

// CreateSongRequest represents the create song request body
type CreateSongRequest struct {
	Title      string `json:"title" binding:"required"`
	AudioURL   string `json:"audio_url"`
	ArtworkURL string `json:"artwork_url"`
}

// PublishSongRequest represents the publish request
type PublishSongRequest struct {
	IsPublished bool `json:"is_published"`
}

// CreateSong creates a new song
// POST /songs
func (h *Handler) CreateSong(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	var req CreateSongRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
			"details": err.Error(),
		})
		return
	}

	token, _ := auth.GetUserToken(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Create song in Supabase (RLS will enforce artist_id = auth.uid())
	songData := map[string]interface{}{
		"artist_id":  userID,
		"title":      req.Title,
		"audio_url":  req.AudioURL,
		"artwork_url": req.ArtworkURL,
		"is_published": false,
	}

	resp, err := h.supabaseClient.Post(ctx, "/rest/v1/songs", songData, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create song",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to create song",
			"details": string(body),
		})
		return
	}

	var songs []map[string]interface{}
	if err := json.Unmarshal(body, &songs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse response",
		})
		return
	}

	if len(songs) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "no song returned from database",
		})
		return
	}

	c.JSON(http.StatusCreated, songs[0])
}

// ListSongs returns public songs or user's own songs
// GET /songs
func (h *Handler) ListSongs(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Check if user is authenticated (optional)
	userID, _ := auth.GetUserID(c)
	token, _ := auth.GetUserToken(c)

	var path string
	if userID != "" {
		// If authenticated, return user's own songs (published and unpublished)
		path = fmt.Sprintf("/rest/v1/songs?artist_id=eq.%s&select=*&order=created_at.desc", userID)
	} else {
		// If not authenticated, only return published songs
		path = "/rest/v1/songs?is_published=eq.true&select=*&order=created_at.desc"
	}

	resp, err := h.supabaseClient.Get(ctx, path, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch songs",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to fetch songs",
			"details": string(body),
		})
		return
	}

	var songs []map[string]interface{}
	if err := json.Unmarshal(body, &songs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse songs",
		})
		return
	}

	c.JSON(http.StatusOK, songs)
}

// GetSong returns a single song by ID
// GET /songs/:id
func (h *Handler) GetSong(c *gin.Context) {
	songID := c.Param("id")
	token, _ := auth.GetUserToken(c)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	path := fmt.Sprintf("/rest/v1/songs?id=eq.%s&select=*", songID)
	resp, err := h.supabaseClient.Get(ctx, path, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch song",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to fetch song",
			"details": string(body),
		})
		return
	}

	var songs []map[string]interface{}
	if err := json.Unmarshal(body, &songs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse song",
		})
		return
	}

	if len(songs) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "song not found",
		})
		return
	}

	c.JSON(http.StatusOK, songs[0])
}

// PublishSong publishes or unpublishes a song (calls Supabase RPC)
// POST /songs/:id/publish
func (h *Handler) PublishSong(c *gin.Context) {
	songID := c.Param("id")
	token, err := auth.GetUserToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	var req PublishSongRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Default to publishing if no body provided
		req.IsPublished = true
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Call Supabase RPC function publish_song
	rpcData := map[string]interface{}{
		"song_id": songID,
	}

	resp, err := h.supabaseClient.Post(ctx, "/rest/v1/rpc/publish_song", rpcData, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to publish song",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to publish song",
			"details": string(body),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "song published successfully",
		"song_id": songID,
	})
}

// UpdateSong updates a song
// PATCH /songs/:id
func (h *Handler) UpdateSong(c *gin.Context) {
	songID := c.Param("id")
	token, err := auth.GetUserToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Update song in Supabase (RLS will enforce ownership)
	path := fmt.Sprintf("/rest/v1/songs?id=eq.%s", songID)
	resp, err := h.supabaseClient.Patch(ctx, path, updates, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update song",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to update song",
			"details": string(body),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "song updated successfully",
	})
}

// DeleteSong deletes a song
// DELETE /songs/:id
func (h *Handler) DeleteSong(c *gin.Context) {
	songID := c.Param("id")
	token, err := auth.GetUserToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Delete song from Supabase (RLS will enforce ownership)
	path := fmt.Sprintf("/rest/v1/songs?id=eq.%s", songID)
	resp, err := h.supabaseClient.Delete(ctx, path, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete song",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to delete song",
			"details": string(body),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "song deleted successfully",
	})
}
