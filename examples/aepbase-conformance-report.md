# AEP Conformance Report

**API:** aepbase  
**Target:** http://localhost:8080  
**Tool:** aep-conformance dev  
**Spec:** aep.dev catalog @ 2026-07-11  
**Generated:** 2026-07-12T02:36:12Z  

❌ **NON-CONFORMANT** — 11 required check(s) failed

| pass | fail | warn | n/a | skip |
| ---: | ---: | ---: | --: | ---: |
|  183 |   11 |    5 |  79 |   28 |

**Legend**

- ✅ **PASS** — conforms.
- ❌ **FAIL** — a required (MUST) rule is violated; breaks conformance.
- ⚠️ **WARN** — a recommended (SHOULD) rule is violated; allowed, but worth fixing.
- ➖ **N/A** — targets an optional feature this API doesn't implement; not tested.
- ⏭️ **SKIP** — could not be evaluated (e.g. a prerequisite step failed).

## API

**Checks**

| status | AEP     | level | check                   | detail                                |
| :----- | :------ | :---- | :---------------------- | :------------------------------------ |
| PASS   | AEP-101 | MUST  | `openapi-version-3-1`   | openapi 3.1.0                         |
| PASS   | AEP-102 | MUST  | `api-name-format`       | api name: aepbase                     |
| PASS   | AEP-122 | MUST  | `resource-paths-unique` | all resource path patterns are unique |

## aep-resource-definitions

**Capabilities**

| capability       | kind    | implemented |
| :--------------- | :------ | :---------: |
| Get              | method  |      ✅      |
| List             | method  |      ✅      |
| Create           | method  |      ✅      |
| Update           | method  |      ✅      |
| Apply            | method  |      ❌      |
| Delete           | method  |      ✅      |
| pagination       | feature |      ✅      |
| filtering        | feature |      ❌      |
| ordering         | feature |      ❌      |
| skip             | feature |      ❌      |
| read_mask        | feature |      ❌      |
| field_mask       | feature |      ❌      |
| soft_delete      | feature |      ❌      |
| cascade_force    | feature |      ❌      |
| idempotency      | feature |      ❌      |
| revisions        | feature |      ❌      |
| expiration       | feature |      ❌      |
| unreachable      | feature |      ❌      |
| user_settable_id | feature |      ❌      |
| lro              | feature |      ❌      |

**Checks**

| status | AEP     | level      | check                                    | detail                                                                             |
| :----- | :------ | :--------- | :--------------------------------------- | :--------------------------------------------------------------------------------- |
| PASS   | AEP-4   | MUST       | `resource-type-format`                   | type: aepbase/aep-resource-definition                                              |
| PASS   | AEP-122 | MUST       | `collection-id-format`                   | collection id: aep-resource-definitions                                            |
| PASS   | AEP-122 | MUST       | `collection-id-plural`                   | collection id matches declared plural: aep-resource-definitions                    |
| PASS   | AEP-122 | MUST       | `id-fields-are-strings`                  | all ID path parameters are strings                                                 |
| PASS   | AEP-122 | MUST       | `path-field-present`                     | path field present                                                                 |
| PASS   | AEP-122 | MUST       | `path-field-readonly`                    | path is readOnly                                                                   |
| PASS   | AEP-122 | MUST       | `path-segments-alternate`                | segments alternate: aep-resource-definitions/{aep_resource_definition_id}          |
| PASS   | AEP-122 | MUST       | `segment-charset-ascii`                  | all literal segments are URL-safe                                                  |
| SKIP   | AEP-122 | MUST       | `path-matches-canonical-pattern`         | prerequisite 'get' did not run                                                     |
| N/A    | AEP-122 | MUST       | `parent-path-is-prefix`                  | feature not present in spec                                                        |
| N/A    | AEP-122 | SHOULD NOT | `user-settable-id-not-uuid`              | feature not present in spec                                                        |
| PASS   | AEP-127 | MUST       | `create-verb-post`                       | POST /aep-resource-definitions                                                     |
| PASS   | AEP-127 | MUST       | `delete-verb-delete`                     | DELETE /aep-resource-definitions/{aep_resource_definition_id}                      |
| PASS   | AEP-127 | MUST       | `get-verb-get`                           | GET /aep-resource-definitions/{aep_resource_definition_id}                         |
| PASS   | AEP-127 | MUST       | `list-verb-get`                          | GET /aep-resource-definitions                                                      |
| PASS   | AEP-127 | MUST       | `no-body-delete`                         | no request body                                                                    |
| PASS   | AEP-127 | MUST       | `no-body-get`                            | no request body                                                                    |
| PASS   | AEP-127 | MUST       | `no-body-list`                           | no request body                                                                    |
| PASS   | AEP-127 | SHOULD     | `update-verb-patch`                      | PATCH /aep-resource-definitions/{aep_resource_definition_id}                       |
| N/A    | AEP-127 | MUST       | `apply-verb-put`                         | feature not present in spec                                                        |
| FAIL   | AEP-130 | MUST       | `operation-id-list-plural`               | List operationId is "ListAepResourceDefinition", want "ListAepResourceDefinitions" |
| PASS   | AEP-130 | MUST       | `operation-id-standard`                  | standard operationIds well-formed                                                  |
| PASS   | AEP-131 | MUST       | `get-path-param-id-form`                 | id path parameter: {aep_resource_definition_id}                                    |
| PASS   | AEP-131 | MUST NOT   | `no-extra-required-query-delete`         | no required query fields                                                           |
| PASS   | AEP-131 | MUST NOT   | `no-extra-required-query-get`            | no required query fields                                                           |
| SKIP   | AEP-131 | MUST       | `get-200-on-existing`                    | prerequisite 'get' did not run                                                     |
| SKIP   | AEP-131 | MUST       | `get-404-on-missing`                     | prerequisite 'get_missing' did not run                                             |
| SKIP   | AEP-131 | MUST       | `get-body-matches-created`               | prerequisite 'get' did not run                                                     |
| SKIP   | AEP-131 | SHOULD     | `get-fully-populated`                    | prerequisite 'get' did not run                                                     |
| SKIP   | AEP-131 | MUST       | `get-safe-no-side-effects`               | prerequisite 'get' did not run                                                     |
| PASS   | AEP-132 | MUST       | `list-controls-in-query`                 | list controls are query parameters                                                 |
| PASS   | AEP-132 | MUST       | `list-next-page-token-present`           | next_page_token declared                                                           |
| SKIP   | AEP-132 | MUST       | `list-200`                               | prerequisite 'list' did not run                                                    |
| SKIP   | AEP-132 | MUST       | `list-results-array-named-results`       | prerequisite 'list' did not run                                                    |
| N/A    | AEP-132 | SHOULD     | `order-by-is-string`                     | feature not present in spec                                                        |
| N/A    | AEP-132 | MAY        | `total-size-is-integer`                  | feature not present in spec                                                        |
| FAIL   | AEP-133 | MUST       | `create-populates-path`                  | created resource has no populated 'path'                                           |
| FAIL   | AEP-133 | MUST       | `create-returns-resource`                | create returned 400                                                                |
| WARN   | AEP-133 | SHOULD     | `create-201-created`                     | create returned 400, expected 201                                                  |
| PASS   | AEP-133 | MUST       | `create-echoes-input`                    | all sent fields echoed                                                             |
| PASS   | AEP-133 | MUST NOT   | `create-no-required-query`               | no required query fields on Create                                                 |
| N/A    | AEP-133 | MUST       | `create-duplicate-already-exists`        | feature not present in spec                                                        |
| N/A    | AEP-133 | MUST       | `create-path-field-ignored`              | feature not present in spec                                                        |
| PASS   | AEP-134 | MUST       | `update-merge-patch-mime`                | declares application/merge-patch+json                                              |
| SKIP   | AEP-134 | MUST       | `update-404-on-missing`                  | prerequisite 'update_missing' did not run                                          |
| SKIP   | AEP-134 | MUST       | `update-consistent-on-get`               | prerequisite 'get_after_update' did not run                                        |
| SKIP   | AEP-134 | SHOULD     | `update-fully-populated`                 | prerequisite 'update' did not run                                                  |
| SKIP   | AEP-134 | SHOULD     | `update-partial-merge`                   | prerequisite 'get' did not run                                                     |
| SKIP   | AEP-134 | MUST       | `update-path-unchanged`                  | prerequisite 'update' did not run                                                  |
| SKIP   | AEP-134 | MUST       | `update-strong-consistency`              | prerequisite 'update' did not run                                                  |
| N/A    | AEP-134 | MUST       | `update-mask-declared`                   | feature not present in spec                                                        |
| SKIP   | AEP-135 | SHOULD     | `delete-204-empty`                       | prerequisite 'delete' did not run                                                  |
| SKIP   | AEP-135 | MUST       | `delete-404-on-missing`                  | prerequisite 'delete_missing' did not run                                          |
| SKIP   | AEP-135 | MUST       | `delete-body-ignored`                    | prerequisite 'delete_with_body' did not run                                        |
| SKIP   | AEP-135 | MUST       | `delete-strong-consistency-404`          | prerequisite 'get_after_delete' did not run                                        |
| N/A    | AEP-135 | MUST       | `delete-cascade-requires-force`          | feature not present in spec                                                        |
| N/A    | AEP-136 | MUST       | `custom-name-no-prepositions`            | feature not present in spec                                                        |
| N/A    | AEP-136 | MUST       | `custom-no-body-on-get-delete`           | feature not present in spec                                                        |
| N/A    | AEP-136 | SHOULD NOT | `custom-not-patch-delete`                | feature not present in spec                                                        |
| N/A    | AEP-136 | MUST       | `custom-verb-in-uri`                     | feature not present in spec                                                        |
| N/A    | AEP-136 | MUST       | `custom-verb-matches-name`               | feature not present in spec                                                        |
| N/A    | AEP-136 | MUST       | `custom-verb-snake-case`                 | feature not present in spec                                                        |
| N/A    | AEP-137 | MUST       | `apply-consistent-on-get`                | feature not present in spec                                                        |
| N/A    | AEP-137 | MUST       | `apply-create-201`                       | feature not present in spec                                                        |
| N/A    | AEP-137 | SHOULD     | `apply-fully-populated`                  | feature not present in spec                                                        |
| N/A    | AEP-137 | MUST       | `apply-missing-required-400`             | feature not present in spec                                                        |
| N/A    | AEP-137 | MUST       | `apply-preserves-absent-fields`          | feature not present in spec                                                        |
| N/A    | AEP-137 | MUST       | `apply-response-is-resource`             | feature not present in spec                                                        |
| N/A    | AEP-137 | MUST       | `apply-update-200`                       | feature not present in spec                                                        |
| PASS   | AEP-140 | SHOULD     | `field-avoid-reserved-keywords`          | all fields conform                                                                 |
| PASS   | AEP-140 | SHOULD     | `field-bool-no-is-prefix`                | all fields conform                                                                 |
| PASS   | AEP-140 | MUST       | `field-lower-snake-case`                 | all fields conform                                                                 |
| PASS   | AEP-140 | MUST       | `field-no-bad-underscores`               | all fields conform                                                                 |
| PASS   | AEP-140 | SHOULD     | `field-uri-not-url`                      | all fields conform                                                                 |
| PASS   | AEP-140 | MUST       | `field-word-not-start-digit`             | all fields conform                                                                 |
| PASS   | AEP-141 | SHOULD     | `field-count-suffix-not-num-prefix`      | all fields conform                                                                 |
| PASS   | AEP-142 | SHOULD     | `field-timestamp-not-past-tense`         | all fields conform                                                                 |
| PASS   | AEP-142 | SHOULD     | `field-timestamp-suffix-time`            | all fields conform                                                                 |
| PASS   | AEP-148 | SHOULD     | `standard-field-create_time-output-only` | create_time is readOnly                                                            |
| PASS   | AEP-148 | SHOULD     | `standard-field-update_time-output-only` | update_time is readOnly                                                            |
| N/A    | AEP-154 | MUST       | `etag-changes-on-update`                 | feature not present in spec                                                        |
| N/A    | AEP-154 | MUST       | `if-match-mismatch-412`                  | feature not present in spec                                                        |
| N/A    | AEP-155 | MAY        | `idempotency-key-string`                 | feature not present in spec                                                        |
| N/A    | AEP-157 | MUST       | `read-mask-optional`                     | feature not present in spec                                                        |
| SKIP   | AEP-158 | MUST       | `page-next-token-when-more`              | pagination not exercised                                                           |
| SKIP   | AEP-158 | MUST       | `page-size-honored`                      | pagination not exercised                                                           |
| SKIP   | AEP-158 | MUST       | `page-size-negative-invalid`             | pagination not exercised                                                           |
| SKIP   | AEP-158 | MUST       | `page-terminates`                        | pagination not exercised                                                           |
| N/A    | AEP-159 | MUST       | `across-collections-real-parent-ids`     | feature not present in spec                                                        |
| N/A    | AEP-160 | MUST       | `filter-invalid-argument`                | feature not present in spec                                                        |
| N/A    | AEP-160 | SHOULD     | `filter-single-field-named-filter`       | feature not present in spec                                                        |
| N/A    | AEP-162 | MUST       | `revision-collection-named-revisions`    | feature not present in spec                                                        |
| N/A    | AEP-164 | SHOULD     | `soft-delete-show-deleted-bool`          | feature not present in spec                                                        |
| SKIP   | AEP-193 | SHOULD     | `error-body-rfc9457-shape`               | no error response captured                                                         |
| SKIP   | AEP-193 | SHOULD     | `error-content-type-json`                | no error response captured                                                         |
| SKIP   | AEP-193 | SHOULD     | `error-status-matches-http`              | no JSON error response captured                                                    |
| SKIP   | AEP-193 | MUST       | `error-type-uri`                         | no JSON error response captured                                                    |
| N/A    | AEP-214 | MUST       | `expire-time-timestamp`                  | feature not present in spec                                                        |
| N/A    | AEP-216 | SHOULD     | `state-field-is-enum`                    | feature not present in spec                                                        |
| N/A    | AEP-216 | SHOULD     | `state-field-output-only`                | feature not present in spec                                                        |
| N/A    | AEP-217 | MUST       | `unreachable-field-present`              | feature not present in spec                                                        |

## publishers

**Capabilities**

| capability       | kind    | implemented |
| :--------------- | :------ | :---------: |
| Get              | method  |      ✅      |
| List             | method  |      ✅      |
| Create           | method  |      ✅      |
| Update           | method  |      ✅      |
| Apply            | method  |      ✅      |
| Delete           | method  |      ✅      |
| pagination       | feature |      ✅      |
| filtering        | feature |      ✅      |
| ordering         | feature |      ❌      |
| skip             | feature |      ✅      |
| read_mask        | feature |      ❌      |
| field_mask       | feature |      ❌      |
| soft_delete      | feature |      ❌      |
| cascade_force    | feature |      ✅      |
| idempotency      | feature |      ❌      |
| revisions        | feature |      ❌      |
| expiration       | feature |      ❌      |
| unreachable      | feature |      ❌      |
| user_settable_id | feature |      ✅      |
| lro              | feature |      ❌      |

**Checks**

| status | AEP     | level      | check                                    | detail                                                                  |
| :----- | :------ | :--------- | :--------------------------------------- | :---------------------------------------------------------------------- |
| PASS   | AEP-4   | MUST       | `resource-type-format`                   | type: aepbase/publisher                                                 |
| PASS   | AEP-122 | MUST       | `collection-id-format`                   | collection id: publishers                                               |
| PASS   | AEP-122 | MUST       | `collection-id-plural`                   | collection id matches declared plural: publishers                       |
| PASS   | AEP-122 | MUST       | `id-fields-are-strings`                  | all ID path parameters are strings                                      |
| PASS   | AEP-122 | MUST       | `path-field-present`                     | path field present                                                      |
| PASS   | AEP-122 | MUST       | `path-field-readonly`                    | path is readOnly                                                        |
| PASS   | AEP-122 | MUST       | `path-matches-canonical-pattern`         | returned path matches canonical pattern: publishers/aepconf-publisher-2 |
| PASS   | AEP-122 | MUST       | `path-segments-alternate`                | segments alternate: publishers/{publisher_id}                           |
| PASS   | AEP-122 | MUST       | `segment-charset-ascii`                  | all literal segments are URL-safe                                       |
| PASS   | AEP-122 | SHOULD NOT | `user-settable-id-not-uuid`              | user-settable id is not a UUID                                          |
| N/A    | AEP-122 | MUST       | `parent-path-is-prefix`                  | feature not present in spec                                             |
| PASS   | AEP-127 | MUST       | `apply-verb-put`                         | PUT /publishers/{publisher_id}                                          |
| PASS   | AEP-127 | MUST       | `create-verb-post`                       | POST /publishers                                                        |
| PASS   | AEP-127 | MUST       | `delete-verb-delete`                     | DELETE /publishers/{publisher_id}                                       |
| PASS   | AEP-127 | MUST       | `get-verb-get`                           | GET /publishers/{publisher_id}                                          |
| PASS   | AEP-127 | MUST       | `list-verb-get`                          | GET /publishers                                                         |
| PASS   | AEP-127 | MUST       | `no-body-delete`                         | no request body                                                         |
| PASS   | AEP-127 | MUST       | `no-body-get`                            | no request body                                                         |
| PASS   | AEP-127 | MUST       | `no-body-list`                           | no request body                                                         |
| PASS   | AEP-127 | SHOULD     | `update-verb-patch`                      | PATCH /publishers/{publisher_id}                                        |
| FAIL   | AEP-130 | MUST       | `operation-id-list-plural`               | List operationId is "ListPublisher", want "ListPublishers"              |
| PASS   | AEP-130 | MUST       | `operation-id-standard`                  | standard operationIds well-formed                                       |
| PASS   | AEP-131 | MUST       | `get-200-on-existing`                    | 200 OK                                                                  |
| PASS   | AEP-131 | MUST       | `get-404-on-missing`                     | 404 on missing                                                          |
| PASS   | AEP-131 | MUST       | `get-body-matches-created`               | path matches created resource                                           |
| PASS   | AEP-131 | SHOULD     | `get-fully-populated`                    | resource fully populated                                                |
| PASS   | AEP-131 | MUST       | `get-path-param-id-form`                 | id path parameter: {publisher_id}                                       |
| PASS   | AEP-131 | MUST       | `get-safe-no-side-effects`               | repeated GET is stable                                                  |
| PASS   | AEP-131 | MUST NOT   | `no-extra-required-query-delete`         | no required query fields                                                |
| PASS   | AEP-131 | MUST NOT   | `no-extra-required-query-get`            | no required query fields                                                |
| PASS   | AEP-132 | MUST       | `list-200`                               | 200 OK                                                                  |
| PASS   | AEP-132 | MUST       | `list-controls-in-query`                 | list controls are query parameters                                      |
| PASS   | AEP-132 | MUST       | `list-next-page-token-present`           | next_page_token declared                                                |
| PASS   | AEP-132 | MUST       | `list-results-array-named-results`       | results array present                                                   |
| N/A    | AEP-132 | SHOULD     | `order-by-is-string`                     | feature not present in spec                                             |
| N/A    | AEP-132 | MAY        | `total-size-is-integer`                  | feature not present in spec                                             |
| WARN   | AEP-133 | SHOULD     | `create-201-created`                     | create returned 200, expected 201                                       |
| PASS   | AEP-133 | MUST       | `create-duplicate-already-exists`        | 409 on duplicate id                                                     |
| PASS   | AEP-133 | MUST       | `create-echoes-input`                    | all sent fields echoed                                                  |
| PASS   | AEP-133 | MUST NOT   | `create-no-required-query`               | no required query fields on Create                                      |
| PASS   | AEP-133 | MUST       | `create-path-field-ignored`              | client path ignored; server path = publishers/aepconf-publisher-4       |
| PASS   | AEP-133 | MUST       | `create-populates-path`                  | path = publishers/aepconf-publisher-2                                   |
| PASS   | AEP-133 | MUST       | `create-returns-resource`                | create returned resource                                                |
| PASS   | AEP-134 | MUST       | `update-404-on-missing`                  | 404 on missing update                                                   |
| PASS   | AEP-134 | MUST       | `update-consistent-on-get`               | update observed on subsequent Get                                       |
| PASS   | AEP-134 | SHOULD     | `update-fully-populated`                 | update response fully populated                                         |
| PASS   | AEP-134 | MUST       | `update-merge-patch-mime`                | declares application/merge-patch+json                                   |
| PASS   | AEP-134 | SHOULD     | `update-partial-merge`                   | unspecified fields preserved across update                              |
| PASS   | AEP-134 | MUST       | `update-path-unchanged`                  | path unchanged after update                                             |
| PASS   | AEP-134 | MUST       | `update-strong-consistency`              | update reflected in response                                            |
| N/A    | AEP-134 | MUST       | `update-mask-declared`                   | feature not present in spec                                             |
| PASS   | AEP-135 | SHOULD     | `delete-204-empty`                       | 204 No Content                                                          |
| PASS   | AEP-135 | MUST       | `delete-404-on-missing`                  | 404 on missing delete                                                   |
| PASS   | AEP-135 | MUST       | `delete-body-ignored`                    | delete-with-body succeeded (204)                                        |
| PASS   | AEP-135 | MUST       | `delete-cascade-requires-force`          | force flag present                                                      |
| PASS   | AEP-135 | MUST       | `delete-strong-consistency-404`          | 404 after delete                                                        |
| N/A    | AEP-136 | MUST       | `custom-name-no-prepositions`            | feature not present in spec                                             |
| N/A    | AEP-136 | MUST       | `custom-no-body-on-get-delete`           | feature not present in spec                                             |
| N/A    | AEP-136 | SHOULD NOT | `custom-not-patch-delete`                | feature not present in spec                                             |
| N/A    | AEP-136 | MUST       | `custom-verb-in-uri`                     | feature not present in spec                                             |
| N/A    | AEP-136 | MUST       | `custom-verb-matches-name`               | feature not present in spec                                             |
| N/A    | AEP-136 | MUST       | `custom-verb-snake-case`                 | feature not present in spec                                             |
| FAIL   | AEP-137 | MUST       | `apply-create-201`                       | apply created a resource but returned 200, expected 201                 |
| PASS   | AEP-137 | MUST       | `apply-consistent-on-get`                | apply observed on subsequent Get                                        |
| PASS   | AEP-137 | SHOULD     | `apply-fully-populated`                  | apply response fully populated                                          |
| PASS   | AEP-137 | MUST       | `apply-preserves-absent-fields`          | omitted field "location" preserved across apply                         |
| PASS   | AEP-137 | MUST       | `apply-response-is-resource`             | apply returned resource                                                 |
| PASS   | AEP-137 | MUST       | `apply-update-200`                       | 200 OK on apply-update                                                  |
| N/A    | AEP-137 | MUST       | `apply-missing-required-400`             | feature not present in spec                                             |
| PASS   | AEP-140 | SHOULD     | `field-avoid-reserved-keywords`          | all fields conform                                                      |
| PASS   | AEP-140 | SHOULD     | `field-bool-no-is-prefix`                | all fields conform                                                      |
| PASS   | AEP-140 | MUST       | `field-lower-snake-case`                 | all fields conform                                                      |
| PASS   | AEP-140 | MUST       | `field-no-bad-underscores`               | all fields conform                                                      |
| PASS   | AEP-140 | SHOULD     | `field-uri-not-url`                      | all fields conform                                                      |
| PASS   | AEP-140 | MUST       | `field-word-not-start-digit`             | all fields conform                                                      |
| PASS   | AEP-141 | SHOULD     | `field-count-suffix-not-num-prefix`      | all fields conform                                                      |
| PASS   | AEP-142 | SHOULD     | `field-timestamp-not-past-tense`         | all fields conform                                                      |
| PASS   | AEP-142 | SHOULD     | `field-timestamp-suffix-time`            | all fields conform                                                      |
| PASS   | AEP-148 | SHOULD     | `standard-field-create_time-output-only` | create_time is readOnly                                                 |
| PASS   | AEP-148 | SHOULD     | `standard-field-update_time-output-only` | update_time is readOnly                                                 |
| N/A    | AEP-154 | MUST       | `etag-changes-on-update`                 | feature not present in spec                                             |
| N/A    | AEP-154 | MUST       | `if-match-mismatch-412`                  | feature not present in spec                                             |
| N/A    | AEP-155 | MAY        | `idempotency-key-string`                 | feature not present in spec                                             |
| N/A    | AEP-157 | MUST       | `read-mask-optional`                     | feature not present in spec                                             |
| FAIL   | AEP-158 | MUST       | `page-size-negative-invalid`             | negative page size returned 200, expected 400                           |
| PASS   | AEP-158 | MUST       | `page-next-token-when-more`              | next_page_token present when more pages exist                           |
| PASS   | AEP-158 | MUST       | `page-size-honored`                      | first page returned ≤ requested size                                    |
| PASS   | AEP-158 | MUST       | `page-terminates`                        | pagination reached end in 3 page(s)                                     |
| N/A    | AEP-159 | MUST       | `across-collections-real-parent-ids`     | feature not present in spec                                             |
| PASS   | AEP-160 | MUST       | `filter-invalid-argument`                | 400 on invalid filter                                                   |
| PASS   | AEP-160 | SHOULD     | `filter-single-field-named-filter`       | filter field present                                                    |
| N/A    | AEP-162 | MUST       | `revision-collection-named-revisions`    | feature not present in spec                                             |
| N/A    | AEP-164 | SHOULD     | `soft-delete-show-deleted-bool`          | feature not present in spec                                             |
| FAIL   | AEP-193 | MUST       | `error-type-uri`                         | error response has no 'type'                                            |
| WARN   | AEP-193 | SHOULD     | `error-body-rfc9457-shape`               | error body missing field(s): type, status, title                        |
| PASS   | AEP-193 | SHOULD     | `error-content-type-json`                | content-type: application/json                                          |
| SKIP   | AEP-193 | SHOULD     | `error-status-matches-http`              | no status field in error body                                           |
| N/A    | AEP-214 | MUST       | `expire-time-timestamp`                  | feature not present in spec                                             |
| N/A    | AEP-216 | SHOULD     | `state-field-is-enum`                    | feature not present in spec                                             |
| N/A    | AEP-216 | SHOULD     | `state-field-output-only`                | feature not present in spec                                             |
| N/A    | AEP-217 | MUST       | `unreachable-field-present`              | feature not present in spec                                             |

## books

**Capabilities**

| capability       | kind    | implemented |
| :--------------- | :------ | :---------: |
| Get              | method  |      ✅      |
| List             | method  |      ✅      |
| Create           | method  |      ✅      |
| Update           | method  |      ✅      |
| Apply            | method  |      ✅      |
| Delete           | method  |      ✅      |
| pagination       | feature |      ✅      |
| filtering        | feature |      ✅      |
| ordering         | feature |      ❌      |
| skip             | feature |      ✅      |
| read_mask        | feature |      ❌      |
| field_mask       | feature |      ❌      |
| soft_delete      | feature |      ❌      |
| cascade_force    | feature |      ❌      |
| idempotency      | feature |      ❌      |
| revisions        | feature |      ❌      |
| expiration       | feature |      ❌      |
| unreachable      | feature |      ❌      |
| user_settable_id | feature |      ✅      |
| lro              | feature |      ❌      |

**Checks**

| status | AEP     | level      | check                                    | detail                                                                                        |
| :----- | :------ | :--------- | :--------------------------------------- | :-------------------------------------------------------------------------------------------- |
| PASS   | AEP-4   | MUST       | `resource-type-format`                   | type: aepbase/book                                                                            |
| PASS   | AEP-122 | MUST       | `collection-id-format`                   | collection id: books                                                                          |
| PASS   | AEP-122 | MUST       | `collection-id-plural`                   | collection id matches declared plural: books                                                  |
| PASS   | AEP-122 | MUST       | `id-fields-are-strings`                  | all ID path parameters are strings                                                            |
| PASS   | AEP-122 | MUST       | `parent-path-is-prefix`                  | publishers/{publisher_id} ⊑ publishers/{publisher_id}/books/{book_id}                         |
| PASS   | AEP-122 | MUST       | `path-field-present`                     | path field present                                                                            |
| PASS   | AEP-122 | MUST       | `path-field-readonly`                    | path is readOnly                                                                              |
| PASS   | AEP-122 | MUST       | `path-matches-canonical-pattern`         | returned path matches canonical pattern: publishers/aepconf-publisher-9/books/aepconf-book-10 |
| PASS   | AEP-122 | MUST       | `path-segments-alternate`                | segments alternate: publishers/{publisher_id}/books/{book_id}                                 |
| PASS   | AEP-122 | MUST       | `segment-charset-ascii`                  | all literal segments are URL-safe                                                             |
| PASS   | AEP-122 | SHOULD NOT | `user-settable-id-not-uuid`              | user-settable id is not a UUID                                                                |
| PASS   | AEP-127 | MUST       | `apply-verb-put`                         | PUT /publishers/{publisher_id}/books/{book_id}                                                |
| PASS   | AEP-127 | MUST       | `create-verb-post`                       | POST /publishers/{publisher_id}/books                                                         |
| PASS   | AEP-127 | MUST       | `delete-verb-delete`                     | DELETE /publishers/{publisher_id}/books/{book_id}                                             |
| PASS   | AEP-127 | MUST       | `get-verb-get`                           | GET /publishers/{publisher_id}/books/{book_id}                                                |
| PASS   | AEP-127 | MUST       | `list-verb-get`                          | GET /publishers/{publisher_id}/books                                                          |
| PASS   | AEP-127 | MUST       | `no-body-delete`                         | no request body                                                                               |
| PASS   | AEP-127 | MUST       | `no-body-get`                            | no request body                                                                               |
| PASS   | AEP-127 | MUST       | `no-body-list`                           | no request body                                                                               |
| PASS   | AEP-127 | SHOULD     | `update-verb-patch`                      | PATCH /publishers/{publisher_id}/books/{book_id}                                              |
| FAIL   | AEP-130 | MUST       | `operation-id-list-plural`               | List operationId is "ListBook", want "ListBooks"                                              |
| PASS   | AEP-130 | MUST       | `operation-id-standard`                  | standard operationIds well-formed                                                             |
| PASS   | AEP-131 | MUST       | `get-200-on-existing`                    | 200 OK                                                                                        |
| PASS   | AEP-131 | MUST       | `get-404-on-missing`                     | 404 on missing                                                                                |
| PASS   | AEP-131 | MUST       | `get-body-matches-created`               | path matches created resource                                                                 |
| PASS   | AEP-131 | SHOULD     | `get-fully-populated`                    | resource fully populated                                                                      |
| PASS   | AEP-131 | MUST       | `get-path-param-id-form`                 | id path parameter: {book_id}                                                                  |
| PASS   | AEP-131 | MUST       | `get-safe-no-side-effects`               | repeated GET is stable                                                                        |
| PASS   | AEP-131 | MUST NOT   | `no-extra-required-query-delete`         | no required query fields                                                                      |
| PASS   | AEP-131 | MUST NOT   | `no-extra-required-query-get`            | no required query fields                                                                      |
| PASS   | AEP-132 | MUST       | `list-200`                               | 200 OK                                                                                        |
| PASS   | AEP-132 | MUST       | `list-controls-in-query`                 | list controls are query parameters                                                            |
| PASS   | AEP-132 | MUST       | `list-next-page-token-present`           | next_page_token declared                                                                      |
| PASS   | AEP-132 | MUST       | `list-results-array-named-results`       | results array present                                                                         |
| N/A    | AEP-132 | SHOULD     | `order-by-is-string`                     | feature not present in spec                                                                   |
| N/A    | AEP-132 | MAY        | `total-size-is-integer`                  | feature not present in spec                                                                   |
| WARN   | AEP-133 | SHOULD     | `create-201-created`                     | create returned 200, expected 201                                                             |
| PASS   | AEP-133 | MUST       | `create-duplicate-already-exists`        | 409 on duplicate id                                                                           |
| PASS   | AEP-133 | MUST       | `create-echoes-input`                    | all sent fields echoed                                                                        |
| PASS   | AEP-133 | MUST NOT   | `create-no-required-query`               | no required query fields on Create                                                            |
| PASS   | AEP-133 | MUST       | `create-path-field-ignored`              | client path ignored; server path = publishers/aepconf-publisher-9/books/aepconf-book-12       |
| PASS   | AEP-133 | MUST       | `create-populates-path`                  | path = publishers/aepconf-publisher-9/books/aepconf-book-10                                   |
| PASS   | AEP-133 | MUST       | `create-returns-resource`                | create returned resource                                                                      |
| PASS   | AEP-134 | MUST       | `update-404-on-missing`                  | 404 on missing update                                                                         |
| PASS   | AEP-134 | MUST       | `update-consistent-on-get`               | update observed on subsequent Get                                                             |
| PASS   | AEP-134 | SHOULD     | `update-fully-populated`                 | update response fully populated                                                               |
| PASS   | AEP-134 | MUST       | `update-merge-patch-mime`                | declares application/merge-patch+json                                                         |
| PASS   | AEP-134 | SHOULD     | `update-partial-merge`                   | unspecified fields preserved across update                                                    |
| PASS   | AEP-134 | MUST       | `update-path-unchanged`                  | path unchanged after update                                                                   |
| PASS   | AEP-134 | MUST       | `update-strong-consistency`              | update reflected in response                                                                  |
| N/A    | AEP-134 | MUST       | `update-mask-declared`                   | feature not present in spec                                                                   |
| PASS   | AEP-135 | SHOULD     | `delete-204-empty`                       | 204 No Content                                                                                |
| PASS   | AEP-135 | MUST       | `delete-404-on-missing`                  | 404 on missing delete                                                                         |
| PASS   | AEP-135 | MUST       | `delete-body-ignored`                    | delete-with-body succeeded (204)                                                              |
| PASS   | AEP-135 | MUST       | `delete-strong-consistency-404`          | 404 after delete                                                                              |
| N/A    | AEP-135 | MUST       | `delete-cascade-requires-force`          | feature not present in spec                                                                   |
| N/A    | AEP-136 | MUST       | `custom-name-no-prepositions`            | feature not present in spec                                                                   |
| N/A    | AEP-136 | MUST       | `custom-no-body-on-get-delete`           | feature not present in spec                                                                   |
| N/A    | AEP-136 | SHOULD NOT | `custom-not-patch-delete`                | feature not present in spec                                                                   |
| N/A    | AEP-136 | MUST       | `custom-verb-in-uri`                     | feature not present in spec                                                                   |
| N/A    | AEP-136 | MUST       | `custom-verb-matches-name`               | feature not present in spec                                                                   |
| N/A    | AEP-136 | MUST       | `custom-verb-snake-case`                 | feature not present in spec                                                                   |
| FAIL   | AEP-137 | MUST       | `apply-create-201`                       | apply created a resource but returned 200, expected 201                                       |
| PASS   | AEP-137 | MUST       | `apply-consistent-on-get`                | apply observed on subsequent Get                                                              |
| PASS   | AEP-137 | SHOULD     | `apply-fully-populated`                  | apply response fully populated                                                                |
| PASS   | AEP-137 | MUST       | `apply-preserves-absent-fields`          | omitted field "author" preserved across apply                                                 |
| PASS   | AEP-137 | MUST       | `apply-response-is-resource`             | apply returned resource                                                                       |
| PASS   | AEP-137 | MUST       | `apply-update-200`                       | 200 OK on apply-update                                                                        |
| N/A    | AEP-137 | MUST       | `apply-missing-required-400`             | feature not present in spec                                                                   |
| PASS   | AEP-140 | SHOULD     | `field-avoid-reserved-keywords`          | all fields conform                                                                            |
| PASS   | AEP-140 | SHOULD     | `field-bool-no-is-prefix`                | all fields conform                                                                            |
| PASS   | AEP-140 | MUST       | `field-lower-snake-case`                 | all fields conform                                                                            |
| PASS   | AEP-140 | MUST       | `field-no-bad-underscores`               | all fields conform                                                                            |
| PASS   | AEP-140 | SHOULD     | `field-uri-not-url`                      | all fields conform                                                                            |
| PASS   | AEP-140 | MUST       | `field-word-not-start-digit`             | all fields conform                                                                            |
| PASS   | AEP-141 | SHOULD     | `field-count-suffix-not-num-prefix`      | all fields conform                                                                            |
| PASS   | AEP-142 | SHOULD     | `field-timestamp-not-past-tense`         | all fields conform                                                                            |
| PASS   | AEP-142 | SHOULD     | `field-timestamp-suffix-time`            | all fields conform                                                                            |
| PASS   | AEP-148 | SHOULD     | `standard-field-create_time-output-only` | create_time is readOnly                                                                       |
| PASS   | AEP-148 | SHOULD     | `standard-field-update_time-output-only` | update_time is readOnly                                                                       |
| N/A    | AEP-154 | MUST       | `etag-changes-on-update`                 | feature not present in spec                                                                   |
| N/A    | AEP-154 | MUST       | `if-match-mismatch-412`                  | feature not present in spec                                                                   |
| N/A    | AEP-155 | MAY        | `idempotency-key-string`                 | feature not present in spec                                                                   |
| N/A    | AEP-157 | MUST       | `read-mask-optional`                     | feature not present in spec                                                                   |
| FAIL   | AEP-158 | MUST       | `page-size-negative-invalid`             | negative page size returned 200, expected 400                                                 |
| PASS   | AEP-158 | MUST       | `page-next-token-when-more`              | next_page_token present when more pages exist                                                 |
| PASS   | AEP-158 | MUST       | `page-size-honored`                      | first page returned ≤ requested size                                                          |
| PASS   | AEP-158 | MUST       | `page-terminates`                        | pagination reached end in 3 page(s)                                                           |
| N/A    | AEP-159 | MUST       | `across-collections-real-parent-ids`     | wildcard list returned no results to verify                                                   |
| PASS   | AEP-160 | MUST       | `filter-invalid-argument`                | 400 on invalid filter                                                                         |
| PASS   | AEP-160 | SHOULD     | `filter-single-field-named-filter`       | filter field present                                                                          |
| N/A    | AEP-162 | MUST       | `revision-collection-named-revisions`    | feature not present in spec                                                                   |
| N/A    | AEP-164 | SHOULD     | `soft-delete-show-deleted-bool`          | feature not present in spec                                                                   |
| FAIL   | AEP-193 | MUST       | `error-type-uri`                         | error response has no 'type'                                                                  |
| WARN   | AEP-193 | SHOULD     | `error-body-rfc9457-shape`               | error body missing field(s): type, status, title                                              |
| PASS   | AEP-193 | SHOULD     | `error-content-type-json`                | content-type: application/json                                                                |
| SKIP   | AEP-193 | SHOULD     | `error-status-matches-http`              | no status field in error body                                                                 |
| N/A    | AEP-214 | MUST       | `expire-time-timestamp`                  | feature not present in spec                                                                   |
| N/A    | AEP-216 | SHOULD     | `state-field-is-enum`                    | feature not present in spec                                                                   |
| N/A    | AEP-216 | SHOULD     | `state-field-output-only`                | feature not present in spec                                                                   |
| N/A    | AEP-217 | MUST       | `unreachable-field-present`              | feature not present in spec                                                                   |

