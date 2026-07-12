#!/usr/bin/env bash
#
# lib.sh — shared helpers for the conformance runner scripts
# (test-dotnet-server.sh, test-aepbase-server.sh).
#
# Source it; don't execute it. It enables strict mode, resolves the tool's
# module directory, and provides the pieces every runner needs: a server
# teardown trap, a readiness poll, and the conformance invocation. Each runner
# supplies only the server-specific build/launch (and, for aepbase, seed) steps.
set -euo pipefail

# log / die — status to stderr so stdout stays clean for report piping ("-").
log() { printf '>> %s\n' "$*" >&2; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }

# require_cmd CMD... — fail early if any required binary is missing from PATH.
require_cmd() {
  local c
  for c in "$@"; do
    command -v "$c" >/dev/null || die "$c not found on PATH"
  done
}

# Resolve the tool's module directory (this file lives in <repo>/scripts).
TOOL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# --- teardown ---------------------------------------------------------------
# Runners populate these before/after launching; cleanup() reads them. Register
# it once with:  trap cleanup EXIT INT TERM
SERVER_PID=""        # pid of the launched server, if any
STARTED_SERVER=""    # set to 1 only after WE start the server
PORT=""              # listen port, used for the fallback port reap
SERVER_LOG=""        # server stdout/stderr log, tailed on a failed startup

cleanup() {
  if [[ -n "${SERVER_PID}" ]] && kill -0 "${SERVER_PID}" 2>/dev/null; then
    kill "${SERVER_PID}" 2>/dev/null || true
    wait "${SERVER_PID}" 2>/dev/null || true
  fi
  # Fallback: only if WE started a server, reap anything still holding the port.
  # Guarding on STARTED_SERVER avoids killing a pre-existing server the user may
  # already be running on this port.
  if [[ -n "${STARTED_SERVER}" && -n "${PORT}" ]]; then
    local held
    held="$(lsof -ti "tcp:${PORT}" 2>/dev/null || true)"
    [[ -n "${held}" ]] && kill ${held} 2>/dev/null || true
  fi
}

# wait_for_openapi BASE_URL — poll until <base>/openapi.json responds, aborting
# early (with the server log) if the server dies during startup.
wait_for_openapi() {
  local base="$1" i
  log "waiting for ${base}/openapi.json"
  for i in $(seq 1 60); do
    if curl -fsS -o /dev/null "${base}/openapi.json" 2>/dev/null; then
      return 0
    fi
    if [[ -n "${SERVER_PID}" ]] && ! kill -0 "${SERVER_PID}" 2>/dev/null; then
      log "server exited during startup; log:"
      [[ -n "${SERVER_LOG}" && -f "${SERVER_LOG}" ]] && tail -n 20 "${SERVER_LOG}" >&2
      die "server did not start"
    fi
    sleep 1
  done
  die "server did not become ready at ${base}/openapi.json"
}

# run_conformance BASE_URL OUTPUT — run the suite with markdown output. OUTPUT
# of "-" writes to stdout; anything else is a file path. Returns the tool's exit
# code (non-zero == NON-CONFORMANT) without aborting, so the report still lands.
run_conformance() {
  local base="$1" out="$2" rc=0
  log "running conformance suite (markdown)"
  if [[ "${out}" == "-" ]]; then
    ( cd "${TOOL_DIR}" && go run . test "${base}" --output markdown ) || rc=$?
  else
    ( cd "${TOOL_DIR}" && go run . test "${base}" --output markdown --output-file "${out}" ) || rc=$?
    log "wrote ${out}"
  fi
  if [[ "${rc}" -eq 0 ]]; then
    log "result: CONFORMANT"
  else
    log "result: NON-CONFORMANT (exit ${rc})"
  fi
  return "${rc}"
}
