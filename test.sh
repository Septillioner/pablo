#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
MODE="${1:-all}"

usage() {
  cat <<'EOF'
Usage: ./test.sh [unit|integration|e2e|all]

  unit         Run unit tests in src/
  integration  Run integration-tagged tests in src/
  e2e          Run Docker-based E2E tests in tests/e2e/
  all          Run unit, then integration, then e2e (default)
EOF
}

require_docker() {
  if [ "${PABLO_E2E_SKIP_DOCKER:-}" = "1" ]; then
    return 0
  fi
  if ! command -v docker >/dev/null 2>&1; then
    echo "error: docker is required for e2e tests (set PABLO_E2E_SKIP_DOCKER=1 to skip Docker setup)" >&2
    exit 1
  fi
  if ! docker info >/dev/null 2>&1; then
    echo "error: docker daemon is not running" >&2
    exit 1
  fi
}

run_unit() {
  echo "==> unit tests (src/)"
  (cd "$ROOT/src" && go test ./...)
}

run_integration() {
  echo "==> integration tests (src/, -tags=integration)"
  (cd "$ROOT/src" && go test -tags=integration ./...)
}

run_e2e() {
  echo "==> e2e tests (tests/e2e/)"
  require_docker
  (cd "$ROOT/tests/e2e" && go test -tags=integration -v -timeout 10m ./...)
}

case "$MODE" in
  unit)
    run_unit
    ;;
  integration)
    run_integration
    ;;
  e2e)
    run_e2e
    ;;
  all)
    run_unit
    run_integration
    run_e2e
    ;;
  -h|--help|help)
    usage
    ;;
  *)
    echo "error: unknown mode '$MODE'" >&2
    usage >&2
    exit 1
    ;;
esac
