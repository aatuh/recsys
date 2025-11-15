#!/bin/bash
set -e

if [ -z "$SKIP_API_WAIT" ]; then
    echo "Tip: tail api/.test_log.txt on the host if console output truncates."
    echo "Waiting for API server to be ready..."
    waited=0
    until wget --no-verbose --tries=1 --spider http://api:8000/health >/dev/null 2>&1; do
        sleep 1
        waited=$((waited + 1))
        if [ $waited -ge 30 ]; then
            echo "API server did not become ready within timeout."
            exit 1
        fi
    done
else
    echo "SKIP_API_WAIT set; not waiting for API"
fi

LOG_FILE="/app/.test_log.txt"
: > "$LOG_FILE"
echo "Streaming go test output to $LOG_FILE"
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
    set -o pipefail
    go test -v -failfast $FLAGS $RUN_FLAGS $pkgs 2>&1 | tee -a "$LOG_FILE"
    exit_code=${PIPESTATUS[0]}
    set +o pipefail
else
    # Clean old per-package coverage
    rm -f coverage.*.out 2>/dev/null || true
    mkdir -p .coverage
    
    for p in $pkgs; do
        echo ">> Testing $p"
        cov_file=".coverage/coverage.$(echo $p | tr '/.' '--').out"
        # -failfast only affects *within* this package; loop handles cross-package bail-out.
        set -o pipefail
        go test -v -failfast -covermode=atomic -coverprofile="$cov_file" $FLAGS $RUN_FLAGS "$p" 2>&1 | tee -a "$LOG_FILE"
        code=${PIPESTATUS[0]}
        set +o pipefail
        if [ $code -ne 0 ]; then
            echo "❌ Tests failed in $p (exit $code)"
            exit_code=$code
            break
        fi
    done
fi

echo "Tests completed with exit code: $exit_code"
exit $exit_code
