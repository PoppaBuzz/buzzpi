// Package version provides the BuzzPi Runtime version information.
// The version is set at build time via -ldflags.
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

var (
	// Version is the semantic version of the Runtime.
	// Overridden at build time: -ldflags="-X github.com/buzzpi/agent/internal/version.Version=v0.1.0"
	Version = "v0.0.0-dev"

	// Commit is the git commit hash.
	Commit = "unknown"

	// Date is the build date.
	Date = "unknown"

	// MinRuntimeVersion is the minimum Runtime version for compatibility checks.
	MinRuntimeVersion = "0.1.0"

	// BPPVersion is the supported BPP protocol version.
	BPPVersion = "1"

	// UserAgent is the HTTP User-Agent header value.
	UserAgent = "BuzzPi-Runtime/" + Version
)

// Info returns a formatted version string.
func Info() string {
	return fmt.Sprintf("BuzzPi Runtime %s (commit: %s, built: %s, %s/%s)",
		Version, Commit, Date, runtime.GOOS, runtime.GOARCH)
}

// GoVersion returns the Go runtime version used to build this binary.
func GoVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return runtime.Version()
	}
	return bi.GoVersion
}

// ModuleVersion returns the version of a Go module dependency.
