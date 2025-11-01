# Leep Audio Backend - Makefile
# Simplifies common development tasks

.PHONY: help setup docker-up docker-down db-migrate db-reset db-studio dev build test clean

# Default target
help:
	@echo "Leep Audio Backend - Available Commands:"
	@echo ""
	@echo "  make setup         - Initial project setup (install deps)"
	@echo "  make docker-up     - Start Docker containers (Postgres + Adminer)"
	@echo "  make docker-down   - Stop Docker containers"
	@echo "  make db-migrate    - Run Prisma migrations (dev)"
	@echo "  make db-deploy     - Deploy migrations (production)"
	@echo "  make db-reset      - Reset database (WARNING: deletes all data)"
	@echo "  make db-studio     - Open Prisma Studio"
	@echo "  make dev           - Run development server"
	@echo "  make build         - Build Go binary"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts"
	@echo ""

# Initial setup
setup:
	@echo "📦 Installing dependencies..."
	npm install
	go mod download
	go mod tidy
	@echo "✅ Dependencies installed"

# Docker commands
docker-up:
	@echo "🐳 Starting Docker containers..."
	docker compose up -d
	@echo "✅ Containers running"
	@echo "   - Postgres: localhost:5432"
	@echo "   - Adminer: http://localhost:8080"

docker-down:
	@echo "🛑 Stopping Docker containers..."
	docker compose down
	@echo "✅ Containers stopped"

docker-logs:
	docker compose logs -f

# Database commands
db-migrate:
	@echo "🗄️  Running database migrations..."
	npx prisma generate
	npx prisma migrate dev
	@echo "✅ Migrations complete"

db-deploy:
	@echo "🚀 Deploying migrations to production..."
	npx prisma migrate deploy
	@echo "✅ Production migrations complete"

db-reset:
	@echo "⚠️  Resetting database (this will delete all data)..."
	npx prisma migrate reset --force
	@echo "✅ Database reset complete"

db-studio:
	@echo "🎨 Opening Prisma Studio..."
	npx prisma studio

db-seed:
	@echo "🌱 Seeding database..."
	# Add seed script here when needed
	@echo "✅ Database seeded"

# Development
dev:
	@echo "🚀 Starting development server..."
	go run main.go

# Build
build:
	@echo "🔨 Building Go binary..."
	go build -o bin/leep-backend main.go
	@echo "✅ Binary created at bin/leep-backend"

# Testing
test:
	@echo "🧪 Running tests..."
	go test -v ./...
	@echo "✅ Tests complete"

test-coverage:
	@echo "🧪 Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Linting
lint:
	@echo "🔍 Running linter..."
	golangci-lint run
	@echo "✅ Linting complete"

# Cleanup
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "✅ Cleanup complete"

# Production
prod-setup:
	@echo "🚀 Setting up production environment..."
	@echo "1. Provision DigitalOcean Managed Postgres"
	@echo "2. Create Spaces bucket"
	@echo "3. Update .env.production with credentials"
	@echo "4. Run: make db-deploy"

# Health check
health:
	@echo "🏥 Checking service health..."
	@curl -s http://localhost:8080/health | jq '.' || echo "Server not running"
	@curl -s http://localhost:8080/health/db | jq '.' || echo "Database health check failed"

# Git shortcuts
git-status:
	git status --short

git-push:
	git add .
	git status
	@echo ""
	@read -p "Commit message: " msg; \
	git commit -m "$$msg"
	git push
