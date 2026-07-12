#!/usr/bin/env bash
#
# test-aepbase-server.sh — run the AEP conformance suite against
# rambleraptor/aepbase and emit a markdown report.
#
# Clones (or reuses) and builds aepbase, launches it against a fresh data
# directory, seeds a publisher -> book resource hierarchy through its meta-API
# (aepbase starts with zero resources, so without seeding there is nothing to
# exercise), runs the conformance tool with markdown output, and always tears
# the server down.
#
# Usage:
#   scripts/test-aepbase-server.sh [output-file]
#
# Environment overrides:
#   PORT      listen port for the server           (default: 8080)
#   REPO      git URL of the target server          (default: the aepbase repo)
#   SRC_DIR   where to clone/build the server        (default: $TMPDIR/aep-conformance-aepbase-server, reused across runs)
#   DATA_DIR  server data dir, wiped for a fresh run (default: $TMPDIR/aep-conformance-aepbase-data)
#   OUTPUT    markdown output path                    (default: ./aepbase-conformance-report.md,
#             or the first positional argument; use "-" for stdout)

source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

PORT="${PORT:-8080}"
REPO="${REPO:-https://github.com/rambleraptor/aepbase}"
SRC_DIR="${SRC_DIR:-${TMPDIR:-/tmp}/aep-conformance-aepbase-server}"
DATA_DIR="${DATA_DIR:-${TMPDIR:-/tmp}/aep-conformance-aepbase-data}"
OUTPUT="${1:-${OUTPUT:-aepbase-conformance-report.md}}"
BASE_URL="http://localhost:${PORT}"

trap cleanup EXIT INT TERM
require_cmd go git curl

log "preparing aepbase in ${SRC_DIR}"
mkdir -p "${SRC_DIR}"
SRC_DIR="$(cd "${SRC_DIR}" && pwd -P)"
if [[ ! -d "${SRC_DIR}/.git" ]]; then
  git clone --depth 1 "${REPO}" "${SRC_DIR}" >&2
fi

log "building server"
( cd "${SRC_DIR}" && go build -o aepbase ./ ) >&2
BIN="${SRC_DIR}/aepbase"

log "launching server on ${BASE_URL}"
# Start from an empty data dir so the run is reproducible: resource definitions
# are persisted to sqlite under --data-dir and reloaded on the next start, so a
# stale dir would carry resources forward between runs.
rm -rf "${DATA_DIR}"
mkdir -p "${DATA_DIR}"
SERVER_LOG="${DATA_DIR}/server.log"
STARTED_SERVER=1
"${BIN}" -port "${PORT}" -data-dir "${DATA_DIR}" > "${SERVER_LOG}" 2>&1 &
SERVER_PID=$!

wait_for_openapi "${BASE_URL}"

# seed_resource NAME JSON — POST a resource definition to the meta-API. aepbase
# needs no auth for this (its user/auth layer is a library-only, off-by-default
# option), so a plain content-type POST is enough.
seed_resource() {
  local name="$1" body="$2" code
  code="$(curl -sS -o /dev/null -w '%{http_code}' \
    -X POST "${BASE_URL}/aep-resource-definitions" \
    -H 'Content-Type: application/json' -d "${body}")"
  [[ "${code}" == "200" || "${code}" == "201" ]] || die "seeding ${name} failed (HTTP ${code})"
  log "seeded resource: ${name}"
}

log "seeding publisher -> book resource hierarchy"
seed_resource publisher \
  '{"singular":"publisher","plural":"publishers","user_settable_create":true,"schema":{"type":"object","properties":{"name":{"type":"string"},"location":{"type":"string"}}}}'
# `parents` nests book under publisher, yielding /publishers/{id}/books/{id} so
# the suite also exercises the parent-chain create/teardown path.
seed_resource book \
  '{"singular":"book","plural":"books","parents":["publisher"],"user_settable_create":true,"schema":{"type":"object","properties":{"title":{"type":"string"},"author":{"type":"string"},"published":{"type":"boolean"}}}}'

rc=0
run_conformance "${BASE_URL}" "${OUTPUT}" || rc=$?
exit "${rc}"
