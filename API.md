# Leep Audio Backend API Documentation

## Base URL
- **Local**: `http://localhost:3000`
- **Production**: `https://your-app.onrender.com`

## Authentication
All authenticated endpoints require a `Bearer` token in the `Authorization` header:
```
Authorization: Bearer <your_jwt_token>
```

---

##  Health & Status

### `GET /health`
Health check endpoint
- **Auth**: None
- **Response**: `{ "status": "ok", "service": "leep-backend", "time": "..." }`

### `GET /api/v1/status`
API status endpoint
- **Auth**: None
- **Response**: `{ "service": "leep-backend", "version": "2.0.0-mvp", "status": "operational" }`

---

##  Authentication Endpoints

### `POST /api/v1/auth/signup`
Create a new user account
- **Auth**: None
- **Body**:
  ```json
  {
    "email": "user@example.com",
    "password": "password123",
    "display_name": "John Doe"
  }
  ```
- **Response**: `{ "access_token": "...", "user": {...} }`

### `POST /api/v1/auth/login`
Login with email and password
- **Auth**: None
- **Body**:
  ```json
  {
    "email": "user@example.com",
    "password": "password123"
  }
  ```
- **Response**: `{ "access_token": "...", "refresh_token": "...", "user": {...} }`

### `GET /api/v1/auth/me`
Get current user from token
- **Auth**: Required
- **Response**: User object from Supabase Auth

### `GET /api/v1/auth/profile`
Get current user's profile from profiles table
- **Auth**: Required
- **Response**: `{ "id": "...", "display_name": "...", "role": "fan", "created_at": "..." }`

### `POST /api/v1/auth/logout`
Logout current user
- **Auth**: Required
- **Response**: `{ "message": "logged out successfully" }`

---

##  Songs Endpoints

### `GET /api/v1/songs`
List public songs (or user's own songs if authenticated)
- **Auth**: Optional
- **Response**: Array of song objects

### `POST /api/v1/songs`
Create a new song
- **Auth**: Required (Artist)
- **Body**:
  ```json
  {
    "title": "My Song",
    "audio_url": "https://...",
    "artwork_url": "https://..."
  }
  ```
- **Response**: Created song object

### `GET /api/v1/songs/:id`
Get a single song by ID
- **Auth**: Optional
- **Response**: Song object

### `PATCH /api/v1/songs/:id`
Update a song (ownership required)
- **Auth**: Required
- **Body**: Any song fields to update
- **Response**: `{ "message": "song updated successfully" }`

### `DELETE /api/v1/songs/:id`
Delete a song (ownership required)
- **Auth**: Required
- **Response**: `{ "message": "song deleted successfully" }`

### `POST /api/v1/songs/:id/publish`
Publish or unpublish a song
- **Auth**: Required (Artist, owner)
- **Body**: `{ "is_published": true }` (optional)
- **Response**: `{ "message": "song published successfully", "song_id": "..." }`

### `GET /api/v1/songs/:id/comments`
Get comments for a song
- **Auth**: Optional
- **Response**: Array of comment objects

### `GET /api/v1/songs/:id/reviews`
Get reviews for a song
- **Auth**: Optional
- **Response**: Array of review objects

---

##  Projects & Collaboration

### `GET /api/v1/projects`
List user's projects
- **Auth**: Required
- **Response**: Array of project objects

### `POST /api/v1/projects`
Create a new project
- **Auth**: Required (Artist)
- **Body**:
  ```json
  {
    "title": "My Project"
  }
  ```
- **Response**: Created project object

### `GET /api/v1/projects/:id`
Get a single project by ID
- **Auth**: Required
- **Response**: Project object

### `POST /api/v1/projects/:id/invite`
Invite a producer to collaborate
- **Auth**: Required (Project owner)
- **Body**:
  ```json
  {
    "invitee_id": "user-uuid"
  }
  ```
- **Response**: Invitation object

### `POST /api/v1/projects/:id/stems`
Upload a stem to a project
- **Auth**: Required (Invited collaborator)
- **Body**:
  ```json
  {
    "name": "Bass Line",
    "file_url": "https://..."
  }
  ```
- **Response**: Created stem object

### `GET /api/v1/projects/:id/stems`
List stems for a project
- **Auth**: Required
- **Response**: Array of stem objects

---

##  Engagement Endpoints

### `POST /api/v1/comments`
Create a comment on a song
- **Auth**: Required
- **Body**:
  ```json
  {
    "song_id": "123",
    "body": "Great song!"
  }
  ```
- **Response**: Created comment object

### `POST /api/v1/reviews`
Create a review for a song
- **Auth**: Required
- **Body**:
  ```json
  {
    "song_id": "123",
    "rating": 5,
    "body": "Amazing track!"
  }
  ```
- **Response**: Created review object

### `POST /api/v1/tips`
Tip an artist
- **Auth**: Required
- **Body**:
  ```json
  {
    "song_id": "123",
    "amount_cents": 500
  }
  ```
- **Response**: Created tip object

---

##  Analytics Endpoints

### `POST /api/v1/events`
Log an analytics event (play, view, etc.)
- **Auth**: Optional
- **Body**:
  ```json
  {
    "song_id": "123",
    "event_type": "play"
  }
  ```
- **Response**: `{ "message": "event logged successfully" }`

### `GET /api/v1/analytics/artist/:id`
Get artist analytics dashboard
- **Auth**: Optional
- **Response**: Analytics data from `artist_dashboard` RPC function

---

##  Admin Endpoints

### `POST /api/v1/admin/songs/:id/takedown`
Takedown a song (admin only)
- **Auth**: Required (Admin)
- **Response**: `{ "message": "song taken down successfully" }`

### `DELETE /api/v1/admin/comments/:id`
Delete a comment (admin only)
- **Auth**: Required (Admin)
- **Response**: `{ "message": "comment deleted successfully" }`

### `DELETE /api/v1/admin/reviews/:id`
Delete a review (admin only)
- **Auth**: Required (Admin)
- **Response**: `{ "message": "review deleted successfully" }`

### `GET /api/v1/admin/users`
Get all users (admin only)
- **Auth**: Required (Admin)
- **Response**: Array of user profiles

### `PATCH /api/v1/admin/users/:id/role`
Update a user's role (admin only)
- **Auth**: Required (Admin)
- **Body**:
  ```json
  {
    "role": "artist"
  }
  ```
- **Response**: `{ "message": "user role updated successfully" }`

---

##  Error Responses

All endpoints return errors in this format:
```json
{
  "error": "error message",
  "details": "additional details if available"
}
```

Common HTTP status codes:
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `429` - Too Many Requests (rate limit)
- `500` - Internal Server Error

---

##  Rate Limiting

- **Limit**: 100 requests per minute per IP
- **Response**: `429 Too Many Requests` with retry_after duration

---

##  CORS

CORS is enabled for all origins. In production, this should be restricted to your Vercel frontend domain.
