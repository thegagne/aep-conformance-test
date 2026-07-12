#!/usr/bin/env bash
#
# test-dotnet-server.sh — run the AEP conformance suite against thegagne/dotnet-aep-server
# and emit a markdown report.
#
# It clones (or reuses) and builds the dotnet server, launches it, runs the
# conformance tool with markdown output, and always tears the server down.
#
# Usage:
#   scripts/test-dotnet-server.sh [output-file]
#
# Environment overrides:
#   PORT       listen port for the server            (default: 5268)
#   REPO       git URL of the target server          (default: the dotnet-aep-server repo)
#   SRC_DIR    where to clone/build the server        (default: $TMPDIR/aep-conformance-dotnet-server, reused across runs)
#   OUTPUT     markdown output path                    (default: ./dotnet-conformance-report.md,
#              or the first positional argument; use "-" for stdout)
#
set -euo pipefail

PORT="${PORT:-5268}"
REPO="${REPO:-https://github.com/thegagne/dotnet-aep-server}"
# A fixed, reusable checkout dir (not mktemp): reused across runs, and a real
# path avoids the /var -> /private/var symlink that trips dotnet's restore.
SRC_DIR="${SRC_DIR:-${TMPDIR:-/tmp}/aep-conformance-dotnet-server}"
OUTPUT="${1:-${OUTPUT:-dotnet-conformance-report.md}}"
BASE_URL="http://localhost:${PORT}"

# Resolve the tool's module directory (this script lives in <repo>/scripts).
TOOL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

SERVER_PID=""
STARTED_SERVER=""
cleanup() {
  if [[ -n "${SERVER_PID}" ]] && kill -0 "${SERVER_PID}" 2>/dev/null; then
    kill "${SERVER_PID}" 2>/dev/null || true
    wait "${SERVER_PID}" 2>/dev/null || true
  fi
  # Fallback: only if we started a server, reap anything still on our port.
  if [[ -n "${STARTED_SERVER}" ]]; then
    local held
    held="$(lsof -ti "tcp:${PORT}" 2>/dev/null || true)"
    [[ -n "${held}" ]] && kill ${held} 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

command -v dotnet >/dev/null || { echo "error: dotnet not found on PATH" >&2; exit 1; }
command -v go     >/dev/null || { echo "error: go not found on PATH" >&2; exit 1; }

echo ">> Preparing dotnet-aep-server in ${SRC_DIR}" >&2
mkdir -p "${SRC_DIR}"
# Canonicalize (resolve symlinks) so dotnet doesn't restore the same project
# under two path spellings and collide on obj/*.nuget.g.props.
SRC_DIR="$(cd "${SRC_DIR}" && pwd -P)"
if [[ ! -d "${SRC_DIR}/.git" ]]; then
  git clone --depth 1 "${REPO}" "${SRC_DIR}" >&2
fi

echo ">> Restoring and building server" >&2
# Restore the server project graph, then build single-threaded (-m:1) with
# --no-restore. Single-threaded build avoids a parallel ordering race where a
# transitive project reference (Aep.Storage.Abstractions) isn't built in time
# (CS0012); the explicit restore avoids a concurrent-restore file collision.
dotnet restore "${SRC_DIR}/src/Aep.Server" >&2
dotnet build "${SRC_DIR}/src/Aep.Server" --no-restore -m:1 >&2

echo ">> Launching server on ${BASE_URL}" >&2
# Run the built DLL directly rather than `dotnet run`: the latter spawns the app
# as a child process, so killing the wrapper would orphan the real server. With
# `dotnet <dll>` the host IS the server process and a single kill stops it.
DLL="$(find "${SRC_DIR}/src/Aep.Server/bin" -name Aep.Server.dll -path '*/Debug/*' 2>/dev/null | head -1)"
[[ -n "${DLL}" ]] || { echo "error: built Aep.Server.dll not found" >&2; exit 1; }
STARTED_SERVER=1
# Run from the project directory so the server finds its resources.yaml, which it
# loads by a path relative to the working directory (as `dotnet run` would).
( cd "${SRC_DIR}/src/Aep.Server" && ASPNETCORE_URLS="${BASE_URL}" exec dotnet "${DLL}" ) \
  > "${SRC_DIR}/server.log" 2>&1 &
SERVER_PID=$!

echo ">> Waiting for ${BASE_URL}/openapi.json" >&2
for _ in $(seq 1 60); do
  if curl -fsS -o /dev/null "${BASE_URL}/openapi.json" 2>/dev/null; then
    ready=1; break
  fi
  if ! kill -0 "${SERVER_PID}" 2>/dev/null; then
    echo "error: server exited during startup; log:" >&2
    tail -n 20 "${SRC_DIR}/server.log" >&2
    exit 1
  fi
  sleep 1
done
[[ "${ready:-}" == "1" ]] || { echo "error: server did not become ready" >&2; exit 1; }

echo ">> Running conformance suite (markdown)" >&2
# `test` exits non-zero when the API is non-conformant; capture that without
# aborting the script so the report is still produced.
status=0
if [[ "${OUTPUT}" == "-" ]]; then
  ( cd "${TOOL_DIR}" && go run . test "${BASE_URL}" --output markdown ) || status=$?
else
  ( cd "${TOOL_DIR}" && go run . test "${BASE_URL}" --output markdown --output-file "${OUTPUT}" ) || status=$?
  echo ">> Wrote ${OUTPUT}" >&2
fi

if [[ "${status}" -eq 0 ]]; then
  echo ">> Result: CONFORMANT" >&2
else
  echo ">> Result: NON-CONFORMANT (exit ${status})" >&2
fi
exit "${status}"
