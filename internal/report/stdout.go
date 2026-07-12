package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/thegagne/aep-conformance-test/internal/checks"
)

var (
	cPass = color.New(color.FgGreen).SprintFunc()
	cFail = color.New(color.FgRed, color.Bold).SprintFunc()
	cWarn = color.New(color.FgYellow).SprintFunc()
	cNA   = color.New(color.FgHiBlack).SprintFunc()
	cHead = color.New(color.Bold, color.FgCyan).SprintFunc()
)

func (s Severity) colored() string {
	switch s {
	case SevPass:
		return cPass("PASS")
	case SevFail:
		return cFail("FAIL")
	case SevWarn:
		return cWarn("WARN")
	case SevNA:
		return cNA("N/A ")
	case SevSkip:
		return cNA("SKIP")
	}
	return "?"
}

// WriteStdout renders a human-readable report. Non-passing results (FAIL/WARN)
// are shown in full; passes and N/A are summarized per section.
func (r *Report) WriteStdout(w io.Writer, verbose bool) {
	fmt.Fprintf(w, "%s\n", cHead("AEP Conformance Report"))
	if r.Title != "" {
		fmt.Fprintf(w, "API: %s\n", r.Title)
	}
	if r.BaseURL != "" {
		fmt.Fprintf(w, "Target: %s\n", r.BaseURL)
	}
	fmt.Fprintln(w)

	api, groups := r.Grouped()
	if len(api) > 0 {
		r.writeBlock(w, "API", api, nil, verbose)
	}
	for _, g := range groups {
		r.writeBlock(w, g.Collection, g.Results, g.Capabilities, verbose)
	}

	sum := r.Summary()
	fmt.Fprintf(w, "%s  %s %d   %s %d   %s %d   %s %d   %s %d\n",
		cHead("Totals:"),
		cPass("pass"), sum.Pass,
		cFail("fail"), sum.Fail,
		cWarn("warn"), sum.Warn,
		cNA("n/a"), sum.NA,
		cNA("skip"), sum.Skip,
	)
	if r.Conformant() {
		fmt.Fprintf(w, "%s\n", cPass("CONFORMANT — no required checks failed"))
	} else {
		fmt.Fprintf(w, "%s\n", cFail(fmt.Sprintf("NON-CONFORMANT — %d required check(s) failed", sum.Fail)))
	}
}

// writeBlock renders one resource (or the API section): a header with counts,
// then its capabilities (implemented ✓ / not implemented ·), then its check
// details. Non-verbose shows only FAIL/WARN; verbose adds passes and evidence.
func (r *Report) writeBlock(w io.Writer, name string, results []checks.Result, caps *CollectionCapabilities, verbose bool) {
	var p, f, wn, na, sk int
	var show []checks.Result
	for _, res := range results {
		sev := r.Severity(res)
		switch sev {
		case SevPass:
			p++
		case SevFail:
			f++
		case SevWarn:
			wn++
		case SevNA:
			na++
		case SevSkip:
			sk++
		}
		if sev == SevFail || sev == SevWarn || (verbose && sev != SevNA) {
			show = append(show, res)
		}
	}
	fmt.Fprintf(w, "%s  (pass %d, fail %d, warn %d, n/a %d, skip %d)\n",
		cHead("▸ "+name), p, f, wn, na, sk)

	if caps != nil {
		r.writeResourceCapabilities(w, *caps)
	}

	r.sortByAEP(show)
	for _, res := range show {
		sev := r.Severity(res)
		fmt.Fprintf(w, "  %s  [AEP-%d %s] %s\n", sev.colored(), res.AEP, res.Level, res.CheckID)
		if res.Message != "" {
			fmt.Fprintf(w, "        %s\n", res.Message)
		}
		if sev == SevFail || verbose {
			for _, ev := range res.Evidence {
				if ev.URL != "" {
					fmt.Fprintf(w, "        %s %s → %d\n", ev.Method, ev.URL, ev.Status)
				}
			}
		}
	}
	fmt.Fprintln(w)
}

// writeResourceCapabilities prints the implemented/not-implemented picture for a
// single resource, indented under its header.
func (r *Report) writeResourceCapabilities(w io.Writer, rc CollectionCapabilities) {
	var methods []string
	for _, m := range rc.Methods {
		methods = append(methods, mark(m.Implemented)+" "+m.Name)
	}
	fmt.Fprintf(w, "      methods:  %s", joinSp(methods))
	if len(rc.Custom) > 0 {
		fmt.Fprintf(w, "   custom: %s", joinSp(rc.Custom))
	}
	fmt.Fprintln(w)
	var impl, absent []string
	for _, f := range rc.Features {
		if f.Implemented {
			impl = append(impl, f.Name)
		} else {
			absent = append(absent, f.Name)
		}
	}
	if len(impl) > 0 {
		fmt.Fprintf(w, "      features: %s\n", cPass(joinSp(impl)))
	}
	if len(absent) > 0 {
		fmt.Fprintf(w, "      not impl: %s\n", cNA(joinSp(absent)))
	}
}

func mark(on bool) string {
	if on {
		return cPass("✓")
	}
	return cNA("·")
}

func joinSp(xs []string) string { return strings.Join(xs, "  ") }
