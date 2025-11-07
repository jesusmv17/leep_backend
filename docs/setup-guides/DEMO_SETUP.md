# Fresh Machine Demo Setup - Complete Guide

**Purpose**: Get from zero to working demo on a brand new machine
**Time Required**: 45-60 minutes (with downloads)
**For**: Tomorrow's sponsor demo

---

## 📋 Pre-Demo Checklist

**Before you start, ensure you have:**
- [ ] Admin access to the new machine
- [ ] Internet connection (for downloads)
- [ ] GitHub credentials ready
- [ ] 45-60 minutes available

---

## 🚀 QUICK START (Step-by-Step)

### **PHASE 1: Install Prerequisites (25-30 min)**

#### Step 1: Install WSL2 (Windows Only) - 10 min

**If Windows 11:**
```powershell
# Open PowerShell as Administrator
wsl --install

# Restart computer when prompted
# After restart, set up Ubuntu username/password
```

**If Windows 10:**
```powershell
# Enable WSL
dism.exe /online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart

# Enable Virtual Machine Platform
dism.exe /online /enable-feature /featurename:VirtualMachinePlatform /all /norestart

# Restart computer

# Download and install WSL2 kernel update
# https://aka.ms/wsl2kernel

# Set WSL2 as default
wsl --set-default-version 2

# Install Ubuntu from Microsoft Store
```

---

#### Step 2: Install Docker Desktop - 10 min

**Download**:
- For AMD/Intel CPUs: https://desktop.docker.com/win/main/amd64/Docker%20Desktop%20Installer.exe
- For Apple Silicon: https://desktop.docker.com/mac/main/arm64/Docker.dmg

**Install**:
1. Run installer as Administrator
2. ✅ Check "Use WSL 2 instead of Hyper-V" (Windows)
3. ✅ Check "Add shortcut to desktop"
4. Click Install
5. **Restart computer if prompted**
6. Launch Docker Desktop
7. **Skip account creation** (click "Continue without signing in")

**Configure Docker Desktop**:
1. Go to Settings (gear icon)
2. Resources → WSL Integration
3. ✅ Enable integration with my default WSL distro
4. ✅ Enable integration with Ubuntu (check the box)
5. Click "Apply & Restart"
6. **Wait for whale icon to be steady** (not animated)

**Time**: ~10 minutes (including restart)

---

#### Step 3: Install Go - 5 min

**Windows (via WSL):**
```bash
# In WSL/Ubuntu terminal:
cd ~
wget https://go.dev/dl/go1.23.3.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.3.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify
go version
# Should show: go version go1.23.3 linux/amd64
```

**Mac:**
```bash
# Download from: https://go.dev/dl/go1.23.3.darwin-amd64.pkg
# Or use Homebrew:
brew install go

# Verify
go version
```

**Time**: ~5 minutes

---

#### Step 4: Install Node.js - 5 min

**Ubuntu/WSL:**
```bash
# Install nvm (Node Version Manager)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.5/install.sh | bash

# Restart terminal or run:
source ~/.bashrc

# Install Node.js LTS
nvm install --lts
nvm use --lts

# Verify
node --version  # Should be v20.x or higher
npm --version   # Should be v10.x or higher
```

**Mac:**
```bash
brew install node

# Verify
node --version
npm --version
```

**Time**: ~5 minutes

---

### **PHASE 2: Clone & Setup Project (10-15 min)**

#### Step 5: Clone Repository - 2 min

```bash
# Navigate to where you want the project
cd ~  # or wherever you prefer

# Clone the repo
git clone https://github.com/jesusmv17/Leep_Backend.git
cd Leep_Backend

# Switch to your branch
git checkout chandler-branch

# Verify you're on the right branch
git branch
# Should show: * chandler-branch
```

---

#### Step 6: Fix Docker Permissions (WSL) - 2 min

```bash
# Add yourself to docker group
sudo usermod -aG docker $USER

# Apply the group
newgrp docker

# Verify docker works
docker ps
# Should show empty list (no errors)
```

---

#### Step 7: Install Project Dependencies - 5 min

```bash
# Install Node dependencies (for Prisma)
npm install

# Install Go dependencies
go mod tidy

# Verify installations
npx prisma --version  # Should show 6.18.0 or higher
go version            # Should show go1.23.3
```

---

### **PHASE 3: Start Services (5-10 min)**

#### Step 8: Start Docker Containers - 3 min

```bash
# Start Postgres + Adminer
docker compose up -d

# Wait 10 seconds for health checks
sleep 10

# Verify containers are running
docker compose ps
# Should show:
# leep_db       Up (healthy)
# leep_adminer  Up
```

**Troubleshooting**:
- If you see "permission denied": Run `newgrp docker` again
- If containers fail: Check Docker Desktop is running (whale icon steady)

---

#### Step 9: Setup Database - 3 min

```bash
# Generate Prisma client
npx prisma generate

# Apply database migrations
npx prisma migrate dev --name init_schema
# Press Enter when prompted for migration name

# Apply CHECK constraints
docker compose exec -T db psql -U leep -d leep_dev < prisma/migrations/add_check_constraints.sql
```

**Expected Output**:
- "✔ Generated Prisma Client"
- "Your database is now in sync with your schema"
- "ALTER TABLE" (3 times for constraints)

---

#### Step 10: Start Go Server - 1 min

```bash
# Start the backend server
go run main.go
```

**Expected Output**:
```
2025/11/XX XX:XX:XX ✓ Database connection established
[GIN-debug] GET /health --> ...
[GIN-debug] GET /health/db --> ...
2025/11/XX XX:XX:XX 🚀 Server starting on port 3000
```

**Keep this terminal running!**

---

### **PHASE 4: Verify Everything Works (5 min)**

#### Step 11: Test API Endpoints - 2 min

**Open a NEW terminal** and run:

```bash
# Test basic health
curl http://localhost:3000/health
# Expected: {"service":"leep-backend","status":"ok",...}

# Test database health
curl http://localhost:3000/health/db
# Expected: {"database":"connected","pool":{...},"status":"ok",...}

# Test API status
curl http://localhost:3000/api/v1/status
# Expected: {"service":"leep-backend","status":"operational",...}

# Test ping
curl http://localhost:3000/ping
# Expected: {"message":"pong"}
```

**All 4 should return JSON** ✅

---

#### Step 12: Verify Adminer - 2 min

**Open browser:**
```
http://localhost:8080
```

**Login**:
- System: **PostgreSQL**
- Server: **db**
- Username: **leep**
- Password: **leep_dev_pw**
- Database: **leep_dev**

**Verify**: You should see **10 tables** in left sidebar:
- profiles
- songs
- projects
- project_invitations
- stems
- comments
- reviews
- tips
- events
- _prisma_migrations

---

#### Step 13: Final Checklist - 1 min

**Verify all systems:**
- [ ] Docker containers running (`docker compose ps`)
- [ ] Go server running (terminal shows "Server starting")
- [ ] All 4 curl endpoints return JSON
- [ ] Adminer shows 10 tables
- [ ] Can click into `profiles` table and see structure

**If all checked** ✅ → **DEMO READY!**

---

## 🎯 DEMO DAY QUICK START (10 minutes)

**If machine is already set up from previous day:**

```bash
# 1. Start Docker Desktop
# (Click icon, wait for whale to be steady)

# 2. Navigate to project
cd ~/Leep_Backend

# 3. Start containers
docker compose up -d

# 4. Wait for health check
sleep 10

# 5. Start server
go run main.go
# Keep this running

# 6. In NEW terminal - test
curl http://localhost:3000/health/db

# 7. Open Adminer
# Browser: http://localhost:8080
# Login with credentials above

# ✅ READY TO DEMO!
```

---

## ⚡ SPEED RUN (Absolute Minimum - 30 min)

**If you're SHORT on time:**

1. **Install Docker Desktop** (10 min) - REQUIRED
2. **Install Go** (5 min) - REQUIRED
3. **Install Node** (5 min) - REQUIRED
4. **Clone repo + setup** (5 min)
5. **Start everything** (5 min)

**Skip**: Detailed verification, just check endpoints work

---

## 🚨 Common Issues & Fixes

### Issue 1: Docker Permission Denied
```bash
# Solution:
sudo usermod -aG docker $USER
newgrp docker
```

### Issue 2: Port 3000 Already in Use
```bash
# Check what's using it:
lsof -i :3000

# Kill the process:
kill -9 <PID>

# Or change port in .env:
echo "PORT=3001" >> .env
```

### Issue 3: Database Connection Failed
```bash
# Check containers:
docker compose ps

# Restart containers:
docker compose down
docker compose up -d
sleep 10
```

### Issue 4: Prisma Migration Fails
```bash
# Reset database:
docker compose down -v
docker compose up -d
sleep 10
npx prisma migrate dev --name init_schema
```

### Issue 5: WSL2 Not Working
```bash
# Update WSL:
wsl --update

# Restart WSL:
wsl --shutdown
# Wait 10 seconds, then reopen terminal
```

---

## 📱 BACKUP PLAN (If Setup Fails)

**If you CAN'T get it running on new machine:**

1. **Use current machine via remote desktop**
   - Set up Chrome Remote Desktop
   - Demo from your current working setup

2. **Use screenshots/recording**
   - Take screenshots NOW of everything working
   - Record a video walkthrough (5 min)
   - Show screenshots during demo

3. **Focus on code walkthrough**
   - Show GitHub code structure
   - Walk through Prisma schema
   - Explain architecture
   - Show documentation

---

## 📸 Take These Screenshots NOW (Backup)

**On your CURRENT working machine, screenshot:**

1. Terminal with `docker compose ps` (showing containers running)
2. Terminal with `go run main.go` (showing successful start)
3. All 4 curl commands and their JSON responses
4. Adminer login page
5. Adminer showing all 10 tables
6. Adminer showing `profiles` table structure
7. Adminer schema diagram (if available)
8. GitHub chandler-branch file structure
9. VS Code showing internal/ packages
10. SETUP.md open in browser/editor

**Save these to USB drive or cloud!**

---

## ⏱️ Time Breakdown

| Phase | Task | Time | Can Skip? |
|-------|------|------|-----------|
| 1 | WSL2 Install | 10 min | No (Windows) |
| 1 | Docker Install | 10 min | No |
| 1 | Go Install | 5 min | No |
| 1 | Node Install | 5 min | No |
| 2 | Clone Repo | 2 min | No |
| 2 | Fix Permissions | 2 min | No |
| 2 | Install Dependencies | 5 min | No |
| 3 | Start Docker | 3 min | No |
| 3 | Setup Database | 3 min | No |
| 3 | Start Server | 1 min | No |
| 4 | Test Endpoints | 2 min | Yes (risky) |
| 4 | Verify Adminer | 2 min | Yes (risky) |
| **TOTAL** | | **50 min** | **39 min minimum** |

---

## ✅ Final Pre-Demo Checklist

**Night Before:**
- [ ] Screenshots taken and backed up
- [ ] GitHub chandler-branch is up to date
- [ ] DEMO_GUIDE.md reviewed
- [ ] Practice demo script 2-3 times

**Demo Day (1 hour before):**
- [ ] All prerequisites installed
- [ ] Repo cloned and on chandler-branch
- [ ] Docker containers running
- [ ] Go server running
- [ ] All endpoints tested
- [ ] Adminer accessible with 10 tables visible
- [ ] Browser tabs ready (GitHub, Adminer, docs)

---

## 🎯 YOU'RE READY!

**Follow this guide step-by-step and you'll have a working demo in under an hour.**

**Good luck! 🚀**
