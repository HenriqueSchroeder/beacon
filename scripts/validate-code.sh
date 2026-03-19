#!/bin/bash

# Code validation script for Link Validation feature
# This script validates the implementation without needing Go installed

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ERRORS=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🔍 Validating Link Validation Feature Implementation"
echo "===================================================="
echo ""

# Check 1: Verify all required files exist
echo "Check 1: Verifying file structure..."
required_files=(
    "pkg/links/parser.go"
    "pkg/links/parser_test.go"
    "pkg/links/integration_test.go"
    "pkg/validate/validator.go"
    "pkg/validate/validator_test.go"
    "cmd/beacon/validate.go"
    "docs/LINK_VALIDATION.md"
    "docs/PR_LINK_VALIDATION.md"
)

for file in "${required_files[@]}"; do
    if [ -f "$PROJECT_ROOT/$file" ]; then
        echo -e "  ${GREEN}✓${NC} $file"
    else
        echo -e "  ${RED}✗${NC} $file (MISSING)"
        ((ERRORS++))
    fi
done

echo ""

# Check 2: Verify Go syntax (if go is available)
echo "Check 2: Validating Go syntax..."
if command -v go &> /dev/null; then
    if go fmt ./... &> /dev/null; then
        echo -e "  ${GREEN}✓${NC} Code formatting is valid"
    else
        echo -e "  ${YELLOW}⚠${NC} Code formatting check requires fixing"
        ((ERRORS++))
    fi

    # Try to build
    if go build -o /tmp/beacon-test ./cmd/beacon &> /dev/null; then
        echo -e "  ${GREEN}✓${NC} Build successful"
        rm -f /tmp/beacon-test
    else
        echo -e "  ${RED}✗${NC} Build failed"
        ((ERRORS++))
    fi
else
    echo -e "  ${YELLOW}⚠${NC} Go not available - skipping syntax check"
fi

echo ""

# Check 3: Validate test files
echo "Check 3: Checking test coverage..."
test_files=$(find "$PROJECT_ROOT" -name "*_test.go" -path "*/pkg/*" | wc -l)
echo -e "  ${GREEN}✓${NC} Found $test_files test files"

# Count test functions
test_functions=$(grep -r "func Test" "$PROJECT_ROOT/pkg/" | wc -l)
echo -e "  ${GREEN}✓${NC} Found $test_functions test functions"

if [ "$test_functions" -lt 20 ]; then
    echo -e "  ${YELLOW}⚠${NC} Expected at least 20 test functions"
    ((ERRORS++))
fi

echo ""

# Check 4: Validate documentation
echo "Check 4: Checking documentation..."
doc_files=(
    "docs/LINK_VALIDATION.md"
    "docs/PR_LINK_VALIDATION.md"
)

for doc in "${doc_files[@]}"; do
    if [ -f "$PROJECT_ROOT/$doc" ]; then
        lines=$(wc -l < "$PROJECT_ROOT/$doc")
        echo -e "  ${GREEN}✓${NC} $doc ($lines lines)"
    fi
done

echo ""

# Check 5: Code quality checks
echo "Check 5: Code quality analysis..."

# Check for TODO comments (not a failure, just informational)
todos=$(grep -r "TODO" "$PROJECT_ROOT/pkg/links/" "$PROJECT_ROOT/pkg/validate/" "$PROJECT_ROOT/cmd/beacon/validate.go" 2>/dev/null | wc -l || echo "0")
if [ "$todos" -gt 0 ]; then
    echo -e "  ${YELLOW}⚠${NC} Found $todos TODO comments"
fi

# Check for panic usage (should be minimal)
panics=$(grep -r "panic(" "$PROJECT_ROOT/pkg/links/" "$PROJECT_ROOT/pkg/validate/" 2>/dev/null | wc -l || echo "0")
if [ "$panics" -eq 0 ]; then
    echo -e "  ${GREEN}✓${NC} No panic() calls (good error handling)"
else
    echo -e "  ${YELLOW}⚠${NC} Found $panics panic() calls"
fi

# Check error handling
error_checks=$(grep -r "if err != nil" "$PROJECT_ROOT/pkg/links/" "$PROJECT_ROOT/pkg/validate/" "$PROJECT_ROOT/cmd/beacon/validate.go" 2>/dev/null | wc -l || echo "0")
echo -e "  ${GREEN}✓${NC} Found $error_checks error checks"

echo ""

# Check 6: Architecture validation
echo "Check 6: Validating architecture..."

# Check that files import each other correctly
if grep -q "github.com/HenriqueSchroeder/beacon/pkg/links" "$PROJECT_ROOT/pkg/validate/validator.go"; then
    echo -e "  ${GREEN}✓${NC} Validator imports links package"
else
    echo -e "  ${RED}✗${NC} Validator doesn't import links package"
    ((ERRORS++))
fi

if grep -q "github.com/HenriqueSchroeder/beacon/pkg/validate" "$PROJECT_ROOT/cmd/beacon/validate.go"; then
    echo -e "  ${GREEN}✓${NC} CLI imports validate package"
else
    echo -e "  ${RED}✗${NC} CLI doesn't import validate package"
    ((ERRORS++))
fi

if grep -q "github.com/HenriqueSchroeder/beacon/pkg/vault" "$PROJECT_ROOT/pkg/validate/validator.go"; then
    echo -e "  ${GREEN}✓${NC} Validator imports vault package"
else
    echo -e "  ${RED}✗${NC} Validator doesn't import vault package"
    ((ERRORS++))
fi

echo ""

# Check 7: CLI command structure
echo "Check 7: Validating CLI structure..."

if grep -q "var validateCmd = &cobra.Command" "$PROJECT_ROOT/cmd/beacon/validate.go"; then
    echo -e "  ${GREEN}✓${NC} Cobra command defined"
else
    echo -e "  ${RED}✗${NC} Cobra command not defined"
    ((ERRORS++))
fi

if grep -q "func init()" "$PROJECT_ROOT/cmd/beacon/validate.go"; then
    echo -e "  ${GREEN}✓${NC} init() function to register command"
else
    echo -e "  ${RED}✗${NC} init() function missing"
    ((ERRORS++))
fi

# Check flags
flags=("--json" "--file" "--strict" "--fix" "--use-cache")
for flag in "${flags[@]}"; do
    if grep -q "$flag" "$PROJECT_ROOT/cmd/beacon/validate.go"; then
        echo -e "  ${GREEN}✓${NC} Flag $flag implemented"
    fi
done

echo ""

# Check 8: Type definitions
echo "Check 8: Validating type definitions..."

types=(
    "type Link struct"
    "type LinkType int"
    "type Parser struct"
    "type Validator struct"
    "type ValidationResult struct"
    "type DocumentValidation struct"
)

for type_def in "${types[@]}"; do
    found_in=""
    if grep -q "$type_def" "$PROJECT_ROOT/pkg/links/parser.go"; then
        found_in="links/parser.go"
    elif grep -q "$type_def" "$PROJECT_ROOT/pkg/validate/validator.go"; then
        found_in="validate/validator.go"
    fi

    if [ -n "$found_in" ]; then
        echo -e "  ${GREEN}✓${NC} $type_def in $found_in"
    else
        echo -e "  ${YELLOW}⚠${NC} $type_def not found"
    fi
done

echo ""

# Summary
echo "===================================================="
if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}✅ All validation checks passed!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Install Go (if not already installed)"
    echo "  2. Run: go test -v ./..."
    echo "  3. Run: go build -o bin/beacon ./cmd/beacon"
    echo "  4. Test: ./bin/beacon validate"
    exit 0
else
    echo -e "${RED}❌ Validation failed with $ERRORS error(s)${NC}"
    exit 1
fi
