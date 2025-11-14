package main

import "time"

type Project struct {
    ID        int64     `json:"id"`
    OwnerID   string    `json:"owner_id"`
    Title     string    `json:"title"`
    CreatedAt time.Time `json:"created_at"`
}

type ProjectInvitation struct {
    ID        int64     `json:"id"`
    ProjectID int64     `json:"project_id"`
    InviteeID string    `json:"invitee_id"`
    CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
    ID        int64     `json:"id"`
    SongID    int64     `json:"song_id"`
    AuthorID  string    `json:"author_id"`
    Body      string    `json:"body"`
    CreatedAt time.Time `json:"created_at"`
}

type Review struct {
    ID         int64     `json:"id"`
    SongID     int64     `json:"song_id"`
    ReviewerID string    `json:"reviewer_id"`
    Rating     int       `json:"rating"`
    Body       string    `json:"body"`
    CreatedAt  time.Time `json:"created_at"`
}

type Tip struct {
    ID        int64     `json:"id"`
    SongID    int64     `json:"song_id"`
    SenderID  string    `json:"sender_id"`
    Amount    float64   `json:"amount"`
    CreatedAt time.Time `json:"created_at"`
}
