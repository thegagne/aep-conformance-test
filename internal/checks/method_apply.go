package checks

import (
	"fmt"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

// AEP-137 Apply semantics. Apply is a declarative PUT that creates-or-replaces a
// resource, but — unlike a plain REST PUT — it preserves fields omitted from the
// request. These checks exercise the status codes, strong consistency, field
// preservation, and required-field failure. All are gated on the resource having
// an Apply method, so non-Apply APIs report NotApplicable.

func hasApply(rc *RunContext) bool { return rc.endpoint(discovery.MethodApply) != nil }

func init() {
	// AEP-137: applying to an existing resource returns 200.
	Register(Check{
		ID: "apply-update-200", AEP: 137, Level: MUST,
		Title:      "Apply to an existing resource returns 200",
		Applicable: hasApply,
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, Apply)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status == 200 {
				return pass("200 OK on apply-update")
			}
			return quote(failf("apply-update returned %d, expected 200", i.Resp.Status),
				"If the resource is updated, the response must return a 200 status code.")
		},
	})

	// AEP-137: apply is strongly consistent — a subsequent Get reflects it.
	Register(Check{
		ID: "apply-consistent-on-get", AEP: 137, Level: MUST,
		Title: "Apply is reflected by a subsequent Get",
		Applicable: func(rc *RunContext) bool {
			return hasApply(rc) && rc.endpoint(discovery.MethodGet) != nil
		},
		Run: func(rc *RunContext) Result {
			ap, skip := need(rc.Probe, Apply)
			if skip != nil {
				return *skip
			}
			get, skip := need(rc.Probe, GetAfterApply)
			if skip != nil {
				return *skip
			}
			sent := parseJSONObject(ap.ReqBody)
			if sent == nil || get.Resp.JSON == nil {
				return Result{Status: Skipped, Message: "no apply/get data to compare"}
			}
			for k, want := range sent {
				if p := rc.Resource.Schema.Prop(k); p != nil && p.ReadOnly {
					continue
				}
				if got, ok := get.Resp.JSON[k]; !ok || !jsonEqualish(want, got) {
					return quote(failf("applied field %q not observed on Get (want %v, got %v)", k, want, get.Resp.JSON[k]),
						"The operation must have strong consistency.")
				}
			}
			return pass("apply observed on subsequent Get")
		},
	})

	// AEP-137: apply should return the fully-populated resource.
	Register(Check{
		ID: "apply-fully-populated", AEP: 137, Level: SHOULD,
		Title:      "Apply returns a fully-populated resource",
		Applicable: hasApply,
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, Apply)
			if skip != nil {
				return *skip
			}
			if i.Resp.JSON == nil {
				return fail("apply response is not a JSON object", i.Evidence())
			}
			var missing []string
			walkTop(rc.Resource.Schema, func(name string, s *discovery.Schema) {
				if s.ReadOnly {
					if _, ok := i.Resp.JSON[name]; !ok {
						missing = append(missing, name)
					}
				}
			})
			if len(missing) > 0 {
				return quote(failf("declared output-only field(s) absent from Apply response: %v", missing),
					"The response should include the fully-populated resource.")
			}
			return pass("apply response fully populated")
		},
	})

	// AEP-137: a field omitted from an apply must not be modified (preserved).
	Register(Check{
		ID: "apply-preserves-absent-fields", AEP: 137, Level: MUST,
		Title:      "Apply preserves fields omitted from the request",
		Applicable: hasApply,
		Run: func(rc *RunContext) Result {
			set, skip := need(rc.Probe, ApplySetOptional)
			if skip != nil {
				return *skip
			}
			drop, skip := need(rc.Probe, ApplyDropOptional)
			if skip != nil {
				return *skip
			}
			after, skip := need(rc.Probe, GetAfterApplyPartial)
			if skip != nil {
				return *skip
			}
			if set.Resp.Status >= 300 || drop.Resp.Status >= 300 || after.Resp.JSON == nil {
				return na("could not establish an apply/preserve comparison")
			}
			setBody := parseJSONObject(set.ReqBody)
			dropBody := parseJSONObject(drop.ReqBody)
			if setBody == nil {
				return na("no set-body to derive the omitted field")
			}
			for name, was := range setBody {
				if _, stillSent := dropBody[name]; stillSent {
					continue // field was not omitted
				}
				if p := rc.Resource.Schema.Prop(name); p == nil || p.ReadOnly || name == "path" {
					continue
				}
				got, ok := after.Resp.JSON[name]
				if !ok {
					return quote(failf("omitted field %q was cleared by Apply (absent after)", name),
						"If a field in the request is not present, the service must not modify this field.")
				}
				if !jsonEqualish(was, got) {
					return quote(failf("omitted field %q was modified by Apply (%v → %v)", name, was, got),
						"If a field in the request is not present, the service must not modify this field.")
				}
				return pass(fmt.Sprintf("omitted field %q preserved across apply", name))
			}
			return na("no omitted field available to verify preservation")
		},
	})

	// AEP-137: applying to a path that doesn't exist creates the resource (201).
	Register(Check{
		ID: "apply-create-201", AEP: 137, Level: MUST,
		Title:      "Apply to a new path creates the resource (201)",
		Applicable: hasApply,
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, ApplyCreate)
			if skip != nil {
				return *skip
			}
			switch {
			case i.Resp.Status == 201:
				return pass("201 Created on apply-create")
			case i.Resp.Status >= 400:
				return na(fmt.Sprintf("apply-as-create not supported (returned %d)", i.Resp.Status))
			default:
				return quote(failf("apply created a resource but returned %d, expected 201", i.Resp.Status),
					"If the resource is created, the response must return a 201 status code.")
			}
		},
	})

	// AEP-137: an apply missing a required field must fail with 400.
	Register(Check{
		ID: "apply-missing-required-400", AEP: 137, Level: MUST,
		Title: "Apply missing a required field returns 400",
		Applicable: func(rc *RunContext) bool {
			return hasApply(rc) && firstRequiredField(rc.Resource.Schema) != ""
		},
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, ApplyMissingRequired)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status == 400 {
				return pass("400 on apply missing a required field")
			}
			return quote(failf("apply missing a required field returned %d, expected 400", i.Resp.Status),
				"If a resource does not have a field set which is labeled with required, the request must fail with INVALID_ARGUMENT / 400 Bad Request.")
		},
	})
}

// firstRequiredField returns the first writable required field of a schema, or
// "" when none exists (used to gate the missing-required check).
func firstRequiredField(s *discovery.Schema) string {
	if s == nil || s.Properties == nil {
		return ""
	}
	for _, name := range s.Properties.Order {
		p := s.Properties.Map[name]
		if p == nil || p.ReadOnly || name == "path" {
			continue
		}
		if s.IsRequired(name) {
			return name
		}
	}
	return ""
}
