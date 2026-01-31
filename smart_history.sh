#!/bin/bash

# =============================================================================
# smart_history.sh - Git History Backfill Script
# Author: Odysse4s
# Description: Rewrites git history to simulate incremental development
# =============================================================================

set -e

# --- Colors for output ---
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# --- OS Detection for date command ---
detect_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        OS_TYPE="macos"
    else
        OS_TYPE="linux"
    fi
    echo -e "${BLUE}[INFO]${NC} Detected OS: $OS_TYPE"
}

# --- Date calculation function ---
# Usage: get_past_date <days_ago>
get_past_date() {
    local days_ago=$1
    if [[ "$OS_TYPE" == "macos" ]]; then
        date -v-${days_ago}d "+%Y-%m-%dT12:00:00"
    else
        date -d "$days_ago days ago" "+%Y-%m-%dT12:00:00"
    fi
}

# --- Make a backdated commit ---
# Usage: make_commit "<message>" "<date>"
make_commit() {
    local msg="$1"
    local commit_date="$2"
    
    export GIT_AUTHOR_DATE="$commit_date"
    export GIT_COMMITTER_DATE="$commit_date"
    
    git commit -m "$msg"
    
    unset GIT_AUTHOR_DATE
    unset GIT_COMMITTER_DATE
    
    echo -e "${GREEN}[COMMIT]${NC} $msg @ $commit_date"
}

# --- Main Script ---
echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}       Git History Backfill Script v2.0         ${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

detect_os

# Step 1: Get repo name from user
echo -e "${YELLOW}[INPUT]${NC} Enter your repository name (e.g., secure-storage):"
read -r REPO_NAME

if [[ -z "$REPO_NAME" ]]; then
    echo -e "${RED}[ERROR]${NC} Repository name cannot be empty!"
    exit 1
fi

GITHUB_USER="Odysse4s"
REMOTE_URL="git@github.com:${GITHUB_USER}/${REPO_NAME}.git"

echo -e "${BLUE}[INFO]${NC} Remote URL: $REMOTE_URL"

# Step 2: Safety Sync - Check remote and pull
echo ""
echo -e "${YELLOW}[STEP 1]${NC} Safety Sync - Checking remote..."

if git remote get-url origin &>/dev/null; then
    echo -e "${BLUE}[INFO]${NC} Remote 'origin' exists. Attempting to pull latest changes..."
    
    # Try to pull from main or master
    if git pull origin main --rebase 2>/dev/null; then
        echo -e "${GREEN}[OK]${NC} Pulled latest from 'main'"
    elif git pull origin master --rebase 2>/dev/null; then
        echo -e "${GREEN}[OK]${NC} Pulled latest from 'master'"
    else
        echo -e "${YELLOW}[WARN]${NC} Could not pull. Remote might be empty or unreachable. Continuing..."
    fi
else
    echo -e "${YELLOW}[WARN]${NC} No remote configured. Skipping pull."
fi

# Step 3: Destroy old history
echo ""
echo -e "${YELLOW}[STEP 2]${NC} Destroying old git history..."
rm -rf .git
echo -e "${GREEN}[OK]${NC} Old .git directory removed"

# Step 4: Initialize fresh repo
echo ""
echo -e "${YELLOW}[STEP 3]${NC} Initializing fresh repository..."
git init
git branch -M main
echo -e "${GREEN}[OK]${NC} Fresh repo initialized with 'main' branch"

# Step 5: Create backdated commits
echo ""
echo -e "${YELLOW}[STEP 4]${NC} Creating backdated commits..."

# --- Commit 1: Project Setup (30 days ago) ---
DATE_1=$(get_past_date 30)
git add go.mod
if [[ -f "go.sum" ]]; then
    git add go.sum
fi
git add .gitignore 2>/dev/null || true
make_commit "chore: initialize Go module and project structure" "$DATE_1"

# --- Commit 2: Core Storage Logic (27 days ago) ---
DATE_2=$(get_past_date 27)
git add internal/storage/storage.go
make_commit "feat: implement core AES-256 encryption storage service

- Add Service struct with GCM cipher
- Implement SaveFile with TeeReader for SHA-256 hashing
- Implement LoadFile with integrity verification
- Add filename validation against path traversal" "$DATE_2"

# --- Commit 3: Storage Tests (25 days ago) ---
DATE_3=$(get_past_date 25)
git add internal/storage/storage_test.go
make_commit "test: add unit tests for storage service

- Test filename validation patterns
- Test save/load roundtrip with encryption
- Test checksum file creation
- Test non-existent file handling" "$DATE_3"

# --- Commit 4: API Handlers (22 days ago) ---
DATE_4=$(get_past_date 22)
git add internal/api/handlers.go
make_commit "feat: implement REST API handlers

- Add UploadHandler with multipart form support
- Add DownloadHandler with streaming response
- Add HealthHandler for container health checks
- Implement JSON response helpers" "$DATE_4"

# --- Commit 5: Main Server with Rate Limiting (18 days ago) ---
DATE_5=$(get_past_date 18)
git add cmd/server/main.go
make_commit "feat: add main server with rate limiting middleware

- Implement RateLimiter struct with per-IP tracking
- Add rate limit of 1 req/sec with burst of 3
- Setup HTTP routes for upload, download, health
- Add automatic visitor cleanup goroutine" "$DATE_5"

# --- Commit 6: Dockerfile (14 days ago) ---
DATE_6=$(get_past_date 14)
git add Dockerfile
make_commit "build: add multi-stage Dockerfile

- Use golang:1.21-alpine as builder
- Create minimal alpine:3.19 production image
- Add non-root user for security
- Configure health check endpoint" "$DATE_6"

# --- Commit 7: Docker Compose (10 days ago) ---
DATE_7=$(get_past_date 10)
git add docker-compose.yml
make_commit "build: add docker-compose configuration

- Configure secure-storage-api service
- Mount persistent data volume
- Add security hardening (cap_drop, read_only)
- Configure tmpfs for /tmp" "$DATE_7"

# --- Commit 8: Makefile (7 days ago) ---
DATE_8=$(get_past_date 7)
git add Makefile
make_commit "build: add Makefile with common tasks

- Add build, run, test targets
- Add docker-build and docker-run targets
- Add clean target for artifacts" "$DATE_8"

# --- Commit 9: Documentation (4 days ago) ---
DATE_9=$(get_past_date 4)
git add README.md
make_commit "docs: add comprehensive README

- Document API endpoints and usage
- Add installation instructions
- Include security considerations
- Add example curl commands" "$DATE_9"

# --- Commit 10: Final polish (1 day ago) ---
DATE_10=$(get_past_date 1)
git add -A
# This catches any remaining files we might have missed
if ! git diff --cached --quiet; then
    make_commit "chore: final cleanup and polish" "$DATE_10"
else
    echo -e "${BLUE}[INFO]${NC} No additional files to commit"
fi

# Step 6: Add remote
echo ""
echo -e "${YELLOW}[STEP 5]${NC} Adding remote origin..."
git remote add origin "$REMOTE_URL"
echo -e "${GREEN}[OK]${NC} Remote added: $REMOTE_URL"

# Step 7: Show summary
echo ""
echo -e "${BLUE}================================================${NC}"
echo -e "${GREEN}            History Rewrite Complete!           ${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""
echo -e "${BLUE}[INFO]${NC} Commit log:"
git log --oneline --all
echo ""
echo -e "${YELLOW}[ACTION REQUIRED]${NC} To push to GitHub, run:"
echo ""
echo -e "    ${GREEN}git push -u origin main --force${NC}"
echo ""
echo -e "${RED}[WARNING]${NC} This will overwrite the remote history!"
echo ""
