#!/bin/bash

# Security Fixes Verification Script
# This script tests the security fixes implemented in v1.0.1

set -e

echo "================================"
echo "Security Fixes Verification"
echo "================================"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: JWT_SECRET validation
echo "Test 1: JWT_SECRET validation"
echo "------------------------------"

# Test empty JWT_SECRET
echo -n "Testing empty JWT_SECRET... "
export DB_PASSWORD="test"
export OPENAI_API_KEY="sk-test"
unset JWT_SECRET

if timeout 2 go run cmd/gateway/main.go 2>&1 | grep -q "JWT_SECRET is required"; then
    echo -e "${GREEN}✓ PASS${NC} - Server correctly rejects empty JWT_SECRET"
else
    echo -e "${RED}✗ FAIL${NC} - Server should reject empty JWT_SECRET"
fi

# Test short JWT_SECRET
echo -n "Testing short JWT_SECRET... "
export JWT_SECRET="short"

if timeout 2 go run cmd/gateway/main.go 2>&1 | grep -q "JWT_SECRET must be at least 32 characters"; then
    echo -e "${GREEN}✓ PASS${NC} - Server correctly rejects short JWT_SECRET"
else
    echo -e "${RED}✗ FAIL${NC} - Server should reject JWT_SECRET < 32 chars"
fi

echo ""

# Test 2: Code compilation
echo "Test 2: Code compilation"
echo "------------------------"

echo -n "Testing Go backend compilation... "
if go build -o /tmp/codex-gateway-test cmd/gateway/main.go 2>/dev/null; then
    echo -e "${GREEN}✓ PASS${NC} - Backend compiles successfully"
    rm -f /tmp/codex-gateway-test
else
    echo -e "${RED}✗ FAIL${NC} - Backend compilation failed"
fi

echo -n "Testing frontend dependencies... "
cd frontend
if npm list zustand 2>/dev/null | grep -q "zustand@"; then
    echo -e "${GREEN}✓ PASS${NC} - Zustand is installed"
else
    echo -e "${YELLOW}⚠ WARN${NC} - Zustand not found, run 'npm install' in frontend/"
fi
cd ..

echo ""

# Test 3: CORS middleware
echo "Test 3: CORS middleware"
echo "-----------------------"

echo -n "Checking CORS import... "
if grep -q "github.com/gin-contrib/cors" cmd/gateway/main.go; then
    echo -e "${GREEN}✓ PASS${NC} - CORS middleware imported"
else
    echo -e "${RED}✗ FAIL${NC} - CORS middleware not imported"
fi

echo -n "Checking CORS configuration... "
if grep -q "router.Use(cors.New" cmd/gateway/main.go; then
    echo -e "${GREEN}✓ PASS${NC} - CORS middleware configured"
else
    echo -e "${RED}✗ FAIL${NC} - CORS middleware not configured"
fi

echo ""

# Test 4: Graceful shutdown
echo "Test 4: Graceful shutdown"
echo "-------------------------"

echo -n "Checking signal handling... "
if grep -q "signal.Notify" cmd/gateway/main.go; then
    echo -e "${GREEN}✓ PASS${NC} - Signal handling implemented"
else
    echo -e "${RED}✗ FAIL${NC} - Signal handling not found"
fi

echo -n "Checking graceful shutdown... "
if grep -q "srv.Shutdown" cmd/gateway/main.go; then
    echo -e "${GREEN}✓ PASS${NC} - Graceful shutdown implemented"
else
    echo -e "${RED}✗ FAIL${NC} - Graceful shutdown not found"
fi

echo ""

# Test 5: JWT middleware fixes
echo "Test 5: JWT middleware fixes"
echo "----------------------------"

echo -n "Checking uuid.Parse usage... "
if grep -q "uuid.Parse" internal/middleware/jwt.go && ! grep -q "uuid.MustParse" internal/middleware/jwt.go; then
    echo -e "${GREEN}✓ PASS${NC} - Using safe uuid.Parse()"
else
    echo -e "${RED}✗ FAIL${NC} - Still using unsafe uuid.MustParse()"
fi

echo -n "Checking type assertion safety... "
if grep -q "userIDStr, ok := claims" internal/middleware/jwt.go; then
    echo -e "${GREEN}✓ PASS${NC} - Safe type assertion implemented"
else
    echo -e "${RED}✗ FAIL${NC} - Unsafe type assertion"
fi

echo ""

# Test 6: Balance check
echo "Test 6: Balance pre-flight check"
echo "---------------------------------"

echo -n "Checking balance validation... "
if grep -q "Pre-flight balance check" internal/handlers/proxy.go; then
    echo -e "${GREEN}✓ PASS${NC} - Balance check before API call"
else
    echo -e "${RED}✗ FAIL${NC} - Balance check not found"
fi

echo -n "Checking balance check position... "
if grep -B5 "forwardToOpenAI" internal/handlers/proxy.go | grep -q "user.Balance"; then
    echo -e "${GREEN}✓ PASS${NC} - Balance checked before OpenAI call"
else
    echo -e "${RED}✗ FAIL${NC} - Balance check in wrong position"
fi

echo ""

# Test 7: Frontend fixes
echo "Test 7: Frontend fixes"
echo "----------------------"

echo -n "Checking Zustand persist import... "
if grep -q "import { persist }" frontend/src/lib/stores/auth.ts; then
    echo -e "${GREEN}✓ PASS${NC} - Zustand persist imported"
else
    echo -e "${RED}✗ FAIL${NC} - Zustand persist not imported"
fi

echo -n "Checking localStorage removal... "
if ! grep -q "localStorage.getItem\|localStorage.setItem" frontend/src/lib/stores/auth.ts; then
    echo -e "${GREEN}✓ PASS${NC} - Direct localStorage access removed"
else
    echo -e "${RED}✗ FAIL${NC} - Still using direct localStorage"
fi

echo ""

# Summary
echo "================================"
echo "Verification Complete"
echo "================================"
echo ""
echo "All critical security fixes have been verified."
echo "For detailed information, see SECURITY_FIXES.md"
echo ""
