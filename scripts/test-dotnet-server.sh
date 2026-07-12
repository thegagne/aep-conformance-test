#!/usr/bin/env bash
#
# test-dotnet-server.sh — run the AEP conformance suite against
# thegagne/dotnet-aep-server and emit a markdown report.
#
# Clones (or reuses) and builds the dotnet server, launches it, runs the
# conformance tool with markdown output, and always tears the server down.
#
# Usage:
#   scripts/test-dotnet-server.sh [output-file]
#
# Environment overrides:
#   PORT     listen port for the server           (default: 5268)
#   REPO     git URL of the target server          (default: the dotnet-aep-server repo)
#   SRC_DIR  where to clone/build the server        (default: $TMPDIR/aep-conformance-dotnet-server, reused across runs)
#   OUTPUT   markdown output path                    (default: ./dotnet-conformance-report.md,
#            or the first positional argument; use "-" for stdout)

source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

PORT="${PORT:-5268}"
REPO="${REPO:-https://github.com/thegagne/dotnet-aep-server}"
# A fixed, reusable checkout dir (not mktemp): reused across runs, and a real
# path avoids the /var -> /private/var symlink that trips dotnet's restore.
SRC_DIR="${SRC_DIR:-${TMPDIR:-/tmp}/aep-conformance-dotnet-server}"
OUTPUT="${1:-${OUTPUT:-dotnet-conformance-report.md}}"
BASE_URL="http://localhost:${PORT}"

trap cleanup EXIT INT TERM
require_cmd dotnet go git curl

log "preparing dotnet-aep-server in ${SRC_DIR}"
mkdir -p "${SRC_DIR}"
# Canonicalize (resolve symlinks) so dotnet doesn't restore the same project
# under two path spellings and collide on obj/*.nuget.g.props.
SRC_DIR="$(cd "${SRC_DIR}" && pwd -P)"
if [[ ! -d "${SRC_DIR}/.git" ]]; then
  git clone --depth 1 "${REPO}" "${SRC_DIR}" >&2
fi

log "restoring and building server"
# Restore the server project graph, then build single-threaded (-m:1) with
# --no-restore. Single-threaded build avoids a parallel ordering race where a
# transitive project reference (Aep.Storage.Abstractions) isn't built in time
# (CS0012); the explicit restore avoids a concurrent-restore file collision.
dotnet restore "${SRC_DIR}/src/Aep.Server" >&2
dotnet build "${SRC_DIR}/src/Aep.Server" --no-restore -m:1 >&2

log "launching server on ${BASE_URL}"
# Run the built DLL directly rather than `dotnet run`: the latter spawns the app
# as a child process, so killing the wrapper would orphan the real server. With
# `dotnet <dll>` the host IS the server process and a single kill stops it.
DLL="$(find "${SRC_DIR}/src/Aep.Server/bin" -name Aep.Server.dll -path '*/Debug/*' 2>/dev/null | head -1)"
[[ -n "${DLL}" ]] || die "built Aep.Server.dll not found"
SERVER_LOG="${SRC_DIR}/server.log"
STARTED_SERVER=1
# Run from the project directory so the server finds its resources.yaml, which it
# loads by a path relative to the working directory (as `dotnet run` would).
( cd "${SRC_DIR}/src/Aep.Server" && ASPNETCORE_URLS="${BASE_URL}" exec dotnet "${DLL}" ) \
  > "${SERVER_LOG}" 2>&1 &
SERVER_PID=$!

wait_for_openapi "${BASE_URL}"

rc=0
run_conformance "${BASE_URL}" "${OUTPUT}" || rc=$?
exit "${rc}"
