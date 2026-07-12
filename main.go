// Command aep-conformance runs AEP.dev conformance tests against a live API.
package main

import (
	"os"

	"github.com/thegagne/aep-conformance-test/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
