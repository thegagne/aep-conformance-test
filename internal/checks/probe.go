package checks

import (
	"github.com/thegagne/aep-conformance-test/internal/client"
	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

// Interaction names captured by the runner and read by dynamic checks.
const (
	Create             = "create"
	Get                = "get"
	GetRepeat          = "get_repeat"
	List               = "list"
	Update             = "update"
	Apply              = "apply"
	Delete             = "delete"
	GetAfterDelete     = "get_after_delete"
	GetMissing         = "get_missing"
	DeleteMissing      = "delete_missing"
	DuplicateCreate    = "duplicate_create"
	MalformedCreate    = "malformed_create"
	PageSizeNegative   = "page_size_negative"
	PageFirst          = "page_first"
	PageSecond         = "page_second"
	CreateWithPathEcho = "create_with_path_echo"
	GetAfterUpdate     = "get_after_update"
	FilterInvalid      = "filter_invalid"
	IfMatchBad         = "if_match_bad"
	UpdateMissing         = "update_missing"
	DeleteBodySeed        = "delete_body_seed"
	DeleteWithBody        = "delete_with_body"
	ListAcrossCollections = "list_across_collections"
	GetAfterApply         = "get_after_apply"
	ApplySetOptional      = "apply_set_optional"
	ApplyDropOptional     = "apply_drop_optional"
	GetAfterApplyPartial  = "get_after_apply_partial"
	ApplyCreate           = "apply_create"
	ApplyMissingRequired  = "apply_missing_required"
)

// Interaction is a single captured request/response.
type Interaction struct {
	Name    string
	Method  string
	URL     string
	ReqBody string
	Resp    *client.Response
	Err     error
}

// Evidence converts an interaction into check Evidence.
func (i *Interaction) Evidence() Evidence {
	if i == nil {
		return Evidence{}
	}
	ev := Evidence{Method: i.Method, URL: i.URL, Request: i.ReqBody}
	if i.Resp != nil {
		ev.Status = i.Resp.Status
		ev.Response = truncate(i.Resp.Body, 600)
	}
	return ev
}

// Probe holds the live interactions performed for one resource, plus the paths
// and body the runner used, so checks can make assertions without re-issuing I/O.
type Probe struct {
	Resource       *discovery.Resource
	ParentPath     string         // created parent chain, e.g. publishers/aepc-1
	CollectionPath string         // e.g. publishers/aepc-1/books
	CreatedPath    string         // e.g. publishers/aepc-1/books/aepc-2
	CreatedID      string         // the id used on create
	SentBody       map[string]any // body sent on the primary create
	Interactions   map[string]*Interaction
	Pagination     *PaginationResult
}

// PaginationResult records the outcome of paging through a collection with a
// small page size, so pagination checks can assert without re-issuing I/O.
type PaginationResult struct {
	SizeParam      string // the size parameter used (max_page_size or page_size)
	Ran            bool
	FirstStatus    int
	FirstCount     int    // results on the first page
	FirstToken     string // next_page_token from the first page
	Pages          int    // number of pages fetched
	TotalSeen      int    // total results across all pages
	Terminated     bool   // pagination ended with an empty token (no infinite loop)
	NegativeStatus int    // status when requesting a negative page size (-1)
	NegativeRan    bool
}

// I returns a captured interaction by name (nil if absent).
func (p *Probe) I(name string) *Interaction {
	if p == nil {
		return nil
	}
	return p.Interactions[name]
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
