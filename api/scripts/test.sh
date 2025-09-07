#!/bin/bash
set -e

if [ -z "$SKIP_API_WAIT" ]; then
    echo "Waiting for API server to be ready..."
    until wget --no-verbose --tries=1 --spider http://api:8000/health >/dev/null 2>&1; do 
        sleep 1
    done
else
    echo "SKIP_API_WAIT set; not waiting for API"
fi

echo "Running tests..."

# Set defaults
PKG="${PKG:-./...}"
RUN_FLAGS=""

if [ -n "$TEST_PATTERN" ]; then
    echo "Running only tests matching: $TEST_PATTERN"
    RUN_FLAGS="-run $TEST_PATTERN"
fi

# Expand package list
set -e
pkgs=$(go list "$PKG")
set +e

exit_code=0

if [ -n "$FAST" ]; then
    echo "FAST mode: single go test invocation, no race/coverage"
    go test -v -failfast $FLAGS $RUN_FLAGS $pkgs
    exit_code=$?
else
    # Clean old per-package coverage
    rm -f coverage.*.out 2>/dev/null || true
    mkdir -p .coverage
    
    for p in $pkgs; do
        echo ">> Testing $p"
        cov_file=".coverage/coverage.$(echo $p | tr '/.' '--').out"
        # -failfast only affects *within* this package; loop handles cross-package bail-out.
        go test -v -failfast -covermode=atomic -coverprofile="$cov_file" $FLAGS $RUN_FLAGS "$p"
        code=$?
        if [ $code -ne 0 ]; then
            echo "❌ Tests failed in $p (exit $code)"
            exit_code=$code
            break
        fi
    done
fi

echo "Tests completed with exit code: $exit_code"
exit $exit_code
