package discovery

import (
	"sort"
	"strings"
)

// build walks the raw document and populates m.Resources: one Resource per
// x-aep-resource schema, with its standard/custom endpoints attached and its
// optional-feature flags detected.
func build(m *APIModel, doc *Document) {
	bySingular := map[string]*Resource{}
	for _, name := range schemaOrder(doc) {
		s := m.Components[name]
		if s == nil || s.XAEPResource == nil {
			continue
		}
		x := s.XAEPResource
		r := &Resource{
			Singular: x.Singular,
			Plural:   x.Plural,
			Type:     x.Type,
			Parents:  x.Parents,
			Patterns: x.Patterns,
			Schema:   s,
			Methods:  map[Method]*Endpoint{},
		}
		m.Resources = append(m.Resources, r)
		if r.Singular != "" {
			bySingular[r.Singular] = r
		}
	}

	// Compute parent-chain depth for ordering (ancestors created first).
	for _, r := range m.Resources {
		r.Depth = depth(r, bySingular, map[string]bool{})
	}
	sort.SliceStable(m.Resources, func(i, j int) bool {
		if m.Resources[i].Depth != m.Resources[j].Depth {
			return m.Resources[i].Depth < m.Resources[j].Depth
		}
		return m.Resources[i].Singular < m.Resources[j].Singular
	})

	// Attach endpoints by matching OpenAPI paths against resource patterns.
	if doc.Paths != nil {
		for _, p := range doc.Paths.Order {
			item := doc.Paths.Map[p]
			for _, vo := range item.operations() {
				attachEndpoint(m, p, vo.Verb, vo.Op)
			}
		}
	}

	for _, r := range m.Resources {
		detectFeatures(m, r)
	}
}

// attachEndpoint classifies one (path, verb, operation) and files it under the
// resource it operates on.
func attachEndpoint(m *APIModel, path, verb string, op *Operation) {
	ep := buildEndpoint(m, path, verb, op)

	// Custom method: path contains a ':verb' suffix on the final segment.
	base, custom, isCustom := strings.Cut(path, ":")
	if isCustom {
		ep.Custom = true
		ep.CustomVerb = custom
		if r := resourceForResourcePath(m, base); r != nil {
			ep.Method = ""
			r.Custom = append(r.Custom, ep)
		}
		return
	}

	// Standard method: match on resource path (has trailing {id}) vs collection.
	if r := resourceForResourcePath(m, path); r != nil {
		switch verb {
		case "GET":
			ep.Method = MethodGet
		case "PATCH":
			ep.Method = MethodUpdate
		case "PUT":
			ep.Method = MethodApply
		case "DELETE":
			ep.Method = MethodDelete
		}
		if ep.Method != "" {
			r.Methods[ep.Method] = ep
		}
		return
	}
	if r := resourceForCollectionPath(m, path); r != nil {
		switch verb {
		case "GET":
			ep.Method = MethodList
		case "POST":
			ep.Method = MethodCreate
		}
		if ep.Method != "" {
			r.Methods[ep.Method] = ep
		}
	}
}

func buildEndpoint(m *APIModel, path, verb string, op *Operation) *Endpoint {
	ep := &Endpoint{
		HTTPVerb:    verb,
		Path:        path,
		OperationID: op.OperationID,
		Responses:   map[string]*Response{},
	}
	for _, p := range op.Parameters {
		param := Param{Name: p.Name, In: p.In, Required: p.Required, Schema: m.resolve(p.Schema)}
		switch p.In {
		case "path":
			ep.PathParams = append(ep.PathParams, param)
		case "query":
			ep.QueryParams = append(ep.QueryParams, param)
		}
	}
	if op.RequestBody != nil {
		ep.RequestBodyRequired = op.RequestBody.Required
		for ct := range op.RequestBody.Content {
			ep.RequestContentTypes = append(ep.RequestContentTypes, ct)
		}
		sort.Strings(ep.RequestContentTypes)
		// Resolve the body schema from JSON or, failing that, merge-patch+json
		// (Update bodies are commonly declared only as application/merge-patch+json).
		if mt, ok := op.RequestBody.Content["application/json"]; ok {
			ep.RequestBodySchema = m.resolve(mt.Schema)
		} else if mt, ok := op.RequestBody.Content["application/merge-patch+json"]; ok {
			ep.RequestBodySchema = m.resolve(mt.Schema)
		}
	}
	for status, resp := range op.Responses {
		r := &Response{Status: status, Description: resp.Description}
		if mt, ok := resp.Content["application/json"]; ok {
			r.Schema = m.resolve(mt.Schema)
		}
		ep.Responses[status] = r
	}
	return ep
}

// resourceForResourcePath returns the resource whose canonical (with-id) path
// equals path.
func resourceForResourcePath(m *APIModel, path string) *Resource {
	path = strings.TrimRight(path, "/")
	for _, r := range m.Resources {
		if "/"+lastPattern(r) == path {
			return r
		}
	}
	return nil
}

// resourceForCollectionPath returns the resource whose collection path
// (pattern minus the trailing id segment) equals path.
func resourceForCollectionPath(m *APIModel, path string) *Resource {
	path = strings.TrimRight(path, "/")
	for _, r := range m.Resources {
		if "/"+collectionPattern(r) == path {
			return r
		}
	}
	return nil
}

// lastPattern returns the resource's canonical path pattern (last declared).
func lastPattern(r *Resource) string {
	if len(r.Patterns) == 0 {
		return ""
	}
	return r.Patterns[len(r.Patterns)-1]
}

// collectionPattern is the resource pattern with the trailing '/{id}' removed.
func collectionPattern(r *Resource) string {
	p := lastPattern(r)
	segs := strings.Split(p, "/")
	if len(segs) > 0 && strings.HasPrefix(segs[len(segs)-1], "{") {
		segs = segs[:len(segs)-1]
	}
	return strings.Join(segs, "/")
}

func depth(r *Resource, bySingular map[string]*Resource, seen map[string]bool) int {
	if len(r.Parents) == 0 || seen[r.Singular] {
		return 0
	}
	seen[r.Singular] = true
	max := 0
	for _, p := range r.Parents {
		if pr, ok := bySingular[p]; ok {
			if d := depth(pr, bySingular, seen) + 1; d > max {
				max = d
			}
		}
	}
	return max
}

// schemaOrder returns component schema names in document order.
func schemaOrder(doc *Document) []string {
	if doc.Components.Schemas == nil {
		return nil
	}
	return doc.Components.Schemas.Order
}

// endpointList returns the resource's standard-method endpoints (non-nil).
func endpointList(r *Resource) []*Endpoint {
	var out []*Endpoint
	for _, m := range []Method{MethodGet, MethodList, MethodCreate, MethodUpdate, MethodApply, MethodDelete} {
		if ep := r.Method(m); ep != nil {
			out = append(out, ep)
		}
	}
	return out
}

// detectFeatures sets the resource's FeatureFlags from its endpoints and schema.
func detectFeatures(m *APIModel, r *Resource) {
	f := &r.Features
	if list := r.Method(MethodList); list != nil {
		f.Pagination = list.QueryParam("page_token") != nil || list.QueryParam("max_page_size") != nil
		f.Skip = list.QueryParam("skip") != nil
		f.Filter = list.QueryParam("filter") != nil
		f.OrderBy = list.QueryParam("order_by") != nil
		f.ShowDeleted = list.QueryParam("show_deleted") != nil
		if list.QueryParam("read_mask") != nil || list.QueryParam("view") != nil {
			f.ReadMask = true
		}
		if resp := list.Responses["200"]; resp != nil && resp.Schema != nil {
			f.Unreachable = resp.Schema.HasProp("unreachable")
		}
	}
	if get := r.Method(MethodGet); get != nil {
		if get.QueryParam("read_mask") != nil || get.QueryParam("view") != nil {
			f.ReadMask = true
		}
	}
	if up := r.Method(MethodUpdate); up != nil {
		f.FieldMask = up.QueryParam("update_mask") != nil ||
			(up.RequestBodySchema != nil && up.RequestBodySchema.HasProp("update_mask"))
		f.AllowMissing = up.QueryParam("allow_missing") != nil
	}
	if del := r.Method(MethodDelete); del != nil {
		f.Force = del.QueryParam("force") != nil
	}
	if cr := r.Method(MethodCreate); cr != nil {
		f.UserSettableID = cr.QueryParam("id") != nil
		f.OverwriteDelete = cr.QueryParam("overwrite_soft_deleted") != nil
	}
	for _, ep := range append([]*Endpoint{r.Method(MethodCreate), r.Method(MethodUpdate), r.Method(MethodDelete)}, r.Custom...) {
		if ep != nil && ep.QueryParam("idempotency_key") != nil {
			f.IdempotencyKey = true
		}
	}
	if r.Schema != nil {
		f.ExpireTime = r.Schema.HasProp("expire_time")
	}
	// Long-running operations: any endpoint (standard or custom) that returns a
	// 202 Accepted response indicates asynchronous / operation-based behavior.
	for _, ep := range append(endpointList(r), r.Custom...) {
		if ep == nil {
			continue
		}
		if _, ok := ep.Responses["202"]; ok {
			f.LRO = true
		}
	}
	// Revisions: any resource whose canonical pattern nests a /revisions/ under
	// this resource's path.
	rp := lastPattern(r)
	for _, other := range m.Resources {
		if other == r {
			continue
		}
		if strings.HasPrefix(lastPattern(other), rp+"/revisions/") {
			f.Revisions = true
		}
	}
}
