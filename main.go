package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type createProjectInput struct {
	OwnerID string `json:"owner_id"`
	Title   string `json:"title"`
}

type inviteInput struct {
	ProjectID int64  `json:"project_id"`
	InviteeID string `json:"invitee_id"`
}

func main() {
	// Connect DB
	InitDB()

	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Server running and DB connected"})
	})

	// ------------------------
	// PROJECTS
	// ------------------------
	r.POST("/projects", func(c *gin.Context) {
		var body createProjectInput
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}

		sql := `
			INSERT INTO projects (owner_id, title)
			VALUES ($1, $2)
			RETURNING id, owner_id, title, created_at;
		`

		var p Project
		err := db.QueryRow(context.Background(), sql,
			body.OwnerID, body.Title,
		).Scan(&p.ID, &p.OwnerID, &p.Title, &p.CreatedAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, p)
	})

	// ------------------------
	// INVITES
	// ------------------------
	r.POST("/invite", func(c *gin.Context) {
		var body inviteInput
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}

		sql := `
			INSERT INTO project_invitations (project_id, invitee_id)
			VALUES ($1, $2)
			RETURNING id, project_id, invitee_id, created_at;
		`

		var inv ProjectInvitation
		err := db.QueryRow(context.Background(), sql,
			body.ProjectID, body.InviteeID,
		).Scan(&inv.ID, &inv.ProjectID, &inv.InviteeID, &inv.CreatedAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, inv)
	})

	// ------------------------
	// COMMENTS
	// ------------------------
	r.POST("/comments", func(c *gin.Context) {
		var body Comment
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}

		sql := `INSERT INTO comments (song_id, author_id, body)
		        VALUES ($1, $2, $3)
		        RETURNING id, song_id, author_id, body, created_at;`

		err := db.QueryRow(context.Background(), sql,
			body.SongID, body.AuthorID, body.Body,
		).Scan(&body.ID, &body.SongID, &body.AuthorID, &body.Body, &body.CreatedAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Record engagement event
		eventSQL := `
			INSERT INTO events (song_id, user_id, event_type)
			VALUES ($1, $2, $3);
		`
		db.Exec(context.Background(), eventSQL, body.SongID, body.AuthorID, "comment")

		c.JSON(http.StatusCreated, body)
	})

	// ------------------------
	// REVIEWS
	// ------------------------
	r.POST("/reviews", func(c *gin.Context) {
		var body Review
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}

		if body.Rating < 1 || body.Rating > 5 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rating must be 1-5"})
			return
		}

		sql := `INSERT INTO reviews (song_id, reviewer_id, rating, body)
		        VALUES ($1, $2, $3, $4)
		        RETURNING id, song_id, reviewer_id, rating, body, created_at;`

		err := db.QueryRow(context.Background(), sql,
			body.SongID, body.ReviewerID, body.Rating, body.Body,
		).Scan(&body.ID, &body.SongID, &body.ReviewerID, &body.Rating, &body.Body, &body.CreatedAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Record engagement event
		eventSQL := `
			INSERT INTO events (song_id, user_id, event_type)
			VALUES ($1, $2, $3);
		`
		db.Exec(context.Background(), eventSQL, body.SongID, body.ReviewerID, "review")

		c.JSON(http.StatusCreated, body)
	})

	// ------------------------
	// TIPS
	// ------------------------
	r.POST("/tips", func(c *gin.Context) {
		var body Tip
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}

		if body.Amount <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be > 0"})
			return
		}

		sql := `INSERT INTO tips (song_id, sender_id, amount)
		        VALUES ($1, $2, $3)
		        RETURNING id, song_id, sender_id, amount, created_at;`

		err := db.QueryRow(context.Background(), sql,
			body.SongID, body.SenderID, body.Amount,
		).Scan(&body.ID, &body.SongID, &body.SenderID, &body.Amount, &body.CreatedAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Record engagement event
		eventSQL := `
			INSERT INTO events (song_id, user_id, event_type)
			VALUES ($1, $2, $3);
		`
		db.Exec(context.Background(), eventSQL, body.SongID, body.SenderID, "tip")

		c.JSON(http.StatusCreated, body)
	})

	// ------------------------
	// ANALYTICS
	// ------------------------
	RegisterAnalyticsRoutes(r)

	// Run server
	r.Run(":8080")
}
