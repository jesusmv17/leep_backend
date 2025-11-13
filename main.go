// Leep Audio Backend - Main Application Entry Point
//
// This is the main server application for the Leep Audio MVP backend.
// It provides a RESTful API for the music collaboration platform with:
//   - Authentication (Supabase Auth)
//   - Song management (CRUD operations)
//   - Project collaboration (artist-producer workflow)
//   - Fan engagement (comments, reviews, tips)
//   - Analytics (play tracking, artist dashboards)
//   - Admin moderation (content management)
//
// Architecture:
//   Frontend (Vercel) <-> Go Backend (Render) <-> Supabase (Auth + DB + Storage)
//
// The backend acts as a middleware layer that:
//   - Validates JWT tokens from Supabase
//   - Enforces business logic
//   - Proxies requests to Supabase REST API
//   - Provides additional features (rate limiting, logging, CORS)
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jesusmv17/leep_backend/internal/admin"
	"github.com/jesusmv17/leep_backend/internal/auth"
	"github.com/jesusmv17/leep_backend/internal/engagement"
	"github.com/jesusmv17/leep_backend/internal/health"
	"github.com/jesusmv17/leep_backend/internal/middleware"
	"github.com/jesusmv17/leep_backend/internal/projects"
	"github.com/jesusmv17/leep_backend/internal/songs"
	"github.com/jesusmv17/leep_backend/internal/supabase"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	// This is only used in local development; production uses system environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize Supabase client with credentials from environment
	// Required env vars: SUPABASE_URL, SUPABASE_ANON_KEY, SUPABASE_SERVICE_ROLE_KEY
	supabaseClient, err := supabase.NewClient()
	if err != nil {
		log.Fatalf("Failed to initialize Supabase client: %v", err)
	}

	log.Println("Supabase client initialized successfully")

	// Initialize Gin router in release mode for production
	// Use gin.New() instead of gin.Default() to have full control over middleware
	r := gin.New()

	// Apply global middleware in order of execution:
	r.Use(gin.Recovery())                           // 1. Recover from panics (prevents crashes)
	r.Use(middleware.CORS())                        // 2. Enable CORS for frontend access
	r.Use(middleware.Logger())                      // 3. Log all requests with user tracking
	r.Use(middleware.RateLimit(100, time.Minute))   // 4. Rate limit: 100 requests per minute per IP

	// Initialize all route handlers with Supabase client
	// Each handler manages a specific domain of the application
	healthHandler := health.NewHandler(nil)                      // Health checks don't need Supabase
	authHandler := auth.NewHandler(supabaseClient)               // Authentication & user management
	songsHandler := songs.NewHandler(supabaseClient)             // Song CRUD operations
	projectsHandler := projects.NewHandler(supabaseClient)       // Collaboration & stems
	engagementHandler := engagement.NewHandler(supabaseClient)   // Comments, reviews, tips, analytics
	adminHandler := admin.NewHandler(supabaseClient)             // Admin moderation

	// Health check routes (public)
	r.GET("/health", healthHandler.BasicHealth)

	// Legacy ping endpoint (for backwards compatibility)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// Status endpoint (public)
		api.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"service": "leep-backend",
				"version": "2.0.0-mvp",
				"status":  "operational",
			})
		})

		// Auth routes (public)
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/signup", authHandler.Signup)
			authGroup.POST("/login", authHandler.Login)
			authGroup.GET("/me", auth.RequireAuth(), authHandler.GetMe)
			authGroup.POST("/logout", auth.RequireAuth(), authHandler.Logout)
			authGroup.GET("/profile", auth.RequireAuth(), authHandler.GetUserProfile)
		}

		// Songs routes
		songsGroup := api.Group("/songs")
		{
			songsGroup.GET("", auth.OptionalAuth(), songsHandler.ListSongs)
			songsGroup.POST("", auth.RequireAuth(), songsHandler.CreateSong)
			songsGroup.GET("/:id", auth.OptionalAuth(), songsHandler.GetSong)
			songsGroup.PATCH("/:id", auth.RequireAuth(), songsHandler.UpdateSong)
			songsGroup.DELETE("/:id", auth.RequireAuth(), songsHandler.DeleteSong)
			songsGroup.POST("/:id/publish", auth.RequireAuth(), songsHandler.PublishSong)

			// Song engagement
			songsGroup.GET("/:id/comments", auth.OptionalAuth(), engagementHandler.ListComments)
			songsGroup.GET("/:id/reviews", auth.OptionalAuth(), engagementHandler.ListReviews)
		}

		// Projects routes (requires auth)
		projectsGroup := api.Group("/projects")
		projectsGroup.Use(auth.RequireAuth())
		{
			projectsGroup.GET("", projectsHandler.ListProjects)
			projectsGroup.POST("", projectsHandler.CreateProject)
			projectsGroup.GET("/:id", projectsHandler.GetProject)
			projectsGroup.POST("/:id/invite", projectsHandler.InviteToProject)
			projectsGroup.POST("/:id/stems", projectsHandler.CreateStem)
			projectsGroup.GET("/:id/stems", projectsHandler.ListStems)
		}

		// Engagement routes (requires auth)
		engagementGroup := api.Group("/")
		engagementGroup.Use(auth.RequireAuth())
		{
			engagementGroup.POST("/comments", engagementHandler.CreateComment)
			engagementGroup.POST("/reviews", engagementHandler.CreateReview)
			engagementGroup.POST("/tips", engagementHandler.CreateTip)
		}

		// Events routes (public for anonymous tracking)
		api.POST("/events", auth.OptionalAuth(), engagementHandler.CreateEvent)

		// Analytics routes (public)
		analyticsGroup := api.Group("/analytics")
		{
			analyticsGroup.GET("/artist/:id", auth.OptionalAuth(), engagementHandler.GetArtistAnalytics)
		}

		// Admin routes (requires admin role - TODO: add role check middleware)
		adminGroup := api.Group("/admin")
		adminGroup.Use(auth.RequireAuth()) // TODO: Add RequireRole("admin")
		{
			adminGroup.POST("/songs/:id/takedown", adminHandler.TakedownSong)
			adminGroup.DELETE("/comments/:id", adminHandler.DeleteComment)
			adminGroup.DELETE("/reviews/:id", adminHandler.DeleteReview)
			adminGroup.GET("/users", adminHandler.GetAllUsers)
			adminGroup.PATCH("/users/:id/role", adminHandler.UpdateUserRole)
		}
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Leep Audio Backend Server")
		log.Printf("🔗 Server starting on port %s", port)
		log.Printf("📡 Supabase URL: %s", os.Getenv("SUPABASE_URL"))
		log.Printf("✅ CORS enabled for Vercel")
		log.Printf("🛡️  Rate limiting: 100 req/min")
		log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}

