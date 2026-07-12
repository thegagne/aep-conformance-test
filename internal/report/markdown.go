package report

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/thegagne/aep-conformance-test/internal/checks"
)

// WriteMarkdown renders the report as a markdown document with human-friendly,
// width-aligned tables. Each collection gets a capabilities table (what it
// implements) followed by its checks, grouped so failures and warnings are
// prominent and not-applicable checks are still accounted for.
func (r *Report) WriteMarkdown(w io.Writer) {
	sum := r.Summary()
	fmt.Fprintf(w, "# AEP Conformance Report\n\n")
	if r.Title != "" {
		fmt.Fprintf(w, "**API:** %s  \n", r.Title)
	}
	if r.BaseURL != "" {
		fmt.Fprintf(w, "**Target:** %s  \n", r.BaseURL)
	}
	if r.ToolVersion != "" {
		fmt.Fprintf(w, "**Tool:** aep-conformance %s  \n", r.ToolVersion)
	}
	if r.SpecRevision != "" {
		fmt.Fprintf(w, "**Spec:** %s  \n", r.SpecRevision)
	}
	if r.GeneratedAt != "" {
		fmt.Fprintf(w, "**Generated:** %s  \n", r.GeneratedAt)
	}
	verdict := "✅ **CONFORMANT**"
	if !r.Conformant() {
		verdict = fmt.Sprintf("❌ **NON-CONFORMANT** — %d required check(s) failed", sum.Fail)
	}
	fmt.Fprintf(w, "\n%s\n\n", verdict)

	writeTable(w, []string{"pass", "fail", "warn", "n/a", "skip"}, "rrrrr", [][]string{
		{strconv.Itoa(sum.Pass), strconv.Itoa(sum.Fail), strconv.Itoa(sum.Warn), strconv.Itoa(sum.NA), strconv.Itoa(sum.Skip)},
	})
	fmt.Fprintln(w)

	fmt.Fprintf(w, "**Legend**\n\n")
	fmt.Fprintf(w, "- ✅ **PASS** — conforms.\n")
	fmt.Fprintf(w, "- ❌ **FAIL** — a required (MUST) rule is violated; breaks conformance.\n")
	fmt.Fprintf(w, "- ⚠️ **WARN** — a recommended (SHOULD) rule is violated; allowed, but worth fixing.\n")
	fmt.Fprintf(w, "- ➖ **N/A** — targets an optional feature this API doesn't implement; not tested.\n")
	fmt.Fprintf(w, "- ⏭️ **SKIP** — could not be evaluated (e.g. a prerequisite step failed).\n\n")

	api, groups := r.Grouped()
	if len(api) > 0 {
		fmt.Fprintf(w, "## API\n\n")
		r.markdownChecks(w, api)
	}
	for _, g := range groups {
		fmt.Fprintf(w, "## %s\n\n", g.Collection)
		if g.Capabilities != nil {
			r.markdownCapabilities(w, *g.Capabilities)
		}
		r.markdownChecks(w, g.Results)
	}
}

// markdownCapabilities renders a collection's capabilities: one row per
// capability with an implemented column.
func (r *Report) markdownCapabilities(w io.Writer, rc CollectionCapabilities) {
	fmt.Fprintf(w, "**Capabilities**\n\n")
	rows := [][]string{}
	for _, m := range rc.Methods {
		rows = append(rows, []string{m.Name, "method", yesNo(m.Implemented)})
	}
	for _, c := range rc.Custom {
		rows = append(rows, []string{c, "custom method", "✅"})
	}
	for _, f := range rc.Features {
		rows = append(rows, []string{f.Name, "feature", yesNo(f.Implemented)})
	}
	writeTable(w, []string{"capability", "kind", "implemented"}, "llc", rows)
	fmt.Fprintln(w)
}

// markdownChecks renders check results in AEP order (spec order), with failures
// surfacing first within each AEP so it's clear what wasn't tested and why.
func (r *Report) markdownChecks(w io.Writer, results []checks.Result) {
	sorted := make([]checks.Result, len(results))
	copy(sorted, results)
	r.sortByAEP(sorted)

	fmt.Fprintf(w, "**Checks**\n\n")
	rows := make([][]string, 0, len(sorted))
	for _, res := range sorted {
		sev := r.Severity(res)
		detail := res.Message
		if sev == SevNA && detail == "" {
			detail = "optional feature not implemented — not tested"
		}
		rows = append(rows, []string{
			r.Severity(res).String(),
			"AEP-" + strconv.Itoa(res.AEP),
			res.Level.String(),
			"`" + res.CheckID + "`",
			mdEscape(detail),
		})
	}
	writeTable(w, []string{"status", "AEP", "level", "check", "detail"}, "llllll", rows)
	fmt.Fprintln(w)
}

// writeTable renders a GitHub-flavored markdown table with cells padded to each
// column's width so the raw text is aligned and readable. align is one rune per
// column: 'l' left, 'r' right, 'c' center.
func writeTable(w io.Writer, headers []string, align string, rows [][]string) {
	cols := len(headers)
	width := make([]int, cols)
	for i, h := range headers {
		width[i] = runeLen(h)
	}
	for _, row := range rows {
		for i := 0; i < cols && i < len(row); i++ {
			if l := runeLen(row[i]); l > width[i] {
				width[i] = l
			}
		}
	}
	al := func(i int) byte {
		if i < len(align) {
			return align[i]
		}
		return 'l'
	}

	writeRow := func(cells []string) {
		var b strings.Builder
		b.WriteString("|")
		for i := 0; i < cols; i++ {
			cell := ""
			if i < len(cells) {
				cell = cells[i]
			}
			b.WriteString(" " + pad(cell, width[i], al(i)) + " |")
		}
		fmt.Fprintln(w, b.String())
	}

	writeRow(headers)
	var b strings.Builder
	b.WriteString("|")
	for i := 0; i < cols; i++ {
		b.WriteString(" " + sep(width[i], al(i)) + " |")
	}
	fmt.Fprintln(w, b.String())
	for _, row := range rows {
		writeRow(row)
	}
}

func pad(s string, w int, align byte) string {
	n := runeLen(s)
	if n >= w {
		return s
	}
	gap := w - n
	switch align {
	case 'r':
		return strings.Repeat(" ", gap) + s
	case 'c':
		l := gap / 2
		return strings.Repeat(" ", l) + s + strings.Repeat(" ", gap-l)
	default:
		return s + strings.Repeat(" ", gap)
	}
}

func sep(w int, align byte) string {
	if w < 3 {
		w = 3
	}
	switch align {
	case 'r':
		return strings.Repeat("-", w-1) + ":"
	case 'c':
		return ":" + strings.Repeat("-", w-2) + ":"
	default:
		return ":" + strings.Repeat("-", w-1)
	}
}

func runeLen(s string) int { return len([]rune(s)) }

func yesNo(b bool) string {
	if b {
		return "✅"
	}
	return "❌"
}

func mdEscape(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '|' || r == '\n' {
			out = append(out, ' ')
			continue
		}
		out = append(out, r)
	}
	return string(out)
}
