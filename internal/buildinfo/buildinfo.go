// Package buildinfo exposes the tool's version and the AEP spec revision its
// check catalog targets, so both the CLI (`version`) and every emitted report
// are self-describing and auditable.
package buildinfo

import (
	"fmt"
	"runtime/debug"
)

// Version, Commit, and Date are overridden at build time via -ldflags
// (see .github/workflows and any release tooling). They default to values
// suitable for `go run`/`go build` without ldflags.
var (
	Version = "dev"
	Commit  = ""
	Date    = ""
)

// AEPSpecRevision identifies the aep.dev specification snapshot that the current
// check catalog encodes. It is source-defined (not build-time): bump it whenever
// checks are added or changed to track a newer spec revision. A conformance
// verdict is only meaningful relative to this value, so it is stamped into
// reports.
const AEPSpecRevision = "aep.dev catalog @ 2026-07-11"

// Info returns the resolved build metadata, filling any gaps left by missing
// ldflags from the Go module's embedded VCS stamp (present in `go install`ed
// binaries).
func Info() (version, commit, date string) {
	version, commit, date = Version, Commit, Date
	if version != "dev" && commit != "" && date != "" {
		return version, commit, date
	}
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return version, commit, date
	}
	if version == "dev" && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		version = bi.Main.Version
	}
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			if commit == "" {
				commit = s.Value
			}
		case "vcs.time":
			if date == "" {
				date = s.Value
			}
		}
	}
	return version, commit, date
}

// String renders a multi-line version summary including the spec revision.
func String() string {
	v, c, d := Info()
	line := "aep-conformance " + v
	if c != "" {
		if len(c) > 12 {
			c = c[:12]
		}
		line += " (" + c + ")"
	}
	if d != "" {
		line += " built " + d
	}
	return fmt.Sprintf("%s\nspec: %s", line, AEPSpecRevision)
}
