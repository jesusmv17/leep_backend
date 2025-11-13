// Package projects provides handlers for collaboration and project management.
// This package enables artists and producers to collaborate on music projects by:
//   - Creating collaborative projects (artists)
//   - Inviting producers to projects (project owners)
//   - Uploading stems (audio files) to projects (invited collaborators)
//   - Managing project access and permissions
//
// The collaboration workflow:
//   1. Artist creates a project
//   2. Artist invites producer(s) via their user ID
//   3. Invited producers can upload stems to the project
//   4. All collaborators can view stems in the project
//
// Access control is enforced through Supabase RLS policies based on
// project ownership and invitation status.
package projects

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

// Handler manages project endpoints
type Handler struct {
	supabaseClient *supabase.Client
}

// NewHandler creates a new projects handler
func NewHandler(supabaseClient *supabase.Client) *Handler {
	return &Handler{
		supabaseClient: supabaseClient,
	}
}

// CreateProjectRequest represents the create project request
type CreateProjectRequest struct {
	Title string `json:"title" binding:"required"`
}

// InviteRequest represents a project invitation request
type InviteRequest struct {
	InviteeID string `json:"invitee_id" binding:"required"`
}

// CreateStemRequest represents a stem upload request
type CreateStemRequest struct {
	Name    string `json:"name" binding:"required"`
	FileURL string `json:"file_url" binding:"required"`
}

// CreateProject creates a new project
// POST /projects
func (h *Handler) CreateProject(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	var req CreateProjectRequest
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

	// Create project in Supabase
	projectData := map[string]interface{}{
		"owner_id": userID,
		"title":    req.Title,
	}

	resp, err := h.supabaseClient.Post(ctx, "/rest/v1/projects", projectData, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create project",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to create project",
			"details": string(body),
		})
		return
	}

	var projects []map[string]interface{}
	if err := json.Unmarshal(body, &projects); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse response",
		})
		return
	}

	if len(projects) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "no project returned from database",
		})
		return
	}

	c.JSON(http.StatusCreated, projects[0])
}

// ListProjects returns user's projects
// GET /projects
func (h *Handler) ListProjects(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	token, _ := auth.GetUserToken(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get user's projects (owned or invited to)
	path := fmt.Sprintf("/rest/v1/projects?owner_id=eq.%s&select=*&order=created_at.desc", userID)
	resp, err := h.supabaseClient.Get(ctx, path, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch projects",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to fetch projects",
			"details": string(body),
		})
		return
	}

	var projects []map[string]interface{}
	if err := json.Unmarshal(body, &projects); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse projects",
		})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// GetProject returns a single project by ID
// GET /projects/:id
func (h *Handler) GetProject(c *gin.Context) {
	projectID := c.Param("id")
	token, _ := auth.GetUserToken(c)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	path := fmt.Sprintf("/rest/v1/projects?id=eq.%s&select=*", projectID)
	resp, err := h.supabaseClient.Get(ctx, path, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch project",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to fetch project",
			"details": string(body),
		})
		return
	}

	var projects []map[string]interface{}
	if err := json.Unmarshal(body, &projects); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse project",
		})
		return
	}

	if len(projects) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "project not found",
		})
		return
	}

	c.JSON(http.StatusOK, projects[0])
}

// InviteToProject invites a user to collaborate on a project
// POST /projects/:id/invite
func (h *Handler) InviteToProject(c *gin.Context) {
	projectID := c.Param("id")
	token, err := auth.GetUserToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	var req InviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
			"details": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Create invitation in Supabase
	inviteData := map[string]interface{}{
		"project_id": projectID,
		"invitee_id": req.InviteeID,
	}

	resp, err := h.supabaseClient.Post(ctx, "/rest/v1/project_invitations", inviteData, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create invitation",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to create invitation",
			"details": string(body),
		})
		return
	}

	var invitations []map[string]interface{}
	if err := json.Unmarshal(body, &invitations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse response",
		})
		return
	}

	if len(invitations) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "no invitation returned from database",
		})
		return
	}

	c.JSON(http.StatusCreated, invitations[0])
}

// CreateStem uploads a stem to a project
// POST /projects/:id/stems
func (h *Handler) CreateStem(c *gin.Context) {
	projectID := c.Param("id")
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	var req CreateStemRequest
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

	// Create stem in Supabase
	stemData := map[string]interface{}{
		"project_id":  projectID,
		"uploader_id": userID,
		"name":        req.Name,
		"file_url":    req.FileURL,
	}

	resp, err := h.supabaseClient.Post(ctx, "/rest/v1/stems", stemData, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create stem",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to create stem",
			"details": string(body),
		})
		return
	}

	var stems []map[string]interface{}
	if err := json.Unmarshal(body, &stems); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse response",
		})
		return
	}

	if len(stems) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "no stem returned from database",
		})
		return
	}

	c.JSON(http.StatusCreated, stems[0])
}

// ListStems returns stems for a project
// GET /projects/:id/stems
func (h *Handler) ListStems(c *gin.Context) {
	projectID := c.Param("id")
	token, _ := auth.GetUserToken(c)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	path := fmt.Sprintf("/rest/v1/stems?project_id=eq.%s&select=*&order=created_at.desc", projectID)
	resp, err := h.supabaseClient.Get(ctx, path, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch stems",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.JSON(resp.StatusCode, gin.H{
			"error": "failed to fetch stems",
			"details": string(body),
		})
		return
	}

	var stems []map[string]interface{}
	if err := json.Unmarshal(body, &stems); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse stems",
		})
		return
	}

	c.JSON(http.StatusOK, stems)
}
