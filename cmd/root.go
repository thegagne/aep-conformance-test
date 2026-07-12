// Package cmd implements the aep-conformance command-line interface.
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

// Options holds the flags shared across subcommands.
type Options struct {
	OpenAPI    string
	BaseURL    string
	Config     string
	Output     string
	OutputFile string
	Resources  []string
	Strict     bool
	Timeout    string
	// Headers are raw -H values ("Key: Value" or "Key=Value") applied to the
	// spec fetch and every request against the live server.
	Headers []string
}

var opts Options

// NewRootCmd builds the root cobra command.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "aep-conformance",
		Short: "Run AEP.dev conformance tests against a live API",
		Long: "aep-conformance discovers an API from its AEP-annotated OpenAPI document\n" +
			"and runs conformance tests against the live server, reporting where it\n" +
			"conforms to the AEP.dev specifications.",
		SilenceUsage: true,
	}
	root.PersistentFlags().StringVar(&opts.OpenAPI, "openapi", "", "OpenAPI spec source: local file path or URL (default <base-url>/openapi.json)")
	root.PersistentFlags().StringVar(&opts.Config, "config", "", "override config file (auth, sample values)")
	root.PersistentFlags().StringArrayVarP(&opts.Headers, "header", "H", nil, "extra request header 'Key: Value' or 'Key=Value' (repeatable); applied to the spec fetch and every request")
	root.AddCommand(newDiscoverCmd())
	root.AddCommand(newTestCmd())
	return root
}

// loadModel resolves the OpenAPI source and parses it into an APIModel. The
// source is, in order of precedence: an explicit --openapi flag; a non-HTTP
// positional argument treated as a spec file; else <base-url>/openapi.json.
func loadModel(baseURL string) (*discovery.APIModel, error) {
	hdrs, err := parseHeaders(opts.Headers)
	if err != nil {
		return nil, err
	}
	if opts.OpenAPI != "" {
		return discovery.Load(opts.OpenAPI, hdrs)
	}
	if baseURL == "" {
		return nil, fmt.Errorf("provide a base URL or --openapi spec source")
	}
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		return discovery.Load(baseURL, hdrs) // positional is a spec file path
	}
	return discovery.LoadFromBaseURL(baseURL, hdrs)
}

// parseHeaders turns raw -H flag values into a header map. Each value is split
// on the first ':' or '=', whichever comes first, so "Authorization: Bearer x"
// and "X-Api-Key=secret" both work. Returns nil when no headers are given.
func parseHeaders(raw []string) (map[string]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	out := make(map[string]string, len(raw))
	for _, h := range raw {
		i := strings.IndexAny(h, ":=")
		if i <= 0 {
			return nil, fmt.Errorf("invalid header %q: expected 'Key: Value' or 'Key=Value'", h)
		}
		k := strings.TrimSpace(h[:i])
		v := strings.TrimSpace(h[i+1:])
		if k == "" {
			return nil, fmt.Errorf("invalid header %q: empty key", h)
		}
		out[k] = v
	}
	return out, nil
}

func newDiscoverCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "discover [base-url]",
		Short: "Discover and print the API's resource hierarchy and detected features",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			base := ""
			if len(args) > 0 {
				base = args[0]
			}
			m, err := loadModel(base)
			if err != nil {
				return err
			}
			printModel(cmd, m)
			return nil
		},
	}
}

func printModel(cmd *cobra.Command, m *discovery.APIModel) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "OpenAPI %s  %s\n", m.OpenAPIVersion, m.Title)
	if m.APIName != "" {
		fmt.Fprintf(out, "API name: %s\n", m.APIName)
	}
	fmt.Fprintf(out, "%d resources:\n", len(m.Resources))
	for _, r := range m.Resources {
		indent := strings.Repeat("  ", r.Depth+1)
		var methods []string
		for _, mth := range []discovery.Method{
			discovery.MethodGet, discovery.MethodList, discovery.MethodCreate,
			discovery.MethodUpdate, discovery.MethodApply, discovery.MethodDelete,
		} {
			if r.Method(mth) != nil {
				methods = append(methods, string(mth))
			}
		}
		for _, c := range r.Custom {
			methods = append(methods, ":"+c.CustomVerb)
		}
		fmt.Fprintf(out, "%s%s (%s)  [%s]\n", indent, r.Singular, r.Plural, strings.Join(methods, " "))
		if feats := featureList(r); len(feats) > 0 {
			fmt.Fprintf(out, "%s  features: %s\n", indent, strings.Join(feats, ", "))
		}
	}
}

func featureList(r *discovery.Resource) []string {
	f := r.Features
	var out []string
	add := func(on bool, name string) {
		if on {
			out = append(out, name)
		}
	}
	add(f.Pagination, "pagination")
	add(f.Skip, "skip")
	add(f.Filter, "filter")
	add(f.OrderBy, "order_by")
	add(f.ReadMask, "read_mask")
	add(f.FieldMask, "field_mask")
	add(f.ShowDeleted, "show_deleted")
	add(f.Force, "force")
	add(f.IdempotencyKey, "idempotency_key")
	add(f.Revisions, "revisions")
	add(f.ExpireTime, "expire_time")
	add(f.Unreachable, "unreachable")
	add(f.UserSettableID, "user_settable_id")
	add(f.AllowMissing, "allow_missing")
	add(f.OverwriteDelete, "overwrite_soft_deleted")
	return out
}
