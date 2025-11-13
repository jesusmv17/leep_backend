# Leep Audio Backend

Production-ready Go backend for Leep Audio MVP with Supabase integration.

##  Quick Start

```bash
# Install dependencies
go mod tidy

# Start server
go run main.go
```

Server runs on `http://localhost:3000`

**Week 2 & 3 Complete**: Full MVP backend with auth, songs, projects, engagement, analytics, and admin features!

##  Documentation

- **[API Documentation](API.md)** - Complete API reference with all endpoints
- **[Deployment Guide](DEPLOY.md)** - Step-by-step Render.com deployment
- **[Week 1 Status](docs/reference/WEEK1_STATUS.md)** - Week 1 deliverables

### Legacy Documentation (Week 1)
- **[Setup Guide](docs/setup-guides/SETUP.md)** - Complete local development setup
- **[Demo Guide](docs/demo/DEMO_GUIDE.md)** - Sponsor demo script

##  Architecture

```
leep_backend/
 internal/
    db/          # Database connection pooling
    health/      # Health check endpoints
    storage/     # DigitalOcean Spaces client
 prisma/
    schema.prisma       # Database schema (9 models)
    migrations/         # Database migrations
 docs/            # All documentation
 main.go          # Application entry point
```

##  Tech Stack

- **Runtime**: Go 1.25
- **Framework**: Gin Web Framework
- **Backend-as-a-Service**: Supabase (Auth, DB, Storage)
- **Database**: PostgreSQL (via Supabase)
- **Authentication**: JWT with Supabase Auth
- **Deployment**: Render.com
- **Middleware**: CORS, Rate Limiting, Structured Logging

##  API Endpoints

### Authentication
- `POST /api/v1/auth/signup` - User registration
- `POST /api/v1/auth/login` - User login
- `GET /api/v1/auth/me` - Get current user
- `GET /api/v1/auth/profile` - Get user profile
- `POST /api/v1/auth/logout` - Logout

### Songs
- `GET /api/v1/songs` - List public songs
- `POST /api/v1/songs` - Create song
- `GET /api/v1/songs/:id` - Get song
- `PATCH /api/v1/songs/:id` - Update song
- `DELETE /api/v1/songs/:id` - Delete song
- `POST /api/v1/songs/:id/publish` - Publish song

### Projects & Collaboration
- `GET /api/v1/projects` - List projects
- `POST /api/v1/projects` - Create project
- `POST /api/v1/projects/:id/invite` - Invite collaborator
- `POST /api/v1/projects/:id/stems` - Upload stem
- `GET /api/v1/projects/:id/stems` - List stems

### Engagement
- `POST /api/v1/comments` - Create comment
- `POST /api/v1/reviews` - Create review
- `POST /api/v1/tips` - Create tip
- `GET /api/v1/songs/:id/comments` - List comments
- `GET /api/v1/songs/:id/reviews` - List reviews

### Analytics
- `POST /api/v1/events` - Log event (play/view)
- `GET /api/v1/analytics/artist/:id` - Artist dashboard

### Admin (Moderation)
- `POST /api/v1/admin/songs/:id/takedown` - Takedown song
- `DELETE /api/v1/admin/comments/:id` - Delete comment
- `GET /api/v1/admin/users` - List users
- `PATCH /api/v1/admin/users/:id/role` - Update user role

See [API.md](API.md) for complete documentation.

##  Database Schema (Supabase)

9 tables supporting the full MVP:
- **profiles** - User management with roles (fan, artist, producer, admin)
- **songs** - Artist content with publish controls (RLS enforced)
- **projects** - Collaboration workspaces
- **project_invitations** - Producer invite system
- **stems** - Individual audio file uploads
- **comments** - Fan engagement (with Realtime)
- **reviews** - 5-star rating system (CHECK constraints)
- **tips** - Artist monetization
- **events** - Analytics tracking (plays, views)

All tables have Row Level Security (RLS) policies enforced via Supabase.

##  Development

### Prerequisites
- Go 1.25+
- Supabase account & credentials

### Environment Variables

Create `.env` file:
```bash
PORT=3000
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key
SUPABASE_JWT_SECRET=your-jwt-secret
JWT_EXPIRY=3600
```

### Run Locally

```bash
# Install dependencies
go mod tidy

# Run server
go run main.go

# Or build and run
go build -o leep_backend main.go
./leep_backend
```

Server will start on `http://localhost:3000`

##  Deployment

Backend is production-ready for Render.com deployment.

See [DEPLOY.md](DEPLOY.md) for step-by-step deployment guide.

**Quick Deploy**:
1. Push to GitHub
2. Connect to Render
3. Set environment variables
4. Deploy!

Live in 5 minutes 

##  Team

- **Jesus** - Backend Core (Auth, Roles)
- **Brendan** - Media & Collaboration
- **Kyle** - Engagement & Analytics
- **Chandler** - DevOps & Infrastructure
- **Yaman** - Frontend Integration & QA

##  License

UNLICENSED - Private project for Leep Inc.
