package cmd

import (
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/thegagne/aep-conformance-test/internal/client"
	"github.com/thegagne/aep-conformance-test/internal/discovery"
	"github.com/thegagne/aep-conformance-test/internal/report"
	"github.com/thegagne/aep-conformance-test/internal/runner"
)

// newTestCmd builds the `test` command: discover the API, run the conformance
// suite against the live server, and render a report. Exits non-zero when a
// required check fails.
func newTestCmd() *cobra.Command {
	var verbose bool
	c := &cobra.Command{
		Use:   "test [base-url]",
		Short: "Run the conformance test suite against a live API",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			base := opts.BaseURL
			if base == "" && len(args) > 0 {
				base = args[0]
			}
			model, err := loadModel(base)
			if err != nil {
				return err
			}

			// A live client is created only when a base URL is available; a
			// file-only run performs static checks and skips dynamic ones.
			var cl *client.Client
			liveURL := liveBaseURL(base)
			if liveURL != "" {
				timeout, _ := time.ParseDuration(opts.Timeout)
				if timeout == 0 {
					timeout = 30 * time.Second
				}
				cl = client.New(liveURL, timeout)
				hdrs, err := parseHeaders(opts.Headers)
				if err != nil {
					return err
				}
				cl.Headers = hdrs
			}

			selected := selectResources(model, opts.Resources)
			rn := runner.New(model, cl)
			results := rn.Run(selected)

			rep := &report.Report{
				Title:        model.Title,
				BaseURL:      liveURL,
				Strict:       opts.Strict,
				Results:      results,
				Capabilities: report.Capabilities(selected),
			}
			if err := render(cmd, rep, verbose); err != nil {
				return err
			}
			if !rep.Conformant() {
				os.Exit(1)
			}
			return nil
		},
	}
	c.Flags().StringVar(&opts.BaseURL, "base-url", "", "live server base URL (defaults to positional arg)")
	c.Flags().StringVar(&opts.Output, "output", "stdout", "output format: stdout|json|markdown")
	c.Flags().StringVar(&opts.OutputFile, "output-file", "", "write the report to a file")
	c.Flags().StringSliceVar(&opts.Resources, "resource", nil, "limit the run to named resources")
	c.Flags().BoolVar(&opts.Strict, "strict", false, "treat SHOULD failures as failures (non-zero exit)")
	c.Flags().StringVar(&opts.Timeout, "timeout", "30s", "per-request timeout")
	c.Flags().BoolVar(&verbose, "verbose", false, "show passing checks and full evidence")
	return c
}

// liveBaseURL returns the HTTP(S) base URL to test against, or "" if only a
// spec file was provided (static-only run).
func liveBaseURL(base string) string {
	if isHTTP(base) {
		return base
	}
	return ""
}

func isHTTP(s string) bool {
	return len(s) >= 7 && (s[:7] == "http://" || (len(s) >= 8 && s[:8] == "https://"))
}

func selectResources(m *discovery.APIModel, names []string) []*discovery.Resource {
	if len(names) == 0 {
		return m.Resources
	}
	want := map[string]bool{}
	for _, n := range names {
		want[n] = true
	}
	var out []*discovery.Resource
	for _, r := range m.Resources {
		if want[r.Singular] || want[r.Plural] {
			out = append(out, r)
		}
	}
	return out
}

func render(cmd *cobra.Command, rep *report.Report, verbose bool) error {
	out := cmd.OutOrStdout()
	if opts.OutputFile != "" {
		f, err := os.Create(opts.OutputFile)
		if err != nil {
			return err
		}
		defer f.Close()
		out = f
	}
	switch opts.Output {
	case "json":
		return rep.WriteJSON(out)
	case "markdown", "md":
		rep.WriteMarkdown(out)
	default:
		rep.WriteStdout(out, verbose)
	}
	return nil
}
