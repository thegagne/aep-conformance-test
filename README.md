# AEP Conformance Test

A behavioral conformance suite for [AEP.dev](https://aep.dev) APIs. Point it at a running server and it checks — against the **live API, not just the spec on paper** — whether the service actually behaves the way the AEPs require: correct status codes, resource lifecycles, pagination, strong consistency, field semantics, error shapes, and more.

It reads your API's own OpenAPI document, figures out the resources and which features they implement, then drives real requests and grades what comes back. **85 checks across ~30 AEPs**, each tagged **MUST / SHOULD / MAY**:

- a failed **MUST** breaks conformance and exits non-zero — so it drops straight into CI;
- a failed **SHOULD** is a warning (promote to failure with `--strict`);
- optional features you don't implement are reported **not applicable**, never held against you.

No configuration and no hand-written test cases — the spec is the test plan. REST/OpenAPI today, Protobuf later. Written in Go.

> Still `v0.x`: the set of checks changes between versions, so pin a version if you gate CI on the result. See [Versioning](#versioning).

## How it works

Three phases, driven entirely by your API's OpenAPI document:

1. **Discover** — loads the AEP-annotated OpenAPI 3.1 doc (from `<base-url>/openapi.json` or a file) and resolves the `x-aep-resource` annotations into a resource hierarchy, the standard and custom methods on each, and which optional features each opts into (pagination, filtering, etags, soft-delete, …).
2. **Exercise** — for every resource it runs a real lifecycle against the live server: build any parent chain, then `Create → Get → List → Update → Apply`, page through collections, probe the `-` wildcard read, and hit the negatives (missing → 404, duplicate → 409, bad input → 400). Consistency is verified through follow-up reads, every request/response is captured as evidence, and **everything it creates is torn down afterward**.
3. **Report** — each check is graded by requirement level and rendered to stdout, JSON, or markdown. Reports are stamped with the tool version, AEP spec revision, target, and timestamp, so a saved verdict is reproducible.

Static (spec-only) checks still run when you pass just a spec file; dynamic checks are skipped without a live server.

## Install

```bash
# build a local binary
go build -o aep-conformance .

# or install from source (installs a binary named `aep-conformance-test`)
go install github.com/thegagne/aep-conformance-test@latest
```

## Usage

```bash
# Test a live server (discovers <base-url>/openapi.json automatically)
aep-conformance test http://localhost:5268

# Supply the OpenAPI document as a file (still hits the live server for dynamic checks)
aep-conformance test http://localhost:5268 --openapi ./openapi.json

# Static-only run: pass just a spec file (dynamic checks are skipped)
aep-conformance test ./openapi.json

# Inspect what was discovered, without running any checks
aep-conformance discover http://localhost:5268
```

### Flags (`test`)

| flag | description |
|------|-------------|
| `--openapi <file\|url>` | OpenAPI spec source (default `<base-url>/openapi.json`) |
| `--base-url <url>` | live server base URL (defaults to the positional argument) |
| `--output <fmt>` | `stdout` (default), `json`, or `markdown` |
| `--output-file <path>` | write the report to a file instead of stdout |
| `--resource <name>` | limit the run to named resources (repeatable) |
| `-H, --header <hdr>` | extra request header, `Key: Value` or `Key=Value` (repeatable); applied to the spec fetch and every request |
| `--strict` | treat SHOULD failures as failures (non-zero exit) |
| `--verbose` | show passing checks and full request/response evidence |
| `--timeout <dur>` | per-request timeout (default `30s`) |

Use `-H` to test an authenticated API — the header goes on both the OpenAPI fetch and every live request:

```bash
aep-conformance test https://api.example.com -H "Authorization: Bearer $TOKEN"
aep-conformance test https://api.example.com -H "X-Api-Key=$KEY"
```

The process exits non-zero when any required (MUST) check fails, or any SHOULD check fails under `--strict`. Report checks are ordered by AEP number, failures first within each AEP.

## What it checks

Structural checks read the spec; behavioral checks exercise the live server. Anything targeting an optional feature reports *not applicable* when your API doesn't implement it.

- **Structure & naming** — OpenAPI 3.1 (101), API name (102), resource types (4); resource paths (122): the `path` field, parent-prefix hierarchy, collection/id segment alternation, path uniqueness, canonical returned paths, collection-id and id naming; HTTP verb mapping & transcoding (127); operationId naming (130); field naming & conventions (140–142); standard output-only fields (148).
- **Standard methods** — Get (131), List (132), Create (133), Update (134), Delete (135), Custom methods (136), Apply (137): status codes, response shapes, echoed input, path immutability, partial-merge (`PATCH`) and field-preservation (`Apply`) semantics, `merge-patch+json` support, strong consistency verified via follow-up reads, create-vs-update (201/200), and the negatives (404 / 409 / 400).
- **Design patterns (conditional)** — Pagination incl. an infinite-loop guard (158), Filtering (160), Reading across collections via the `-` wildcard (159), Read masks (157), field masks / `update_mask` (134), ETags & preconditions (154), Idempotency (155), Revisions (162), Soft delete (164), Expiration (214), Resource states (216), Unreachable / partial-failure (217).
- **Errors** — RFC 9457 Problem Details: `type`, `status`, content-type, and overall shape (193).

The full requirement-by-requirement coverage map — including what's deliberately out of scope (proto-only IDL rules, auth/permission behavior) — lives in [AEP-REQUIREMENTS-CHECKLIST.md](AEP-REQUIREMENTS-CHECKLIST.md).

## Versioning

This is still a work in progress, but once we hit 1.0 we may present stored conformance tests for server implementations in the repo. For now, see the [examples](examples/).

## References

- https://aep.dev — the specifications ([llms.txt](https://aep.dev/llms.txt) index)
- https://github.com/aep-dev/aepc — AEP compiler
- https://github.com/aep-dev/api-linter · [aep-openapi-linter](https://github.com/aep-dev/aep-openapi-linter) — static linters
- https://github.com/thegagne/dotnet-aep-server — a server this suite is tested against
