// Package runner drives the conformance run: for each resource it performs the
// scripted lifecycle (creating any parent chain, then Create → Get → List →
// Update → Apply → negatives → Delete) into a Probe, guarantees teardown, and
// then evaluates every registered check against the captured interactions.
package runner

import (
	"net/url"
	"sort"
	"strings"

	"github.com/thegagne/aep-conformance-test/internal/checks"
	"github.com/thegagne/aep-conformance-test/internal/client"
	"github.com/thegagne/aep-conformance-test/internal/discovery"
	"github.com/thegagne/aep-conformance-test/internal/sampler"
)

// Runner executes the suite. Client may be nil to run static checks only.
type Runner struct {
	Model   *discovery.APIModel
	Client  *client.Client
	Sampler *sampler.Sampler

	// created tracks every path created during the run for teardown, in order.
	created []string
}

// New builds a Runner. If baseURL is empty, only static checks run.
func New(model *discovery.APIModel, cl *client.Client) *Runner {
	return &Runner{Model: model, Client: cl, Sampler: sampler.New(nil)}
}

// Run evaluates all checks for the given resources (nil = all discovered) and
// returns the flat list of results. Teardown of created resources always runs.
func (rn *Runner) Run(resources []*discovery.Resource) []checks.Result {
	if resources == nil {
		resources = rn.Model.Resources
	}
	defer rn.teardown()

	var results []checks.Result

	// API-scope checks run once.
	apiCtx := &checks.RunContext{Model: rn.Model}
	for _, c := range checks.All() {
		if c.PerAPI {
			results = append(results, checks.Evaluate(c, apiCtx))
		}
	}

	// Resource-scope checks run per resource, driven by a live probe.
	for _, r := range resources {
		var probe *checks.Probe
		if rn.Client != nil {
			probe = rn.probe(r)
		}
		rc := &checks.RunContext{Model: rn.Model, Resource: r, Probe: probe}
		for _, c := range checks.All() {
			if c.PerAPI {
				continue
			}
			results = append(results, checks.Evaluate(c, rc))
		}
	}
	return results
}

// probe performs the lifecycle for a single resource and returns the captured
// interactions. Steps are guarded by the presence of each standard method.
func (rn *Runner) probe(r *discovery.Resource) *checks.Probe {
	p := &checks.Probe{Resource: r, Interactions: map[string]*checks.Interaction{}}

	// A resource with no Create method can't be exercised dynamically.
	if r.Method(discovery.MethodCreate) == nil {
		return p
	}

	parentPath, ok := rn.ensureParents(r)
	if !ok {
		return p
	}
	p.ParentPath = parentPath
	p.CollectionPath = joinPath(parentPath, r.CollectionID())

	// Create the primary resource.
	id := rn.Sampler.NextID(r.Singular)
	p.CreatedID = id
	body := rn.Sampler.Body(r.Schema, false)
	p.SentBody = body
	createPath := p.CollectionPath + "?id=" + id
	rn.do(p, checks.Create, "POST", createPath, body)
	created := p.I(checks.Create)
	if created.Err != nil || created.Resp == nil || created.Resp.Status >= 300 {
		return p // create failed; dependent steps would be meaningless
	}
	p.CreatedPath = pathFromResponse(created, p.CollectionPath+"/"+id)
	rn.created = append(rn.created, p.CreatedPath)

	// Get, then Get again (side-effect check).
	if r.Method(discovery.MethodGet) != nil {
		rn.do(p, checks.Get, "GET", p.CreatedPath, nil)
		rn.do(p, checks.GetRepeat, "GET", p.CreatedPath, nil)
	}
	// List the collection.
	if r.Method(discovery.MethodList) != nil {
		rn.do(p, checks.List, "GET", p.CollectionPath, nil)
	}
	// Update (partial): change one mutable field, then re-Get to observe the
	// merge (unspecified fields preserved, specified field changed).
	if r.Method(discovery.MethodUpdate) != nil {
		rn.do(p, checks.Update, "PATCH", p.CreatedPath, rn.mutation(r))
		if r.Method(discovery.MethodGet) != nil {
			rn.do(p, checks.GetAfterUpdate, "GET", p.CreatedPath, nil)
		}
	}
	// Apply (PUT). Beyond the basic apply, this exercises the AEP-137 semantics:
	// strong consistency (Get reflects the apply), field preservation (a field
	// omitted from a later apply is not cleared), apply-as-create (a fresh path
	// returns 201), and the required-field failure (400).
	if r.Method(discovery.MethodApply) != nil {
		rn.probeApply(p, r)
	}

	// Create with a bogus 'path' in the body: the server must ignore it.
	if r.Features.UserSettableID {
		echoID := rn.Sampler.NextID(r.Singular)
		echoBody := rn.Sampler.Body(r.Schema, false)
		echoBody["path"] = "bogus/" + r.CollectionID() + "/injected"
		rn.do(p, checks.CreateWithPathEcho, "POST", p.CollectionPath+"?id="+echoID, echoBody)
		if e := p.I(checks.CreateWithPathEcho); e.Resp != nil && e.Resp.Status < 300 {
			rn.created = append(rn.created, pathFromResponse(e, p.CollectionPath+"/"+echoID))
		}
	}

	// Pagination: seed extra siblings and page through the collection.
	if r.Features.Pagination && r.Method(discovery.MethodList) != nil {
		rn.probePagination(p, r)
	}

	// Conditional pattern probes, gated on detected features.
	if r.Features.Filter && r.Method(discovery.MethodList) != nil {
		rn.do(p, checks.FilterInvalid, "GET", p.CollectionPath+"?filter="+url.QueryEscape("=="), nil)
	}
	rn.probeETag(p, r)
	rn.probeAcrossCollections(p, r)

	// Negatives that don't disturb the created resource.
	missingID := rn.Sampler.NextID(r.Singular) + "-missing"
	rn.do(p, checks.GetMissing, "GET", joinPath(p.CollectionPath, missingID), nil)
	rn.do(p, checks.DuplicateCreate, "POST", p.CollectionPath+"?id="+id, body)
	if r.Method(discovery.MethodUpdate) != nil {
		// AEP-134: updating a missing resource must return NOT_FOUND (404).
		rn.do(p, checks.UpdateMissing, "PATCH", joinPath(p.CollectionPath, missingID), rn.mutation(r))
	}
	if r.Method(discovery.MethodDelete) != nil {
		rn.do(p, checks.DeleteMissing, "DELETE", joinPath(p.CollectionPath, missingID), nil)
		// AEP-135: a DELETE that carries a body must ignore it, not error. Use a
		// throwaway sibling so the primary lifecycle is undisturbed.
		rn.probeDeleteWithBody(p, r)
	}

	// Delete, then confirm 404 steady-state.
	if r.Method(discovery.MethodDelete) != nil {
		rn.do(p, checks.Delete, "DELETE", p.CreatedPath, nil)
		if p.I(checks.Delete).Resp != nil && p.I(checks.Delete).Resp.Status < 300 {
			rn.forget(p.CreatedPath) // deleted; drop from teardown list
		}
		rn.do(p, checks.GetAfterDelete, "GET", p.CreatedPath, nil)
	}
	return p
}

// ensureParents creates the full ancestor chain for r and returns the concrete
// parent path (empty for top-level). ok is false if a parent couldn't be made.
func (rn *Runner) ensureParents(r *discovery.Resource) (string, bool) {
	if len(r.Parents) == 0 {
		return "", true
	}
	parent := rn.Model.ResourceBySingular(r.Parents[0])
	if parent == nil || parent.Method(discovery.MethodCreate) == nil {
		return "", false
	}
	path, ok := rn.createInstance(parent)
	return path, ok
}

// createInstance creates one instance of r (recursively creating its parents)
// and returns its concrete resource path.
func (rn *Runner) createInstance(r *discovery.Resource) (string, bool) {
	parentPath, ok := rn.ensureParents(r)
	if !ok {
		return "", false
	}
	collection := joinPath(parentPath, r.CollectionID())
	id := rn.Sampler.NextID(r.Singular)
	body := rn.Sampler.Body(r.Schema, false)
	_, resp, err := rn.Client.Do("POST", collection+"?id="+id, body)
	if err != nil || resp == nil || resp.Status >= 300 {
		return "", false
	}
	tmp := &checks.Interaction{Resp: resp}
	path := pathFromResponse(tmp, collection+"/"+id)
	rn.created = append(rn.created, path)
	return path, true
}

// mutation returns a small PATCH body that changes one field of the resource.
func (rn *Runner) mutation(r *discovery.Resource) map[string]any {
	if r.Schema == nil || r.Schema.Properties == nil {
		return map[string]any{}
	}
	for _, name := range r.Schema.Properties.Order {
		prop := r.Schema.Properties.Map[name]
		if prop == nil || prop.ReadOnly || name == "path" {
			continue
		}
		switch prop.Type {
		case "string":
			return map[string]any{name: "updated"}
		case "integer", "number":
			return map[string]any{name: 2}
		case "boolean":
			return map[string]any{name: false}
		}
	}
	return map[string]any{}
}

// do issues a request and records it as a named interaction on the probe.
func (rn *Runner) do(p *checks.Probe, name, method, path string, body any, headers ...map[string]string) {
	req, resp, err := rn.Client.Do(method, path, body, headers...)
	p.Interactions[name] = &checks.Interaction{
		Name: name, Method: method, URL: req.URL, ReqBody: req.Body, Resp: resp, Err: err,
	}
}

// probeDeleteWithBody creates a throwaway sibling and deletes it with a request
// body present, so the AEP-135 "body must be ignored" check has evidence. If the
// delete does not succeed, the sibling is queued for teardown.
func (rn *Runner) probeDeleteWithBody(p *checks.Probe, r *discovery.Resource) {
	seedID := rn.Sampler.NextID(r.Singular)
	rn.do(p, checks.DeleteBodySeed, "POST", p.CollectionPath+"?id="+seedID, rn.Sampler.Body(r.Schema, false))
	seed := p.I(checks.DeleteBodySeed)
	if seed.Resp == nil || seed.Resp.Status >= 300 {
		return // couldn't seed; the check will report Skipped
	}
	seedPath := pathFromResponse(seed, p.CollectionPath+"/"+seedID)
	rn.do(p, checks.DeleteWithBody, "DELETE", seedPath, map[string]any{"aepConformanceIgnoredField": "x"})
	if d := p.I(checks.DeleteWithBody); d.Resp == nil || d.Resp.Status >= 300 {
		rn.created = append(rn.created, seedPath) // not deleted; ensure teardown
	}
}

// probeApply drives the AEP-137 lifecycle for a resource with an Apply method.
func (rn *Runner) probeApply(p *checks.Probe, r *discovery.Resource) {
	// Primary apply-as-update on the existing resource, then observe via Get.
	rn.do(p, checks.Apply, "PUT", p.CreatedPath, rn.Sampler.Body(r.Schema, false))
	if r.Method(discovery.MethodGet) != nil {
		rn.do(p, checks.GetAfterApply, "GET", p.CreatedPath, nil)
	}

	// Field preservation: set an optional field, then apply again without it and
	// confirm (via Get) that it was preserved rather than cleared.
	if drop := firstOptionalMutableField(r.Schema); drop != "" && r.Method(discovery.MethodGet) != nil {
		rn.do(p, checks.ApplySetOptional, "PUT", p.CreatedPath, rn.Sampler.Body(r.Schema, true))
		partial := rn.Sampler.Body(r.Schema, true)
		delete(partial, drop)
		rn.do(p, checks.ApplyDropOptional, "PUT", p.CreatedPath, partial)
		rn.do(p, checks.GetAfterApplyPartial, "GET", p.CreatedPath, nil)
	}

	// Apply-as-create: PUT to a fresh path should create the resource (201).
	newID := rn.Sampler.NextID(r.Singular)
	newPath := joinPath(p.CollectionPath, newID)
	rn.do(p, checks.ApplyCreate, "PUT", newPath, rn.Sampler.Body(r.Schema, true))
	if e := p.I(checks.ApplyCreate); e.Resp != nil && e.Resp.Status < 300 {
		rn.created = append(rn.created, pathFromResponse(e, newPath))
	}

	// Missing a required field must fail with 400.
	if req := firstRequiredMutableField(r.Schema); req != "" {
		bad := rn.Sampler.Body(r.Schema, true)
		delete(bad, req)
		rn.do(p, checks.ApplyMissingRequired, "PUT", p.CreatedPath, bad)
	}
}

// firstOptionalMutableField returns the first writable, non-required field name.
func firstOptionalMutableField(s *discovery.Schema) string {
	return firstField(s, func(required bool) bool { return !required })
}

// firstRequiredMutableField returns the first writable, required field name.
func firstRequiredMutableField(s *discovery.Schema) string {
	return firstField(s, func(required bool) bool { return required })
}

func firstField(s *discovery.Schema, want func(required bool) bool) string {
	if s == nil || s.Properties == nil {
		return ""
	}
	for _, name := range s.Properties.Order {
		prop := s.Properties.Map[name]
		if prop == nil || prop.ReadOnly || name == "path" {
			continue
		}
		if want(s.IsRequired(name)) {
			return name
		}
	}
	return ""
}

// probeAcrossCollections exercises AEP-159: for a resource with a parent, it
// seeds a second instance under a *different* parent, then issues a List with
// the immediate parent id replaced by the '-' wildcard. The check verifies the
// returned resources carry canonical (real) parent ids rather than '-'. If the
// server rejects the wildcard, the feature is simply unsupported (reported N/A).
func (rn *Runner) probeAcrossCollections(p *checks.Probe, r *discovery.Resource) {
	if len(r.Parents) == 0 || r.Method(discovery.MethodList) == nil || p.ParentPath == "" {
		return
	}
	// Best-effort: a second instance under its own fresh parent chain, so the
	// wildcard genuinely spans more than one collection where possible.
	rn.createInstance(r)

	wildParent := replaceLastSegment(p.ParentPath, "-")
	wildPath := joinPath(wildParent, r.CollectionID())
	rn.do(p, checks.ListAcrossCollections, "GET", wildPath, nil)
}

// probeETag exercises optimistic concurrency when the resource returns an ETag:
// a stale If-Match must be rejected with 412.
func (rn *Runner) probeETag(p *checks.Probe, r *discovery.Resource) {
	get := p.I(checks.Get)
	if get == nil || get.Resp == nil || get.Resp.Headers.Get("ETag") == "" {
		return // resource does not use ETags; conditional checks report N/A
	}
	if r.Method(discovery.MethodUpdate) == nil {
		return
	}
	rn.do(p, checks.IfMatchBad, "PATCH", p.CreatedPath, rn.mutation(r),
		map[string]string{"If-Match": "\"aepconf-stale-etag\""})
}

// teardown deletes everything created, children before parents.
func (rn *Runner) teardown() {
	if rn.Client == nil {
		return
	}
	for i := len(rn.created) - 1; i >= 0; i-- {
		_, _, _ = rn.Client.Do("DELETE", rn.created[i], nil) // best-effort cleanup
	}
	rn.created = nil
}

func (rn *Runner) forget(path string) {
	out := rn.created[:0]
	for _, p := range rn.created {
		if p != path {
			out = append(out, p)
		}
	}
	rn.created = out
}

// pathFromResponse returns the server-assigned path from a create/response body
// if present, else the fallback.
func pathFromResponse(i *checks.Interaction, fallback string) string {
	if i != nil && i.Resp != nil && i.Resp.JSON != nil {
		if v, ok := i.Resp.JSON["path"].(string); ok && v != "" {
			return strings.TrimLeft(v, "/")
		}
	}
	return fallback
}

func joinPath(base, seg string) string {
	if base == "" {
		return seg
	}
	return strings.TrimRight(base, "/") + "/" + seg
}

// replaceLastSegment swaps the final '/'-delimited segment of path for seg,
// e.g. replaceLastSegment("publishers/pub-1", "-") == "publishers/-".
func replaceLastSegment(path, seg string) string {
	i := strings.LastIndex(path, "/")
	if i < 0 {
		return seg
	}
	return path[:i+1] + seg
}

// SortByDepth orders resources ancestors-first (stable), matching discovery.
func SortByDepth(rs []*discovery.Resource) {
	sort.SliceStable(rs, func(i, j int) bool { return rs[i].Depth < rs[j].Depth })
}
