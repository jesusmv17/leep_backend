// Package engagement provides handlers for fan engagement and analytics features.
// This package manages all user interactions with songs, including:
//   - Comments: Fans can comment on songs
//   - Reviews: Fans can rate songs (1-5 stars) and leave reviews
//   - Tips: Fans can monetarily tip artists
//   - Events: Track analytics events (plays, views)
//   - Analytics: Retrieve artist dashboard data (play counts, engagement stats)
//
// These features enable the core engagement loop:
//   Artist uploads song → Fan discovers song → Fan engages (comment/review/tip) →
//   Artist sees engagement metrics → Both parties continue interacting
//
// All engagement data is stored in Supabase with RLS policies to ensure
// data integrity and proper access control.
package engagement

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

// Handler manages engagement endpoints
type Handler struct {
	supabaseClient *supabase.Client
}

// NewHandler creates a new engagement handler
func NewHandler(supabaseClient *supabase.Client) *Handler {
	return &Handler{
		supabaseClient: supabaseClient,
	}
}

// CreateCommentRequest represents a comment request
type CreateCommentRequest struct {
	SongID string `json:"song_id" binding:"required"`
	Body   string `json:"body" binding:"required"`
}

// CreateReviewRequest represents a review request
type CreateReviewRequest struct {
	SongID string `json:"song_id" binding:"required"`
	Rating int    `json:"rating" binding:"required,min=1,max=5"`
	Body   string `json:"body"`
}

// CreateTipRequest represents a tip request
type CreateTipRequest struct {
	SongID      string `json:"song_id" binding:"required"`
	AmountCents int    `json:"amount_cents" binding:"required,min=1"`
}

// CreateEventRequest represents an analytics event request
type CreateEventRequest struct {
	SongID    string `json:"song_id" binding:"required"`
	EventType string `json:"event_type" binding:"required"`
}

// CreateComment creates a new comment on a song
// POST /comments
func (h *Handler) CreateComment(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	var req CreateCommentRequest
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

	commentData := map[string]interface{}{
		"song_id":   req.SongID,
		"author_id": userID,
		"body":      req.Body,
	}

	resp, err := h.supabaseClient.Post(ctx, "/rest/v1/comments", commentData, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create comment",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to create comment",
			"details": string(body),
		})
		return
	}

	var comments []map[string]interface{}
	if err := json.Unmarshal(body, &comments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse response",
		})
		return
	}

	if len(comments) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "no comment returned from database",
		})
		return
	}

	c.JSON(http.StatusCreated, comments[0])
}

// ListComments returns comments for a song
// GET /songs/:id/comments
func (h *Handler) ListComments(c *gin.Context) {
	songID := c.Param("id")
	token, _ := auth.GetUserToken(c)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	path := fmt.Sprintf("/rest/v1/comments?song_id=eq.%s&select=*&order=created_at.desc", songID)
	resp, err := h.supabaseClient.Get(ctx, path, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch comments",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to fetch comments",
			"details": string(body),
		})
		return
	}

	var comments []map[string]interface{}
	if err := json.Unmarshal(body, &comments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse comments",
		})
		return
	}

	c.JSON(http.StatusOK, comments)
}

// CreateReview creates a new review for a song
// POST /reviews
func (h *Handler) CreateReview(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	var req CreateReviewRequest
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

	reviewData := map[string]interface{}{
		"song_id":     req.SongID,
		"reviewer_id": userID,
		"rating":      req.Rating,
		"body":        req.Body,
	}

	resp, err := h.supabaseClient.Post(ctx, "/rest/v1/reviews", reviewData, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create review",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to create review",
			"details": string(body),
		})
		return
	}

	var reviews []map[string]interface{}
	if err := json.Unmarshal(body, &reviews); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse response",
		})
		return
	}

	if len(reviews) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "no review returned from database",
		})
		return
	}

	c.JSON(http.StatusCreated, reviews[0])
}

// ListReviews returns reviews for a song
// GET /songs/:id/reviews
func (h *Handler) ListReviews(c *gin.Context) {
	songID := c.Param("id")
	token, _ := auth.GetUserToken(c)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	path := fmt.Sprintf("/rest/v1/reviews?song_id=eq.%s&select=*&order=created_at.desc", songID)
	resp, err := h.supabaseClient.Get(ctx, path, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch reviews",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to fetch reviews",
			"details": string(body),
		})
		return
	}

	var reviews []map[string]interface{}
	if err := json.Unmarshal(body, &reviews); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse reviews",
		})
		return
	}

	c.JSON(http.StatusOK, reviews)
}

// CreateTip creates a new tip for a song
// POST /tips
func (h *Handler) CreateTip(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	var req CreateTipRequest
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

	tipData := map[string]interface{}{
		"song_id":      req.SongID,
		"tipper_id":    userID,
		"amount_cents": req.AmountCents,
	}

	resp, err := h.supabaseClient.Post(ctx, "/rest/v1/tips", tipData, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create tip",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to create tip",
			"details": string(body),
		})
		return
	}

	var tips []map[string]interface{}
	if err := json.Unmarshal(body, &tips); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse response",
		})
		return
	}

	if len(tips) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "no tip returned from database",
		})
		return
	}

	c.JSON(http.StatusCreated, tips[0])
}

// CreateEvent logs an analytics event (play, view, etc.)
// POST /events
func (h *Handler) CreateEvent(c *gin.Context) {
	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// User ID is optional for events (can track anonymous plays)
	userID, _ := auth.GetUserID(c)
	token, _ := auth.GetUserToken(c)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	eventData := map[string]interface{}{
		"song_id":    req.SongID,
		"event_type": req.EventType,
		"user_id":    userID,
	}

	resp, err := h.supabaseClient.Post(ctx, "/rest/v1/events", eventData, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create event",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to create event",
			"details": string(body),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "event logged successfully",
	})
}

// GetArtistAnalytics returns analytics dashboard for an artist
// GET /analytics/artist/:id
func (h *Handler) GetArtistAnalytics(c *gin.Context) {
	artistID := c.Param("id")
	token, _ := auth.GetUserToken(c)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Call Supabase RPC function artist_dashboard
	rpcData := map[string]interface{}{
		"artist_id": artistID,
	}

	resp, err := h.supabaseClient.Post(ctx, "/rest/v1/rpc/artist_dashboard", rpcData, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch analytics",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to fetch analytics",
			"details": string(body),
		})
		return
	}

	var analytics interface{}
	if err := json.Unmarshal(body, &analytics); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse analytics",
		})
		return
	}

	c.JSON(http.StatusOK, analytics)
}
