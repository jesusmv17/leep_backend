package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jesusmv17/leep_backend/internal/db"
)

// Handler manages health check endpoints
type Handler struct {
	dbPool *db.Pool
}

// NewHandler creates a new health check handler
func NewHandler(dbPool *db.Pool) *Handler {
	return &Handler{
		dbPool: dbPool,
	}
}

// BasicHealth returns a simple 200 OK response
// GET /health
func (h *Handler) BasicHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "leep-backend",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

// DatabaseHealth checks database connectivity
// GET /health/db
func (h *Handler) DatabaseHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Perform health check query
	if err := h.dbPool.HealthCheck(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "error",
			"database": "unreachable",
			"error":    err.Error(),
			"time":     time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	// Get pool stats
	stats := h.dbPool.Stat()

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"database": "connected",
		"pool": gin.H{
			"total_connections": stats.TotalConns(),
			"idle_connections":  stats.IdleConns(),
			"acquired_conns":    stats.AcquiredConns(),
		},
		"time": time.Now().UTC().Format(time.RFC3339),
	})
}
