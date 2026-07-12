package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/thegagne/aep-conformance-test/internal/buildinfo"
)

// newVersionCmd prints the tool version and the AEP spec revision the checks
// target. The same information is also available via the root --version flag.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version and the AEP spec revision the checks target",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintln(cmd.OutOrStdout(), buildinfo.String())
		},
	}
}
