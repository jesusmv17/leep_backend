# Leep Audio Backend - Week 1 Setup Guide

This guide walks through setting up the local development environment and preparing for DigitalOcean deployment.

## Prerequisites

- Docker & Docker Compose installed
- Go 1.25+ installed
- Node.js 16+ (for Prisma tooling)
- DigitalOcean account (for production deployment)

## Quick Start (Local Development)

### 1. Start Docker Database

```bash
# Start Postgres + Adminer
docker compose up -d

# Verify containers are running
docker compose ps

# View logs (if needed)
docker compose logs -f db
```

**Adminer** will be available at `http://localhost:8080`:
- System: **PostgreSQL**
- Server: **db**
- Username: **leep**
- Password: **leep_dev_pw**
- Database: **leep_dev**

### 2. Install Dependencies

```bash
# Go dependencies
go get github.com/jackc/pgx/v5/pgxpool
go get github.com/joho/godotenv
go get github.com/aws/aws-sdk-go-v2/aws
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/credentials
go get github.com/aws/aws-sdk-go-v2/service/s3

# Or use go mod tidy
go mod tidy

# Node/Prisma dependencies (already installed)
npm install
```

### 3. Run Prisma Migrations

```bash
# Generate Prisma client
npx prisma generate

# Run migrations (creates all tables)
npx prisma migrate dev --name init_schema

# Optional: Apply CHECK constraints
psql "postgresql://leep:leep_dev_pw@localhost:5432/leep_dev" < prisma/migrations/add_check_constraints.sql
```

### 4. Start the Go Server

```bash
# Run the server
go run main.go

# Server will start on http://localhost:8080
```

### 5. Test Endpoints

```bash
# Basic health check
curl http://localhost:8080/health

# Database health check
curl http://localhost:8080/health/db

# API status
curl http://localhost:8080/api/v1/status

# Legacy ping
curl http://localhost:8080/ping
```

## Verify Database Schema

### Using Adminer (Browser)

1. Go to `http://localhost:8080`
2. Login with credentials above
3. Check tables: `profiles`, `songs`, `projects`, `stems`, `comments`, `reviews`, `tips`, `events`
4. Verify enum exists: `user_role` (fan, artist, producer, admin)

### Using Prisma Studio

```bash
npx prisma studio
# Opens at http://localhost:5555
```

### Using SQL

```bash
# Connect via psql
psql "postgresql://leep:leep_dev_pw@localhost:5432/leep_dev"

# List all tables
\dt

# Check enum values
SELECT enum_range(NULL::user_role);

# View profiles structure
\d profiles

# Test constraints (should fail)
INSERT INTO reviews (rating) VALUES (6);  -- rating must be 1-5
INSERT INTO tips (amount_cents) VALUES (0);  -- must be > 0
INSERT INTO events (event_type) VALUES ('invalid');  -- must be 'play' or 'view'
```

## Production Setup (DigitalOcean)

### 1. Provision Managed Postgres

1. Log into DigitalOcean
2. Navigate to **Databases**  **Create Database**
3. Choose **PostgreSQL 16**
4. Select your region (e.g., NYC3)
5. Choose plan (start with Basic/Dev for staging)
6. Name it: `leep-postgres-prod`
7. Click **Create**
8. **Copy connection string** (will look like):
   ```
   postgresql://doadmin:PASSWORD@host-cluster.db.ondigitalocean.com:25060/defaultdb?sslmode=require
   ```

### 2. Configure Production Database

```bash
# Set production DATABASE_URL
export DATABASE_URL="postgresql://doadmin:PASSWORD@host-cluster.db.ondigitalocean.com:25060/defaultdb?sslmode=require&schema=public"

# Run migrations to production
npx prisma migrate deploy

# Apply CHECK constraints
psql "$DATABASE_URL" < prisma/migrations/add_check_constraints.sql

# Verify connection
npx prisma studio
```

### 3. Create DigitalOcean Spaces Bucket

1. Navigate to **Spaces**  **Create Space**
2. Choose region (same as database for low latency)
3. Enable **CDN** (optional but recommended)
4. Restrict File Listing: **Yes**
5. Name: `leep-audio`
6. Click **Create Space**

**Create folder structure** (via web UI or CLI):
```
leep-audio/
 audio/
    (user uploads will be in audio/<user_id>/<uuid>.mp3)
 artwork/
     (artwork uploads in artwork/<user_id>/<uuid>.jpg)
```

### 4. Generate Spaces Access Keys

1. Go to **API**  **Spaces Keys**
2. Click **Generate New Key**
3. Name: `leep-backend-prod`
4. **Copy Access Key and Secret** (secret only shown once!)

### 5. Create Production Environment File

Create `.env.production` (NOT committed to git):

```bash
NODE_ENV=production
PORT=8080

# Production Database
DATABASE_URL="postgresql://doadmin:PASSWORD@host.db.ondigitalocean.com:25060/defaultdb?sslmode=require&schema=public"

# Spaces Configuration
SPACES_ENDPOINT=https://nyc3.digitaloceanspaces.com
SPACES_REGION=us-east-1
SPACES_BUCKET=leep-audio
SPACES_KEY=YOUR_ACCESS_KEY_HERE
SPACES_SECRET=YOUR_SECRET_KEY_HERE
```

### 6. Test Production Connection Locally

```bash
# Load production env and test connection
export $(cat .env.production | xargs)
go run main.go

# Test endpoints
curl http://localhost:8080/health/db
```

## Week 1 Deliverables Checklist

### Local Environment 
- [x] Docker Postgres running on port 5432
- [x] Adminer UI accessible
- [x] Prisma schema mirrors Supabase design
- [x] All tables created with proper constraints
- [x] Go server connects to database
- [x] Health checks working (`/health`, `/health/db`)

### Production Prep 
- [ ] DigitalOcean Managed Postgres provisioned
- [ ] Migrations deployed to production
- [ ] Spaces bucket created (`leep-audio`)
- [ ] Access keys generated (stored securely)
- [ ] `.env.production` created (not committed)
- [ ] Production connection tested

## Common Issues & Solutions

### Docker Port Already in Use
```bash
# Change host port in docker-compose.yml
ports:
  - "5433:5432"  # Use 5433 instead of 5432

# Update DATABASE_URL accordingly
```

### Prisma Can't Connect
```bash
# Ensure Docker is running
docker compose ps

# Check DATABASE_URL in .env
cat .env | grep DATABASE_URL

# Test direct connection
psql "postgresql://leep:leep_dev_pw@localhost:5432/leep_dev"
```

### Go Module Errors
```bash
# Clean and re-download modules
go clean -modcache
go mod download
go mod tidy
```

### Spaces 403 Forbidden
- Verify `SPACES_KEY` and `SPACES_SECRET` are correct
- Check endpoint matches your region
- Ensure `UsePathStyle = true` in Go client

## Next Steps (Week 2)

Once Week 1 is complete, Week 2 will focus on:
- JWT authentication middleware
- Role-based access control (RBAC)
- Song upload endpoints
- Project/collaboration endpoints
- Integration with frontend

## Resources

- [Prisma Documentation](https://www.prisma.io/docs)
- [pgx Documentation](https://pkg.go.dev/github.com/jackc/pgx/v5)
- [DigitalOcean Spaces API](https://docs.digitalocean.com/products/spaces/)
- [Supabase Schema Reference](../docs/)
