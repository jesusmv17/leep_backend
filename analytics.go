package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterAnalyticsRoutes defines the analytics endpoints
func RegisterAnalyticsRoutes(r *gin.Engine) {
	// GET /analytics/realtime
	r.GET("/analytics/realtime", func(c *gin.Context) {
		sql := `
			SELECT 
				songs.id AS song_id,
				songs.title AS song_title,
				COUNT(events.id) AS total_events,
				COUNT(CASE WHEN events.event_type = 'comment' THEN 1 END) AS total_comments,
				COUNT(CASE WHEN events.event_type = 'review' THEN 1 END) AS total_reviews,
				COUNT(CASE WHEN events.event_type = 'tip' THEN 1 END) AS total_tips
			FROM songs
			LEFT JOIN events ON songs.id = events.song_id
			GROUP BY songs.id
			ORDER BY total_events DESC;
		`

		rows, err := db.Query(context.Background(), sql)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		type SongAnalytics struct {
			SongID        int64  `json:"song_id"`
			SongTitle     string `json:"song_title"`
			TotalEvents   int64  `json:"total_events"`
			TotalComments int64  `json:"total_comments"`
			TotalReviews  int64  `json:"total_reviews"`
			TotalTips     int64  `json:"total_tips"`
		}

		var analytics []SongAnalytics
		for rows.Next() {
			var a SongAnalytics
			if err := rows.Scan(&a.SongID, &a.SongTitle, &a.TotalEvents, &a.TotalComments, &a.TotalReviews, &a.TotalTips); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			analytics = append(analytics, a)
		}

		c.JSON(http.StatusOK, analytics)
	})
}
