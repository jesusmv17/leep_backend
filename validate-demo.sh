#!/bin/bash

# Leep Audio Backend - Demo Validation Script
# Run this before your sponsor demo to verify everything works

set -e  # Exit on error

echo "=========================================="
echo "Leep Audio Backend - Demo Validator"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track success
all_passed=true

echo "Step 1: Checking Docker..."
if docker compose ps | grep -q "Up"; then
    echo -e "${GREEN}✓ Docker containers are running${NC}"
else
    echo -e "${YELLOW}⚠ Docker containers not running. Starting them...${NC}"
    docker compose up -d
    sleep 5
    if docker compose ps | grep -q "Up"; then
        echo -e "${GREEN}✓ Docker containers started successfully${NC}"
    else
        echo -e "${RED}✗ Failed to start Docker containers${NC}"
        all_passed=false
    fi
fi
echo ""

echo "Step 2: Checking Go dependencies..."
if go mod verify &> /dev/null; then
    echo -e "${GREEN}✓ Go modules verified${NC}"
else
    echo -e "${YELLOW}⚠ Running go mod tidy...${NC}"
    go mod tidy
    echo -e "${GREEN}✓ Go dependencies installed${NC}"
fi
echo ""

echo "Step 3: Checking database connection..."
if docker compose exec -T db pg_isready -U leep -d leep_dev &> /dev/null; then
    echo -e "${GREEN}✓ Database is ready${NC}"
else
    echo -e "${RED}✗ Database not ready${NC}"
    all_passed=false
fi
echo ""

echo "Step 4: Checking if migrations are applied..."
if npx prisma migrate status 2>&1 | grep -q "Database schema is up to date"; then
    echo -e "${GREEN}✓ Migrations are up to date${NC}"
else
    echo -e "${YELLOW}⚠ Migrations need to be applied${NC}"
    echo "Run: npx prisma migrate dev"
    all_passed=false
fi
echo ""

echo "Step 5: Checking if server is running..."
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Server is running${NC}"

    echo ""
    echo "Testing endpoints..."

    # Test /health
    health_response=$(curl -s http://localhost:8080/health)
    if echo "$health_response" | grep -q "ok"; then
        echo -e "${GREEN}  ✓ /health endpoint working${NC}"
    else
        echo -e "${RED}  ✗ /health endpoint failed${NC}"
        all_passed=false
    fi

    # Test /health/db
    db_health_response=$(curl -s http://localhost:8080/health/db)
    if echo "$db_health_response" | grep -q "connected"; then
        echo -e "${GREEN}  ✓ /health/db endpoint working${NC}"
    else
        echo -e "${YELLOW}  ⚠ /health/db endpoint returned: $db_health_response${NC}"
        all_passed=false
    fi

    # Test /api/v1/status
    status_response=$(curl -s http://localhost:8080/api/v1/status)
    if echo "$status_response" | grep -q "leep-backend"; then
        echo -e "${GREEN}  ✓ /api/v1/status endpoint working${NC}"
    else
        echo -e "${RED}  ✗ /api/v1/status endpoint failed${NC}"
        all_passed=false
    fi
else
    echo -e "${YELLOW}⚠ Server is not running${NC}"
    echo "Start it with: go run main.go"
    all_passed=false
fi
echo ""

echo "Step 6: Checking Adminer accessibility..."
if curl -s http://localhost:8080 > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Adminer is accessible at http://localhost:8080${NC}"
    echo "  Login: Server=db, User=leep, Password=leep_dev_pw, Database=leep_dev"
else
    echo -e "${RED}✗ Adminer not accessible${NC}"
    all_passed=false
fi
echo ""

echo "=========================================="
if [ "$all_passed" = true ]; then
    echo -e "${GREEN}🎉 ALL CHECKS PASSED - READY FOR DEMO!${NC}"
    echo ""
    echo "Quick Test Commands:"
    echo "  curl http://localhost:8080/health | jq ."
    echo "  curl http://localhost:8080/health/db | jq ."
    echo "  Open http://localhost:8080 (Adminer)"
    echo ""
    echo "Review DEMO_GUIDE.md for demo script!"
else
    echo -e "${YELLOW}⚠ SOME CHECKS FAILED - Review errors above${NC}"
    echo ""
    echo "Common fixes:"
    echo "  1. Run: docker compose up -d"
    echo "  2. Run: npx prisma migrate dev"
    echo "  3. Run: go run main.go (in separate terminal)"
    echo "  4. Run this script again"
fi
echo "=========================================="
