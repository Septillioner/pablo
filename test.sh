#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
MODE="${1:-all}"

declare -A SUMMARY
declare -A E2E_SCENARIO_MAP=(
  [TestSSH_StaticDeploy]=ssh-static
  [TestSSH_RenameReplace]=ssh-rename-replace
  [TestSSH_DockerRemoteDeploy]=ssh-docker-remote
)

usage() {
  cat <<'EOF'
Usage: ./test.sh [unit|integration|e2e|all]

  unit         Run unit tests in src/
  integration  Run integration-tagged tests in src/
  e2e          Run Docker-based E2E tests in tests/e2e/
  all          Run unit, then integration, then e2e (default)
EOF
}

section_header() {
  echo ""
  echo "======== $1 ========"
}

json_field() {
  local json="$1" field="$2"
  printf '%s' "$json" | sed -n "s/.*\"${field}\":\"\\([^\"]*\\)\".*/\\1/p" | head -1
}

json_field_num() {
  local json="$1" field="$2"
  printf '%s' "$json" | sed -n "s/.*\"${field}\":\\([0-9.]*\\).*/\\1/p" | head -1
}

format_elapsed() {
  local seconds="$1"
  if [ -z "$seconds" ] || [ "$seconds" = "0" ]; then
    echo ""
    return
  fi
  printf "%.2fs" "$seconds"
}

run_go_packages() {
  local section="$1"
  local summary_key="$2"
  shift 2
  local -a go_args=("$@")

  section_header "$section"

  local passed=0 failed=0 exit_code=0
  local line action pkg test elapsed output

  output="$(cd "$ROOT/src" && go test "${go_args[@]}" -json ./... 2>&1)" || exit_code=$?

  while IFS= read -r line; do
    [ -z "$line" ] && continue

    action="$(json_field "$line" Action)"
    test="$(json_field "$line" Test)"
    pkg="$(json_field "$line" Package)"

    [ "$action" = "output" ] && continue
    [ -n "$test" ] && continue

    if [ "$action" = "skip" ]; then
      echo "$line" | grep -q "no test files" && continue
    fi

    if [ "$action" = "pass" ]; then
      elapsed="$(format_elapsed "$(json_field_num "$line" Elapsed)")"
      printf "  PASS  %-42s %s\n" "$pkg" "$elapsed"
      passed=$((passed + 1))
    elif [ "$action" = "fail" ]; then
      printf "  FAIL  %s\n" "$pkg"
      failed=$((failed + 1))
    fi
  done <<< "$output"

  echo ""
  echo "  ${passed} packages passed, ${failed} failed"

  if [ "$exit_code" -ne 0 ] || [ "$failed" -gt 0 ]; then
    SUMMARY[$summary_key]="FAIL"
    return 1
  fi
  SUMMARY[$summary_key]="PASS"
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

list_e2e_scenarios() {
  local dir="$ROOT/tests/e2e/scenarios"
  if [ ! -d "$dir" ]; then
    return
  fi
  local names=()
  local name
  for name in "$dir"/*; do
    [ -d "$name" ] || continue
    names+=("$(basename "$name")")
  done
  if [ "${#names[@]}" -gt 0 ]; then
    local IFS=', '
    echo "  Scenarios: ${names[*]}"
    echo ""
  fi
}

run_e2e() {
  section_header "E2E"
  list_e2e_scenarios
  require_docker

  local passed=0 failed=0 exit_code=0
  local line action test scenario elapsed output

  output="$(cd "$ROOT/tests/e2e" && go test -tags=integration -json -timeout 10m ./... 2>&1)" || exit_code=$?

  while IFS= read -r line; do
    [ -z "$line" ] && continue

    action="$(json_field "$line" Action)"
    test="$(json_field "$line" Test)"

    [ -z "$test" ] && continue
    echo "$test" | grep -q '/' && continue

    scenario="${E2E_SCENARIO_MAP[$test]:-unknown}"

    if [ "$action" = "pass" ]; then
      elapsed="$(format_elapsed "$(json_field_num "$line" Elapsed)")"
      printf "  PASS  %-32s (%-22s) %s\n" "$test" "$scenario" "$elapsed"
      passed=$((passed + 1))
    elif [ "$action" = "fail" ]; then
      printf "  FAIL  %-32s (%s)\n" "$test" "$scenario"
      failed=$((failed + 1))
    fi
  done <<< "$output"

  echo ""
  echo "  ${passed} scenarios passed, ${failed} failed"

  if [ "$exit_code" -ne 0 ] || [ "$failed" -gt 0 ]; then
    SUMMARY[e2e]="FAIL"
    return 1
  fi
  SUMMARY[e2e]="PASS"
}

print_summary() {
  section_header "SUMMARY"
  local key
  for key in unit integration e2e; do
    if [ -n "${SUMMARY[$key]:-}" ]; then
      printf "  %-14s %s\n" "${key}:" "${SUMMARY[$key]}"
    fi
  done
}

run_unit() {
  run_go_packages "UNIT" unit
}

run_integration() {
  run_go_packages "INTEGRATION" integration -tags=integration
}

run_all() {
  local failed=0
  run_unit || failed=1
  run_integration || failed=1
  run_e2e || failed=1
  print_summary
  return "$failed"
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
    run_all
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
