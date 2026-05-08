package main

// flag.go - backwards-compatibility pre-parser for legacy single-dash long flags.
//
// Background
// ----------
// Ahoy was originally built on urfave/cli, which accepted single-dash long
// flags such as `-version`, `-help`, `-verbose`, `-file foo.yml`. The
// cobra/pflag stack used from v3 onwards does NOT support that form -
// pflag interprets `-version` as a cluster of single-character short
// flags (`-v -e -r -s -i -o -n`), which fails parsing.
//
// Plenty of existing scripts, CI pipelines, and shell aliases still use
// the single-dash form, so dropping support would be a silent breakage
// for users upgrading from v2. To keep them working we pre-parse the
// incoming arguments here, before cobra ever sees them, using the stdlib
// `flag` package - which natively understands single-dash long names.
//
// The results are written to package-level globals consumed by main()
// and setupApp():
//
//   sourcefile             value of -f / -file / --file
//   verbose                value of -v / -verbose / --verbose
//   simulateVersion        value of --simulate-version (test-only flag)
//   versionFlagSet         true if -version / --version was seen
//   helpFlagSet            true if -h / -help / --help was seen
//   bashCompletionFlagSet  true if --generate-bash-completion was seen
//   invalidFlagError       non-empty if stdlib parsing failed
//
// Cobra also defines the same flags via PersistentFlags so they appear in
// the help output and are parsed normally for the double-dash form when it
// follows a subcommand (e.g. `ahoy mycmd --verbose`). The pre-parser only
// catches flags ahead of the subcommand position.
//
// Environment-variable fallbacks (AHOY_FILE, AHOY_VERBOSE) are applied
// here too, after flag parsing, so an explicit flag always wins over the
// env var.
//
// If single-dash compatibility is ever dropped, this whole file can be
// deleted, the early-exit blocks in main() can go, and `setupApp` can
// stop saving/restoring parsed defaults around the cobra flag setup.

import (
	"bytes"
	"flag"
	"os"
	"strings"
)

var (
	versionFlagSet        bool
	helpFlagSet           bool
	bashCompletionFlagSet bool
	invalidFlagError      string
)

// initFlags pre-parses the incoming arguments to honour legacy single-dash
// long flags before cobra runs. See the file-level doc comment above for
// the full rationale.
func initFlags(incomingFlags []string) {
	resetFlagState()

	normalisedFlags := normaliseLongFlagPrefixes(incomingFlags)

	// Local sinks for flags we only need as on/off signals; copied into
	// package globals after parsing so the FlagSet's pointer bindings
	// remain valid for the duration of Parse().
	var versionFlag, helpFlag, bashCompletionFlag bool

	fs, errBuf := newLegacyFlagSet(&versionFlag, &helpFlag, &bashCompletionFlag)

	if err := fs.Parse(normalisedFlags); err != nil {
		invalidFlagError = errBuf.String()
	}

	versionFlagSet = versionFlag
	helpFlagSet = helpFlag
	bashCompletionFlagSet = bashCompletionFlag

	applyEnvFallbacks()
}

// resetFlagState clears all pre-parser globals. Required because tests
// reuse the package-level state between runs.
func resetFlagState() {
	AhoyConf.srcDir = ""
	versionFlagSet = false
	helpFlagSet = false
	bashCompletionFlagSet = false
	invalidFlagError = ""
}

// normaliseLongFlagPrefixes rewrites `--foo` to `-foo` so the stdlib flag
// package - which only understands single-dash - can parse both forms in
// a single pass. This is the inverse of what we'd like (we'd rather rewrite
// `-foo` to `--foo` and feed it to pflag), but stdlib flag is the only
// parser that natively accepts single-dash long names without ambiguity.
func normaliseLongFlagPrefixes(args []string) []string {
	out := make([]string, len(args))
	for i, arg := range args {
		if strings.HasPrefix(arg, "--") {
			out[i] = "-" + strings.TrimPrefix(arg, "--")
		} else {
			out[i] = arg
		}
	}
	return out
}

// newLegacyFlagSet builds the stdlib FlagSet used for pre-parsing. The
// caller passes pointers for the on/off flags so that the FlagSet's
// internal pointer bindings remain valid for the duration of Parse().
// The returned errBuf captures any parse-error text for later replay.
func newLegacyFlagSet(versionFlag, helpFlag, bashCompletionFlag *bool) (*flag.FlagSet, *bytes.Buffer) {
	fs := flag.NewFlagSet("ahoyLegacyFlags", flag.ContinueOnError)
	errBuf := &bytes.Buffer{}
	fs.SetOutput(errBuf)

	// Flags whose values flow into package globals.
	fs.StringVar(&sourcefile, "f", "", "specify the sourcefile")
	fs.StringVar(&sourcefile, "file", "", "specify the sourcefile")
	fs.BoolVar(&verbose, "v", false, "verbose output")
	fs.BoolVar(&verbose, "verbose", false, "verbose output")
	fs.StringVar(&simulateVersion, "simulate-version", "", "")

	// Flags we only need to detect; cobra also defines them but we exit
	// early in main() if they were given in the legacy single-dash form.
	fs.BoolVar(versionFlag, "version", false, "print version")
	fs.BoolVar(helpFlag, "help", false, "print help")
	fs.BoolVar(helpFlag, "h", false, "print help")
	fs.BoolVar(bashCompletionFlag, "generate-bash-completion", false, "")

	return fs, errBuf
}

// applyEnvFallbacks fills in sourcefile / verbose from AHOY_FILE /
// AHOY_VERBOSE when the equivalent flag was not given. Explicit flags
// always take precedence.
func applyEnvFallbacks() {
	if sourcefile == "" {
		if v := os.Getenv("AHOY_FILE"); v != "" {
			sourcefile = v
		}
	}
	if !verbose && os.Getenv("AHOY_VERBOSE") == "true" {
		verbose = true
	}
}
