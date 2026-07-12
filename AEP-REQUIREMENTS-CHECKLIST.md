# AEP MUST / SHOULD Requirements — Coverage Checklist

Source: the AEPs indexed at <https://aep.dev/llms.txt> (standard methods + resource
design + the design patterns this tool references). Each normative requirement
(RFC-2119 MUST / MUST NOT / SHOULD / SHOULD NOT / MAY) is listed and marked:

- ✅ **Checked** — this tool has a registered check (check ID in parentheses).
- ❌ **Not checked** — in scope for a REST conformance tool, but no check exists.
- ⚪ **N/A** — proto-only (`google.api.*` annotations, message-naming) or
  auth/permission behavior this REST tool does not exercise.

76 checks are registered today. Coverage is strong on the standard-method
happy-path + negative cases, resource-path structure, custom methods (136), and
the design patterns; the main remaining gaps are proto-shaped naming and
anything auth-related (both largely out of scope for a REST tool).

---

## AEP-4 — Resource types
- ✅ Resource type is `{apiName}/{TypeName}` (`resource-type-format`)

## AEP-101 — OpenAPI
- ✅ Document uses OpenAPI 3.1.x (`openapi-version-3-1`)

## AEP-102 — API terminology
- ✅ API name is lowercase / starts with a letter / valid chars (`api-name-format`)

## AEP-122 — Resource paths
- ✅ Each resource has a `path` field (`path-field-present`)
- ✅ `path` field is `readOnly: true` (`path-field-readonly`)
- ✅ A resource path MUST be unique within an API (`resource-paths-unique`)
- ✅ Components MUST alternate collection-id / resource-id (`path-segments-alternate`)
- ❌ Segments MUST NOT contain `/`; MUST NOT need URL-escaping / be non-ASCII
- ✅ Parent path MUST be a prefix of the collection path (`parent-path-is-prefix`)
- ✅ All ID fields MUST be strings (`id-fields-are-strings`)
- ❌ All returned data MUST use the canonical resource path
- ❌ Resources MUST NOT expose tuples / self-links
- ✅ Collection identifiers MUST be lower-case kebab-case (`collection-id-format`)
- ✅ Collection identifiers MUST be plural (`collection-id-plural`, conservative)
- ❌ Resource IDs SHOULD be lower-case / conform to RFC-1034 / DNS-safe (SHOULD)
- ❌ User-settable IDs SHOULD NOT be a UUID (SHOULD NOT)

## AEP-127 — HTTP / gRPC transcoding
- ✅ GET/DELETE operations MUST NOT declare a request body (`no-body-<method>`)
- ✅ Get/Create/Delete MUST NOT require extra query fields (`no-extra-required-query-<method>`)
- ✅ Verb mapping per method (see per-method entries below)

## AEP-130 — Methods (naming)
- ✅ operationId format for standard non-list methods (`operation-id-standard`)
- ✅ List operationId is `List{Plural}` (`operation-id-list-plural`)

## AEP-131 — Get
- ✅ HTTP verb MUST be `GET` (`get-verb-get`)
- ✅ Response MUST be the resource itself (`get-200-on-existing`, `get-body-matches-created`)
- ✅ MUST NOT be a `body` / request body (`no-body-get`)
- ✅ MUST NOT require query-string fields (`no-extra-required-query-get`)
- ✅ Resource `path` field MUST be included (`path-field-present`)
- ✅ Permission-but-missing MUST reply 404 (`get-404-on-missing`)
- ✅ Request MUST be safe / no side effects (`get-safe-no-side-effects`)
- ✅ Response SHOULD be fully-populated (`get-fully-populated`)
- ⚪ RPC name MUST begin `Get` / request msg `…Request` (proto/schema naming)
- ❌ Path parameter MUST be `{resource-singular}`
- ⚪ No-permission SHOULD reply 404 / cannot-access SHOULD 403 (auth)
- ⚪ Permission MUST be checked before existence (auth)

## AEP-132 — List
- ✅ HTTP verb MUST be `GET` (`list-verb-get`)
- ✅ Empty collection with access MUST return `200 OK` (`list-200`)
- ✅ Resource array MUST be named `results` (`list-results-array-named-results`)
- ✅ `nextPageToken` MUST be present (`list-next-page-token-present`)
- ✅ `max_page_size`/`page_token` controls MUST be in query (`list-controls-in-query`)
- ✅ MUST NOT be a `body` / request body (`no-body-<list is GET>`)
- ⚪ Name MUST begin `List`; request/response `…Request`/`…Response` (proto/schema)
- ❌ Parent field MUST be present (unless top-level), MUST be named `parent`,
  MUST be the only URI variable, MUST be required, MUST accept the parent path
- ❌ Request MUST be safe / no side effects (list-specific; only Get is tested)
- ⚪ Permission MUST be checked before the collection exists (auth)
- ❌ `order_by` SHOULD exist / values comma-separated (SHOULD)
- ❌ `show_deleted` SHOULD exist for soft-delete (SHOULD)
- ❌ `total_size` MAY exist / reflect post-filter size (MAY/SHOULD)

## AEP-133 — Create
- ✅ HTTP verb MUST be `POST` (`create-verb-post`)
- ✅ Response MUST be the resource itself (`create-returns-resource`)
- ✅ Response MUST include provided fields (`create-echoes-input`)
- ✅ Response populates the `path` (`create-populates-path`)
- ✅ `path` on request body MUST be ignored (`create-path-field-ignored`)
- ✅ Duplicate MUST error `ALREADY_EXISTS`/409 (`create-duplicate-already-exists`)
- ✅ SHOULD return `201 Created` (`create-201-created`)
- ⚪ Name MUST begin `Create`; request msg `…Request` (proto/schema naming)
- ❌ Operation MUST have strong consistency (create→get not asserted as a check)
- ❌ Parent MUST be called `parent`; collection id MUST be a literal string
- ❌ Request MUST NOT contain other required fields
- ❌ `id` field MUST be a query parameter (when user-specified IDs supported)
- ⚪ Hidden-duplicate MUST error `PERMISSION_DENIED` (auth)

## AEP-134 — Update
- ✅ Request `path` MUST map to URI / be immutable (`update-path-unchanged`)
- ✅ Partial update MUST declare a field mask (`update-mask-declared`)
- ✅ MUST have strong consistency (`update-strong-consistency`, `update-consistent-on-get`)
- ✅ SHOULD support partial merge, RFC-7396 (`update-partial-merge`)
- ✅ HTTP verb SHOULD be `PATCH` (`update-verb-patch`)
- ✅ Missing resource MUST error `NOT_FOUND` (`update-404-on-missing`)
- ⚪ Name MUST begin `Update`; request schema `…Request` (proto/schema naming)
- ⚪ `path` MUST be the only URI variable (proto — REST expands ids into path)
- ❌ MUST support `application/merge-patch+json` MIME type (content-types not modeled)
- ❌ Response SHOULD be fully-populated (SHOULD)
- ❌ SHOULD NOT trigger side effects (SHOULD NOT)

## AEP-135 — Delete
- ✅ HTTP verb MUST be `DELETE` (`delete-verb-delete`)
- ✅ MUST NOT be a `body` key (`no-body-delete`)
- ✅ `path` field MUST be included (`path-field-present`)
- ✅ MUST NOT require other fields (`no-extra-required-query-delete`)
- ✅ Strong consistency: subsequent read MUST be 404 (`delete-strong-consistency-404`)
- ✅ Missing resource MUST 404 (`delete-404-on-missing`)
- ✅ Children + no `force` MUST fail 409/`FAILED_PRECONDITION` (`delete-cascade-requires-force`)
- ✅ SHOULD return `204 No Content`, empty body (`delete-204-empty`)
- ✅ A request body MUST be ignored, not error (`delete-body-ignored`)
- ⚪ No permission MUST error `403`/`PERMISSION_DENIED` (auth)

## AEP-136 — Custom methods
- ✅ URI MUST use `:` + custom verb (`custom-verb-in-uri`)
- ✅ URI verb MUST match the RPC/operationId verb (`custom-verb-matches-name`)
- ✅ Verb separators MUST be snake_case (`custom-verb-snake-case`)
- ✅ RPC name MUST NOT contain prepositions (`custom-name-no-prepositions`)
- ✅ `GET`/`DELETE` custom methods MUST omit the request body (`custom-no-body-on-get-delete`)
- ✅ SHOULD NOT use `PATCH`/`DELETE` (`custom-not-patch-delete`)
- ⚪ Path param MUST be `path` and the only URI variable (proto — REST expands ids)
- ❌ Custom methods SHOULD only cover what standard methods cannot (intent — not statically decidable)
- ❌ SHOULD NOT be used by declarative clients (intent)

## AEP-137 — Apply
- ✅ HTTP verb MUST be `PUT` (`apply-verb-put`)
- ✅ Response MUST be the resource itself (`apply-response-is-resource`)
- ✅ Apply-as-update MUST return `200` (`apply-update-200`)
- ✅ Apply-as-create MUST return `201` (`apply-create-201`)
- ✅ Absent request field MUST NOT be modified (`apply-preserves-absent-fields`)
- ✅ Missing required field MUST fail with `400` (`apply-missing-required-400`)
- ✅ MUST have strong consistency (`apply-consistent-on-get`)
- ✅ Response SHOULD be fully-populated (`apply-fully-populated`)
- ⚪ Name MUST begin `Apply`; request msg naming (proto/schema)
- ⚪ URI path MUST be the resource path; `path` MUST be the only variable (proto)

## AEP-140/141/142/148 — Field naming & standard fields
- ✅ Field names are lower_snake_case (`field-lower-snake-case`)
- ✅ No leading/trailing/adjacent underscores (`field-no-bad-underscores`)
- ✅ Words MUST NOT start with a digit (`field-word-not-start-digit`)
- ✅ Booleans SHOULD NOT carry `is_` prefix (`field-bool-no-is-prefix`)
- ✅ Prefer `uri` over `url` (`field-uri-not-url`)
- ✅ Avoid reserved keywords (`field-avoid-reserved-keywords`)
- ✅ Absolute-time fields SHOULD be `*_time` (`field-timestamp-suffix-time`)
- ✅ Time fields SHOULD NOT be past-tense (`field-timestamp-not-past-tense`)
- ✅ Counts use `_count` suffix, not `num_` (`field-count-suffix-not-num-prefix`)
- ✅ `create_time`/`update_time` are output-only (`standard-field-*-output-only`)

## Design patterns (conditional — only when the feature is present)
- ✅ **158 Pagination**: page size honored, next-token when more, terminates,
  negative size → 400 (`page-size-honored`, `page-next-token-when-more`,
  `page-terminates`, `page-size-negative-invalid`)
- ✅ **160 Filtering**: single field `filter`, invalid → 400
  (`filter-single-field-named-filter`, `filter-invalid-argument`)
- ✅ **154 ETags**: etag changes on update, stale If-Match → 412
  (`etag-changes-on-update`, `if-match-mismatch-412`)
- ✅ **157 Read masks**: view/read_mask optional (`read-mask-optional`)
- ✅ **161 Field masks**: update_mask declared (`update-mask-declared`)
- ✅ **162 Revisions**: `revisions` subcollection (`revision-collection-named-revisions`)
- ✅ **214 Expiration**: `expire_time` timestamp (`expire-time-timestamp`)
- ✅ **217 Unreachable**: partial-failure List declares `unreachable` (`unreachable-field-present`)
- ✅ **155 Idempotency**: idempotency key is a string, if present (`idempotency-key-string`, MAY)
- ✅ **164 Soft delete**: `show_deleted` is boolean when soft delete is offered
  (`soft-delete-show-deleted-bool`) — restore/expire semantics still unchecked
- ✅ **216 Resource states**: `state` is an enum and output-only
  (`state-field-is-enum`, `state-field-output-only`)
- ✅ **159 Reading across collections**: behavioral probe issues a `-` wildcard
  List; if honored, results MUST carry canonical parent ids, not `-`
  (`across-collections-real-parent-ids`)

## AEP-193 — Errors (RFC 9457 Problem Details)
- ✅ Error `type` present and a URI reference (`error-type-uri`, MUST)
- ✅ Error body follows Problem Details shape (`error-body-rfc9457-shape`, SHOULD)
- ✅ Error `status` matches HTTP code (`error-status-matches-http`, SHOULD)
- ✅ Errors served as JSON (`error-content-type-json`, SHOULD)

---

## Recently added

Round 1 — **custom methods (136)**: `custom-verb-in-uri`, `custom-verb-matches-name`,
`custom-verb-snake-case`, `custom-name-no-prepositions`, `custom-no-body-on-get-delete`,
`custom-not-patch-delete`; **resource paths (122)**: `collection-id-format`,
`collection-id-plural`, `id-fields-are-strings`.

Round 2 — **dynamic (134/135)**: `update-404-on-missing`, `delete-body-ignored`;
**path structure (122)**: `parent-path-is-prefix`, `path-segments-alternate`,
`resource-paths-unique`; **patterns (164/216)**: `soft-delete-show-deleted-bool`,
`state-field-is-enum`, `state-field-output-only`.

Round 3 — **reading across collections (159)**, behavioral: `across-collections-real-parent-ids`
issues a `-` wildcard List and verifies canonical parent ids in the response.

Round 4 — **Apply semantics (137)**: `apply-update-200`, `apply-create-201`,
`apply-consistent-on-get`, `apply-preserves-absent-fields`, `apply-missing-required-400`,
`apply-fully-populated`.

## Remaining gaps (in-scope but not checked)

1. **Update (134)** — `application/merge-patch+json` MIME (request content-types aren't
   modeled by discovery yet).
2. **Create (133)** — explicit strong-consistency assertion; `parent` naming / literal
   collection id; "no other required fields"; `id` as a query param.
3. **List extras (132)** — `order_by`, `show_deleted`, `total_size` (SHOULD/MAY).
4. **Resource-ID naming rules (122)** — DNS-safe, no-UUID, lower-case (SHOULD-level).
5. **Segment ASCII/URL-escape rules (122)** — literal-segment charset beyond the final
   collection id.

## Out of scope by design
- Proto-only requirements (`google.api.http`/`method_signature` annotations, message
  `…Request`/`…Response` naming) — this tool validates the REST/OpenAPI surface.
- Auth / permission behavior (403 vs 404 ordering, `PERMISSION_DENIED`) — the tool
  does not run authenticated negative cases.
