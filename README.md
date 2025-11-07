# Leep Audio Backend

Production-ready Go backend for Leep Audio MVP with PostgreSQL, Docker, and DigitalOcean Spaces integration.

## 🚀 Quick Start

```bash
# Start Docker containers
docker compose up -d

# Install dependencies
npm install
go mod tidy

# Setup database
npx prisma generate
npx prisma migrate dev

# Start server
go run main.go
```

Server runs on `http://localhost:3000`

## 📚 Documentation

All documentation is in the [`/docs`](/docs) folder:

- **[Setup Guide](docs/setup-guides/SETUP.md)** - Complete local development setup
- **[Demo Setup](docs/setup-guides/DEMO_SETUP.md)** - Fresh machine setup (45-60 min)
- **[Demo Guide](docs/demo/DEMO_GUIDE.md)** - Sponsor demo script
- **[Week 1 Status](docs/reference/WEEK1_STATUS.md)** - Deliverables completion report

## 🏗️ Architecture

```
leep_backend/
├── internal/
│   ├── db/          # Database connection pooling
│   ├── health/      # Health check endpoints
│   └── storage/     # DigitalOcean Spaces client
├── prisma/
│   ├── schema.prisma       # Database schema (9 models)
│   └── migrations/         # Database migrations
├── docs/            # All documentation
└── main.go          # Application entry point
```

## 🔧 Tech Stack

- **Runtime**: Go 1.25
- **Framework**: Gin Web Framework
- **Database**: PostgreSQL 16 (via Docker)
- **ORM**: Prisma (for migrations)
- **Driver**: pgx/v5 (connection pooling)
- **Storage**: DigitalOcean Spaces (S3-compatible)
- **DevOps**: Docker Compose, GitHub Actions

## 📊 API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /health` | Basic health check |
| `GET /health/db` | Database health with pool stats |
| `GET /api/v1/status` | API version and status |
| `GET /ping` | Legacy ping endpoint |

## 🗄️ Database Schema

9 tables supporting the full MVP:
- **profiles** - User management with roles (fan, artist, producer, admin)
- **songs** - Artist content with publish controls
- **projects** - Collaboration workspaces
- **project_invitations** - Producer invite system
- **stems** - Individual audio file uploads
- **comments** - Fan engagement
- **reviews** - 5-star rating system
- **tips** - Artist monetization
- **events** - Analytics tracking (plays, views)

## 🛠️ Development

### Prerequisites
- Docker Desktop
- Go 1.25+
- Node.js 16+ (for Prisma)

### Common Commands

```bash
# Start environment
make docker-up

# Apply migrations
make db-migrate

# Run server
make dev

# View database
make db-studio
```

See `Makefile` for all commands.

## 📦 Deployment

Backend is ready for deployment to DigitalOcean or Render.

See [SETUP.md](docs/setup-guides/SETUP.md) for production deployment guide.

## 👥 Team

- **Jesus** - Backend Core (Auth, Roles)
- **Brendan** - Media & Collaboration
- **Kyle** - Engagement & Analytics
- **Chandler** - DevOps & Infrastructure
- **Yaman** - Frontend Integration & QA

## 📝 License

UNLICENSED - Private project for Leep Inc.
