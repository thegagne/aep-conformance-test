# AEP Conformance Test

This tool runs a series of conformance tests against a live running API to validate its conformance with [AEP.dev](https://aep.dev) specifications.

It outputs a report (JSON, markdown, or stdout) detailing conformance with different parts of the specification.

Some areas are required and will be marked as non-conformant if violated; other areas may be not applicable — for example, if the API does not implement an optional feature, that is acceptable, but it will still be noted.

It can be used to test REST implementations (Protobuf is a future addition). Written in Go.

## How it works

1. **Discover** — the tool loads the API's AEP-annotated OpenAPI 3.1 document (fetched from `<base-url>/openapi.json` or supplied as a file) and parses the `x-aep-resource` annotations into a resource hierarchy, per-method endpoints, and a set of detected optional features (pagination, filtering, etags, soft-delete, …).
2. **Exercise** — for each resource, it runs the full lifecycle against the live server: create any parent chain, then `Create → Get → List → Update → Apply`, pagination, negative cases (missing → 404, duplicate → 409), and `Delete → Get (404)`. Everything created is torn down afterward.
3. **Report** — every check is classified by its requirement level. A failed **MUST** breaks conformance (non-zero exit); a failed **SHOULD** is a warning; **MAY** and absent optional features are noted as *not applicable*.

## Install / build

```bash
go build -o aep-conformance .
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
| `--strict` | treat SHOULD failures as failures (non-zero exit) |
| `--verbose` | show passing checks and full request/response evidence |
| `--timeout <dur>` | per-request timeout (default `30s`) |

The process exits non-zero when any required (MUST) check fails, or any SHOULD check fails under `--strict`.

## What it checks

Checks are grouped by AEP and gated on applicability — a check for an optional feature reports *not applicable* when the feature is absent from the spec.

- **Structure & naming** — OpenAPI 3.1 (101), API name (102), HTTP verb mapping (127), method/operationId naming (130), resource paths & the `path` field (122), resource types (4), field naming and conventions (140–145), standard fields (148), state enums (216).
- **Standard methods** — Get (131), List (132), Create (133), Update (134), Delete (135), Custom methods (136), Apply (137): status codes, response shapes, echoed input, path immutability, partial-merge semantics, strong-consistency, and the negative cases (404 / 409 / 400).
- **Design patterns (conditional)** — Pagination (158, incl. an infinite-loop guard), Filtering (160), Read masks (157), Field masks (161), Reading across collections (159), Unreachable (217), ETags/preconditions (154), Idempotency (155), Revisions (162), Soft delete (164), Expiration (214).
- **Errors** — RFC 9457 Problem Details shape and status/type/content-type (193).

## Architecture

```
cmd/                 cobra CLI (test, discover)
internal/discovery/  OpenAPI 3.1 -> APIModel (resources, endpoints, feature flags)
internal/client/     REST client that captures requests/responses as evidence
internal/sampler/    builds valid request bodies from JSON Schema
internal/checks/     data-driven check registry (static + dynamic)
internal/runner/     per-resource lifecycle orchestration + teardown
internal/report/     stdout / json / markdown renderers
```

Checks are small registered values (metadata + a closure). Static checks read the discovered model; dynamic checks read a *probe* of captured live interactions, so the network I/O happens once per resource and the assertions are pure.

## References

- https://aep.dev/llms.txt
- https://aep.dev
- https://github.com/aep-dev/aepcli
- https://github.com/aep-dev/aepc
- https://github.com/aep-dev/api-linter
- https://github.com/aep-dev/aep-openapi-linter
- https://github.com/rambleraptor/aepbase
- https://github.com/thegagne/dotnet-aep-server
