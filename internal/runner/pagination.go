package runner

import (
	"net/url"
	"strconv"

	"github.com/thegagne/aep-conformance-test/internal/checks"
	"github.com/thegagne/aep-conformance-test/internal/client"
	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

// pageCap bounds the pagination loop so a server that never returns an empty
// next_page_token (an infinite loop) is detected rather than hanging.
const pageCap = 12

// probePagination seeds extra siblings in the collection, then pages through it
// with a small page size, recording the outcome on the probe.
func (rn *Runner) probePagination(p *checks.Probe, r *discovery.Resource) {
	listEp := r.Method(discovery.MethodList)
	sizeParam := "max_page_size"
	if listEp.QueryParam("max_page_size") == nil && listEp.QueryParam("page_size") != nil {
		sizeParam = "page_size"
	}
	res := &checks.PaginationResult{SizeParam: sizeParam, Ran: true}
	p.Pagination = res

	// Seed two more siblings so the collection spans multiple pages of size 2.
	for range 2 {
		id := rn.Sampler.NextID(r.Singular)
		_, resp, err := rn.Client.Do("POST", p.CollectionPath+"?id="+id, rn.Sampler.Body(r.Schema, false))
		if err == nil && resp != nil && resp.Status < 300 {
			rn.created = append(rn.created, pathFromResponse(&checks.Interaction{Resp: resp}, p.CollectionPath+"/"+id))
		}
	}

	const size = 2
	token := ""
	for res.Pages < pageCap {
		u := p.CollectionPath + "?" + sizeParam + "=" + strconv.Itoa(size)
		if token != "" {
			u += "&page_token=" + url.QueryEscape(token)
		}
		_, resp, err := rn.Client.Do("GET", u, nil)
		if err != nil || resp == nil {
			break
		}
		count := listResultCount(resp)
		next := nextPageToken(resp)
		if res.Pages == 0 {
			res.FirstStatus = resp.Status
			res.FirstCount = count
			res.FirstToken = next
			// Record the first page as an interaction for evidence.
			p.Interactions[checks.PageFirst] = &checks.Interaction{
				Name: checks.PageFirst, Method: "GET", URL: rn.Client.BaseURL + "/" + u, Resp: resp,
			}
		}
		res.Pages++
		res.TotalSeen += count
		if next == "" {
			res.Terminated = true
			break
		}
		token = next
	}

	// Negative page size must be rejected.
	_, resp, err := rn.Client.Do("GET", p.CollectionPath+"?"+sizeParam+"=-1", nil)
	if err == nil && resp != nil {
		res.NegativeRan = true
		res.NegativeStatus = resp.Status
	}
}

// listResultCount returns the number of items in a List response's results array.
func listResultCount(resp *client.Response) int {
	if resp == nil || resp.JSON == nil {
		return 0
	}
	if arr, ok := resp.JSON["results"].([]any); ok {
		return len(arr)
	}
	return 0
}

// nextPageToken extracts the pagination token under either JSON spelling.
func nextPageToken(resp *client.Response) string {
	if resp == nil || resp.JSON == nil {
		return ""
	}
	for _, k := range []string{"next_page_token", "nextPageToken"} {
		if v, ok := resp.JSON[k].(string); ok {
			return v
		}
	}
	return ""
}
