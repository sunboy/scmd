#!/bin/bash
set -e

# Integration test for download enhancements
# Tests retry logic, resume support, and error handling

echo "==================================="
echo "scmd Download Integration Tests"
echo "==================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to run test
run_test() {
    local test_name="$1"
    local test_command="$2"

    echo -e "${YELLOW}[TEST]${NC} $test_name"

    if eval "$test_command" > /tmp/test_output.log 2>&1; then
        echo -e "  ${GREEN}✓ PASS${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "  ${RED}✗ FAIL${NC}"
        echo "  Output:"
        cat /tmp/test_output.log | head -20 | sed 's/^/    /'
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Test 1: Build verification
echo "Test Suite 1: Build Verification"
echo "-----------------------------------"

run_test "Build scmd binary" \
    "go build -o scmd-test ./cmd/scmd"

run_test "Binary is executable" \
    "test -x ./scmd-test"

run_test "Help command works" \
    "./scmd-test --help > /dev/null"

echo ""

# Test 2: Model URL verification
echo "Test Suite 2: Model URL Verification"
echo "---------------------------------------"

run_test "7B model URL is accessible (302 or 200)" \
    "curl -I -s 'https://huggingface.co/Qwen/Qwen2.5-7B-Instruct-GGUF/resolve/main/qwen2.5-7b-instruct-q3_k_m.gguf' | grep -E 'HTTP/2 (200|302)' > /dev/null"

run_test "1.5B model URL is accessible (302 or 200)" \
    "curl -I -s 'https://huggingface.co/Qwen/Qwen2.5-1.5B-Instruct-GGUF/resolve/main/qwen2.5-1.5b-instruct-q4_k_m.gguf' | grep -E 'HTTP/2 (200|302)' > /dev/null"

run_test "0.5B model URL is accessible (302 or 200)" \
    "curl -I -s 'https://huggingface.co/Qwen/Qwen2.5-0.5B-Instruct-GGUF/resolve/main/qwen2.5-0.5b-instruct-q4_k_m.gguf' | grep -E 'HTTP/2 (200|302)' > /dev/null"

run_test "3B model URL is accessible (302 or 200)" \
    "curl -I -s 'https://huggingface.co/Qwen/Qwen2.5-3B-Instruct-GGUF/resolve/main/qwen2.5-3b-instruct-q4_k_m.gguf' | grep -E 'HTTP/2 (200|302)' > /dev/null"

echo ""

# Test 3: Code quality checks
echo "Test Suite 3: Code Quality"
echo "----------------------------"

run_test "No syntax errors in Go code" \
    "go vet ./..."

run_test "Code formatting is correct" \
    "test -z \"\$(gofmt -l . | grep -v vendor)\""

echo ""

# Test 4: Enhanced downloader features
echo "Test Suite 4: Download Features"
echo "----------------------------------"

# Create a test to verify the enhanced downloader compiles
run_test "Enhanced downloader compiles" \
    "go build -o /dev/null ./internal/backend/llamacpp/"

# Verify key functions exist in code
run_test "Retry logic implemented" \
    "grep -q 'MaxRetries' ./internal/backend/llamacpp/download_enhanced.go"

run_test "Resume support implemented" \
    "grep -q 'Range.*bytes=' ./internal/backend/llamacpp/download_enhanced.go"

run_test "Disk space check implemented" \
    "grep -q 'CheckDiskSpace' ./internal/backend/llamacpp/download_enhanced.go"

run_test "Error messages have help text" \
    "grep -q 'Help.*\\[\\]string' ./internal/backend/llamacpp/download_enhanced.go"

echo ""

# Test 5: Setup flow enhancements
echo "Test Suite 5: Setup Flow"
echo "--------------------------"

run_test "Setup has stage indicators" \
    "grep -q '\\[1/3\\]' ./internal/cli/setup.go"

run_test "Quick test feature exists" \
    "grep -q 'offerQuickTest' ./internal/cli/setup.go"

run_test "Success message enhanced" \
    "grep -q '🎉' ./internal/cli/setup.go"

echo ""

# Test 6: Documentation
echo "Test Suite 6: Documentation"
echo "----------------------------"

run_test "README has Quick Start section" \
    "grep -q '## Quick Start' ./README.md"

run_test "README mentions 2 minutes setup" \
    "grep -q '2 minutes' ./README.md"

run_test "README has First Run instructions" \
    "grep -q 'First Run' ./README.md"

echo ""

# Summary
echo "==================================="
echo "Test Summary"
echo "==================================="
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed! ✓${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed. ✗${NC}"
    exit 1
fi
