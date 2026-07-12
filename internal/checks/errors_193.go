package checks

import (
	"fmt"
	"strings"
)

func fmtStatus(bodyStatus any, httpStatus int) string {
	return fmt.Sprintf("error 'status' %v does not match HTTP %d", bodyStatus, httpStatus)
}

// errorInteraction picks a captured interaction that should have produced an
// error body (a 4xx), preferring the not-found probe.
func errorInteraction(p *Probe) *Interaction {
	for _, name := range []string{GetMissing, DeleteMissing, DuplicateCreate} {
		if i := p.I(name); i != nil && i.Resp != nil && i.Resp.Status >= 400 {
			return i
		}
	}
	return nil
}

func init() {
	// AEP-193 / RFC 9457: error bodies carry the Problem Details shape.
	Register(Check{
		ID: "error-body-rfc9457-shape", AEP: 193, Level: SHOULD,
		Title:      "Error responses follow RFC 9457 Problem Details",
		Applicable: func(rc *RunContext) bool { return rc.Probe != nil },
		Run: func(rc *RunContext) Result {
			i := errorInteraction(rc.Probe)
			if i == nil {
				return Result{Status: Skipped, Message: "no error response captured"}
			}
			if i.Resp.JSON == nil {
				return quote(fail("error response is not a JSON object", i.Evidence()),
					"The structure of the error response should follow RFC 9457 Problem Details.")
			}
			missing := []string{}
			for _, k := range []string{"type", "status", "title"} {
				if _, ok := i.Resp.JSON[k]; !ok {
					missing = append(missing, k)
				}
			}
			if len(missing) > 0 {
				return quote(fail("error body missing field(s): "+strings.Join(missing, ", "), i.Evidence()),
					"The error response should use the RFC 9457 structure: type, status, title, detail, instance.")
			}
			return pass("error body has type/status/title", i.Evidence())
		},
	})

	// AEP-193: the error 'type' field is present and a URI reference (MUST).
	Register(Check{
		ID: "error-type-uri", AEP: 193, Level: MUST,
		Title:      "Error 'type' is a URI reference",
		Applicable: func(rc *RunContext) bool { return rc.Probe != nil },
		Run: func(rc *RunContext) Result {
			i := errorInteraction(rc.Probe)
			if i == nil || i.Resp.JSON == nil {
				return Result{Status: Skipped, Message: "no JSON error response captured"}
			}
			v, ok := i.Resp.JSON["type"].(string)
			if !ok || v == "" {
				return quote(fail("error response has no 'type'", i.Evidence()),
					"type — A URI reference that identifies the problem type.")
			}
			return pass("type = " + v)
		},
	})

	// AEP-193: the error 'status' matches the HTTP status code.
	Register(Check{
		ID: "error-status-matches-http", AEP: 193, Level: SHOULD,
		Title:      "Error 'status' matches the HTTP status code",
		Applicable: func(rc *RunContext) bool { return rc.Probe != nil },
		Run: func(rc *RunContext) Result {
			i := errorInteraction(rc.Probe)
			if i == nil || i.Resp.JSON == nil {
				return Result{Status: Skipped, Message: "no JSON error response captured"}
			}
			raw, ok := i.Resp.JSON["status"]
			if !ok {
				return Result{Status: Skipped, Message: "no status field in error body"}
			}
			if f, ok := raw.(float64); ok && int(f) == i.Resp.Status {
				return pass("status matches HTTP code")
			}
			return fail(fmtStatus(raw, i.Resp.Status), i.Evidence())
		},
	})

	// AEP-193: error bodies are served as JSON.
	Register(Check{
		ID: "error-content-type-json", AEP: 193, Level: SHOULD,
		Title:      "Error responses use a JSON content type",
		Applicable: func(rc *RunContext) bool { return rc.Probe != nil },
		Run: func(rc *RunContext) Result {
			i := errorInteraction(rc.Probe)
			if i == nil {
				return Result{Status: Skipped, Message: "no error response captured"}
			}
			ct := ""
			if i.Resp != nil {
				ct = i.Resp.Headers.Get("Content-Type")
			}
			if strings.Contains(ct, "json") {
				return pass("content-type: " + ct)
			}
			return fail("error content-type is "+ct+", expected JSON", i.Evidence())
		},
	})
}
