package checks

func paginationRan(rc *RunContext) bool {
	return rc.Resource != nil && rc.Resource.Features.Pagination && rc.Probe != nil
}

// pg returns the pagination result, or a Skipped result if it did not run.
func pg(rc *RunContext) (*PaginationResult, *Result) {
	if rc.Probe == nil || rc.Probe.Pagination == nil || !rc.Probe.Pagination.Ran {
		s := Result{Status: Skipped, Message: "pagination not exercised"}
		return nil, &s
	}
	return rc.Probe.Pagination, nil
}

func init() {
	// AEP-158: the server honors the requested page size.
	Register(Check{
		ID: "page-size-honored", AEP: 158, Level: MUST,
		Title:      "List honors max_page_size",
		Applicable: paginationRan,
		Run: func(rc *RunContext) Result {
			r, skip := pg(rc)
			if skip != nil {
				return *skip
			}
			if r.FirstStatus != 200 {
				return failf("first page returned %d", r.FirstStatus)
			}
			if r.FirstCount <= 2 {
				return pass("first page returned ≤ requested size")
			}
			return quote(failf("requested page size 2 but got %d results", r.FirstCount),
				"max_page_size is the maximum number of resources to return in a single page.")
		},
	})

	// AEP-158: a non-final page provides a next_page_token.
	Register(Check{
		ID: "page-next-token-when-more", AEP: 158, Level: MUST,
		Title:      "next_page_token is set when more results exist",
		Applicable: paginationRan,
		Run: func(rc *RunContext) Result {
			r, skip := pg(rc)
			if skip != nil {
				return *skip
			}
			// We seeded the collection to exceed one page of size 2.
			if r.TotalSeen > 2 && r.FirstToken == "" {
				return quote(fail("more than one page of results but the first page had no next_page_token"),
					"If the end of the collection has not been reached, the API must provide a next_page_token.")
			}
			return pass("next_page_token present when more pages exist")
		},
	})

	// AEP-158: pagination terminates (empty token at the end; no infinite loop).
	Register(Check{
		ID: "page-terminates", AEP: 158, Level: MUST,
		Title:      "Pagination terminates with an empty token",
		Applicable: paginationRan,
		Run: func(rc *RunContext) Result {
			r, skip := pg(rc)
			if skip != nil {
				return *skip
			}
			if r.Terminated {
				return pass("pagination reached end in " + itoa(r.Pages) + " page(s)")
			}
			return quote(failf("pagination did not terminate within %d pages (possible infinite next_page_token)", r.Pages),
				"If the end of the collection has been reached, the next_page_token field must be empty. This is the only way to communicate end-of-collection.")
		},
	})

	// AEP-158: a negative page size is an INVALID_ARGUMENT (400).
	Register(Check{
		ID: "page-size-negative-invalid", AEP: 158, Level: MUST,
		Title:      "Negative page size returns 400 INVALID_ARGUMENT",
		Applicable: paginationRan,
		Run: func(rc *RunContext) Result {
			r, skip := pg(rc)
			if skip != nil {
				return *skip
			}
			if !r.NegativeRan {
				return Result{Status: Skipped, Message: "negative page-size request did not run"}
			}
			if r.NegativeStatus == 400 {
				return pass("400 on negative page size")
			}
			return quote(failf("negative page size returned %d, expected 400", r.NegativeStatus),
				"If the user specifies a negative value for max_page_size, the API must send an INVALID_ARGUMENT error.")
		},
	})
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
