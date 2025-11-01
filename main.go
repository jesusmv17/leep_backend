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
	"github.com/jesusmv17/leep_backend/internal/db"
	"github.com/jesusmv17/leep_backend/internal/health"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file (local dev only)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database connection pool
	ctx := context.Background()
	dbPool, err := db.NewPool(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize database pool: %v", err)
	}
	defer dbPool.Close()

	log.Println("✓ Database connection established")

	// Initialize Gin router
	r := gin.Default()

	// Initialize health handler
	healthHandler := health.NewHandler(dbPool)

	// Health check routes
	r.GET("/health", healthHandler.BasicHealth)
	r.GET("/health/db", healthHandler.DatabaseHealth)

	// Legacy ping endpoint (for backwards compatibility)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// API routes placeholder (will be expanded in Week 2)
	api := r.Group("/api/v1")
	{
		api.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"service": "leep-backend",
				"version": "1.0.0-mvp",
				"status":  "operational",
			})
		})
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

