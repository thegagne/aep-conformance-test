package checks

import (
	"strings"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

func init() {
	// AEP-133: the path field on a create request body must be ignored.
	Register(Check{
		ID: "create-path-field-ignored", AEP: 133, Level: MUST,
		Title: "Create ignores a client-supplied path field",
		Applicable: func(rc *RunContext) bool {
			return rc.endpoint(discovery.MethodCreate) != nil && rc.Resource.Features.UserSettableID
		},
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, CreateWithPathEcho)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status >= 300 || i.Resp.JSON == nil {
				return fail("create-with-path echo did not return a resource", i.Evidence())
			}
			got, _ := i.Resp.JSON["path"].(string)
			if strings.Contains(got, "bogus/") {
				return quote(failf("server honored client-supplied path: %q", got),
					"The path field on the resource must be ignored during Create.")
			}
			return pass("client path ignored; server path = " + got)
		},
	})

	// AEP-134: PATCH performs a partial merge — fields not in the request are
	// left unchanged (compared before/after via Get).
	Register(Check{
		ID: "update-partial-merge", AEP: 134, Level: SHOULD,
		Title: "Update merges partially (unspecified fields preserved)",
		Applicable: func(rc *RunContext) bool {
			return rc.endpoint(discovery.MethodUpdate) != nil && rc.endpoint(discovery.MethodGet) != nil
		},
		Run: func(rc *RunContext) Result {
			before, skip := need(rc.Probe, Get)
			if skip != nil {
				return *skip
			}
			after, skip := need(rc.Probe, GetAfterUpdate)
			if skip != nil {
				return *skip
			}
			upd := rc.Probe.I(Update)
			if before.Resp.JSON == nil || after.Resp.JSON == nil || upd == nil {
				return Result{Status: Skipped, Message: "insufficient data to compare merge"}
			}
			changed := parseJSONObject(upd.ReqBody)
			for k, was := range before.Resp.JSON {
				if k == "update_mask" {
					continue
				}
				if _, inPatch := changed[k]; inPatch {
					continue // field was intentionally updated
				}
				// Skip server-managed output-only fields (e.g. update_time),
				// which are expected to change on any update.
				if p := rc.Resource.Schema.Prop(k); p != nil && p.ReadOnly {
					continue
				}
				if now, ok := after.Resp.JSON[k]; ok && !jsonEqualish(was, now) {
					return quote(failf("field %q changed by an unrelated PATCH (%v → %v)", k, was, now),
						"The method should support partial resource update; unspecified fields must not be modified.")
				}
			}
			return pass("unspecified fields preserved across update")
		},
	})

	// AEP-134: the update is observable via a subsequent Get (strong consistency).
	Register(Check{
		ID: "update-consistent-on-get", AEP: 134, Level: MUST,
		Title: "Update is reflected by a subsequent Get",
		Applicable: func(rc *RunContext) bool {
			return rc.endpoint(discovery.MethodUpdate) != nil && rc.endpoint(discovery.MethodGet) != nil
		},
		Run: func(rc *RunContext) Result {
			after, skip := need(rc.Probe, GetAfterUpdate)
			if skip != nil {
				return *skip
			}
			upd := rc.Probe.I(Update)
			if upd == nil || after.Resp.JSON == nil {
				return Result{Status: Skipped, Message: "no update/get data"}
			}
			changed := parseJSONObject(upd.ReqBody)
			for k, want := range changed {
				if k == "update_mask" {
					continue
				}
				if got, ok := after.Resp.JSON[k]; !ok || !jsonEqualish(want, got) {
					return quote(failf("updated field %q not observed on Get (want %v, got %v)", k, want, after.Resp.JSON[k]),
						"The operation must have strong consistency.")
				}
			}
			return pass("update observed on subsequent Get")
		},
	})

	// AEP-131: Get returns a fully-populated resource (at least the path plus the
	// server-assigned standard fields it declares).
	Register(Check{
		ID: "get-fully-populated", AEP: 131, Level: SHOULD,
		Title:      "Get returns a fully-populated resource",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodGet) != nil },
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, Get)
			if skip != nil {
				return *skip
			}
			if i.Resp.JSON == nil {
				return fail("get body is not a JSON object", i.Evidence())
			}
			// Every output-only field declared on the resource should be present.
			var missing []string
			walkTop(rc.Resource.Schema, func(name string, s *discovery.Schema) {
				if s.ReadOnly {
					if _, ok := i.Resp.JSON[name]; !ok {
						missing = append(missing, name)
					}
				}
			})
			if len(missing) > 0 {
				return quote(failf("declared output-only field(s) absent from Get: %s", strings.Join(missing, ", ")),
					"The response should usually include the fully-populated resource.")
			}
			return pass("resource fully populated")
		},
	})
}

// walkTop visits only the top-level properties of a schema.
func walkTop(s *discovery.Schema, fn func(name string, s *discovery.Schema)) {
	if s == nil || s.Properties == nil {
		return
	}
	for _, name := range s.Properties.Order {
		if p := s.Properties.Map[name]; p != nil {
			fn(name, p)
		}
	}
}
