package checks

import (
	"strings"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

func init() {
	// AEP-160: filtering uses a single field named 'filter'.
	Register(Check{
		ID: "filter-single-field-named-filter", AEP: 160, Level: SHOULD, Static: true,
		Title:      "Filtering is a single 'filter' query field",
		Applicable: func(rc *RunContext) bool { return rc.Resource.Features.Filter },
		Run: func(rc *RunContext) Result {
			ep := rc.Resource.Method(discovery.MethodList)
			if ep != nil && ep.QueryParam("filter") != nil {
				return pass("filter field present")
			}
			return quote(fail("filtering present but no 'filter' query field"),
				"A request message should have exactly one filtering field, string filter.")
		},
	})

	// AEP-160: an invalid filter is rejected with INVALID_ARGUMENT (400).
	Register(Check{
		ID: "filter-invalid-argument", AEP: 160, Level: MUST,
		Title:      "Invalid filter returns 400 INVALID_ARGUMENT",
		Applicable: func(rc *RunContext) bool { return rc.Resource.Features.Filter && rc.Probe != nil },
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, FilterInvalid)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status == 400 {
				return pass("400 on invalid filter")
			}
			return quote(failStatus("invalid filter", i.Resp.Status, 400, i.Evidence()),
				"A non-compliant filter query must error with INVALID_ARGUMENT.")
		},
	})

	// AEP-154: an ETag changes when the resource changes.
	Register(Check{
		ID: "etag-changes-on-update", AEP: 154, Level: MUST,
		Title: "ETag changes after an update",
		Applicable: func(rc *RunContext) bool {
			return rc.Probe != nil && etagOf(rc.Probe.I(Get)) != ""
		},
		Run: func(rc *RunContext) Result {
			before := etagOf(rc.Probe.I(Get))
			after := etagOf(rc.Probe.I(GetAfterUpdate))
			if after == "" {
				return Result{Status: Skipped, Message: "no post-update ETag captured"}
			}
			if before != after {
				return pass("ETag changed after update")
			}
			return quote(fail("ETag did not change after the resource was updated"),
				"ETags must be based on a checksum/hash that changes if the resource changes.")
		},
	})

	// AEP-154: a stale If-Match is rejected with 412 Precondition Failed.
	Register(Check{
		ID: "if-match-mismatch-412", AEP: 154, Level: MUST,
		Title: "Stale If-Match returns 412",
		Applicable: func(rc *RunContext) bool {
			return rc.Probe != nil && rc.Probe.I(IfMatchBad) != nil
		},
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, IfMatchBad)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status == 412 {
				return pass("412 on stale If-Match")
			}
			return quote(failStatus("stale If-Match", i.Resp.Status, 412, i.Evidence()),
				"If the If-Match header value does not match the ETag, the service must reply with an HTTP 412 error.")
		},
	})

	// AEP-217: a List capable of partial failure declares an unreachable field.
	Register(Check{
		ID: "unreachable-field-present", AEP: 217, Level: MUST, Static: true,
		Title:      "List response declares an unreachable field",
		Applicable: func(rc *RunContext) bool { return rc.Resource.Features.Unreachable },
		Run: func(rc *RunContext) Result {
			ep := rc.Resource.Method(discovery.MethodList)
			if ep != nil {
				if resp := ep.Responses["200"]; resp != nil && resp.Schema != nil && resp.Schema.HasProp("unreachable") {
					return pass("unreachable field present")
				}
			}
			return na("no unreachable field")
		},
	})

	// AEP-162: revisions are modeled as a 'revisions' subcollection named
	// {Resource}Revision.
	Register(Check{
		ID: "revision-collection-named-revisions", AEP: 162, Level: MUST, Static: true,
		Title:      "Revisions live under a 'revisions' subcollection",
		Applicable: func(rc *RunContext) bool { return rc.Resource.Features.Revisions },
		Run: func(rc *RunContext) Result {
			base := rc.Resource.CanonicalPattern()
			for _, other := range rc.Model.Resources {
				if strings.HasPrefix(other.CanonicalPattern(), base+"/revisions/") {
					return pass("revisions subcollection present")
				}
			}
			return quote(fail("revisions detected but no /revisions/ subcollection"),
				"The subcollection name for revisions must be 'revisions'.")
		},
	})

	// AEP-214: an expiring resource uses an expire_time Timestamp field.
	Register(Check{
		ID: "expire-time-timestamp", AEP: 214, Level: MUST, Static: true,
		Title:      "Expiration uses an expire_time timestamp",
		Applicable: func(rc *RunContext) bool { return rc.Resource.Features.ExpireTime },
		Run: func(rc *RunContext) Result {
			p := rc.Resource.Schema.Prop("expire_time")
			if p != nil && p.Type == "string" && (p.Format == "date-time" || p.Format == "") {
				return pass("expire_time present")
			}
			return quote(fail("expire_time is not a timestamp string"),
				"APIs conveying an expiration must rely on a Timestamp field called expire_time.")
		},
	})

	// AEP-155: an idempotency key, if supported, is a string field.
	Register(Check{
		ID: "idempotency-key-string", AEP: 155, Level: MAY, Static: true,
		Title:      "idempotency_key is a string field",
		Applicable: func(rc *RunContext) bool { return rc.Resource.Features.IdempotencyKey },
		Run: func(rc *RunContext) Result {
			for _, m := range []discovery.Method{discovery.MethodCreate, discovery.MethodUpdate, discovery.MethodDelete} {
				if ep := rc.Resource.Method(m); ep != nil {
					if q := ep.QueryParam("idempotency_key"); q != nil {
						if q.Schema == nil || q.Schema.Type == "string" {
							return pass("idempotency_key is a string")
						}
						return fail("idempotency_key is not a string")
					}
				}
			}
			return na("no idempotency_key found")
		},
	})
}

func init() {
	// AEP-157: the read_mask / view parameter must be optional.
	Register(Check{
		ID: "read-mask-optional", AEP: 157, Level: MUST, Static: true,
		Title:      "read_mask / view parameter is optional",
		Applicable: func(rc *RunContext) bool { return rc.Resource.Features.ReadMask },
		Run: func(rc *RunContext) Result {
			for _, m := range []discovery.Method{discovery.MethodGet, discovery.MethodList} {
				ep := rc.Resource.Method(m)
				if ep == nil {
					continue
				}
				for _, name := range []string{"read_mask", "view"} {
					if q := ep.QueryParam(name); q != nil && q.Required {
						return quote(failf("%s %q is required; it must be optional", m, name),
							"The read_mask parameter must be optional.")
					}
				}
			}
			return pass("read_mask/view is optional")
		},
	})

	// AEP-134/161: partial update declares an update_mask field.
	Register(Check{
		ID: "update-mask-declared", AEP: 134, Level: MUST, Static: true,
		Title:      "Update declares an update_mask",
		Applicable: func(rc *RunContext) bool { return rc.Resource.Features.FieldMask },
		Run: func(rc *RunContext) Result {
			ep := rc.Resource.Method(discovery.MethodUpdate)
			if ep != nil && (ep.QueryParam("update_mask") != nil ||
				(ep.RequestBodySchema != nil && ep.RequestBodySchema.HasProp("update_mask"))) {
				return pass("update_mask declared")
			}
			return na("no update_mask")
		},
	})
}

func etagOf(i *Interaction) string {
	if i == nil || i.Resp == nil {
		return ""
	}
	return i.Resp.Headers.Get("ETag")
}
