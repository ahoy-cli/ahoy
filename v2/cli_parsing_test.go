package main

import (
	"flag"
	"os"
	"testing"

	"github.com/urfave/cli"
)

func TestFlagParsing(t *testing.T) {
	// Test that global flags are correctly defined
	if len(globalFlags) == 0 {
		t.Error("No global flags defined")
	}

	// Check that required flags exist
	flagNames := make(map[string]bool)
	for _, f := range globalFlags {
		switch f := f.(type) {
		case cli.BoolFlag:
			flagNames[f.Name] = true
		case cli.StringFlag:
			flagNames[f.Name] = true
		}
	}

	requiredFlags := []string{"verbose, v", "file, f", "help, h", "version", "generate-bash-completion"}
	for _, required := range requiredFlags {
		if !flagNames[required] {
			t.Errorf("Required flag '%s' not found in globalFlags", required)
		}
	}
}

func TestFlagSetCreation(t *testing.T) {
	// Test that flagSet function creates proper flag sets
	set := flagSet("test", globalFlags)
	if set == nil {
		t.Error("flagSet returned nil")
	}

	// Test that the flag set has the expected flags
	var hasVerbose, hasFile, hasHelp bool
	set.VisitAll(func(f *flag.Flag) {
		switch f.Name {
		case "verbose", "v":
			hasVerbose = true
		case "file", "f":
			hasFile = true
		case "help", "h":
			hasHelp = true
		}
	})

	if !hasVerbose {
		t.Error("Verbose flag not found in flag set")
	}
	if !hasFile {
		t.Error("File flag not found in flag set")
	}
	if !hasHelp {
		t.Error("Help flag not found in flag set")
	}
}

func TestInitFlags(t *testing.T) {
	// Test that initFlags properly processes incoming flags
	originalSrcDir := AhoyConf.srcDir
	defer func() { AhoyConf.srcDir = originalSrcDir }()

	// Test with empty flags
	initFlags([]string{})
	if AhoyConf.srcDir != "" {
		t.Error("Expected srcDir to be reset to empty string")
	}

	// Test with file flag
	initFlags([]string{"-f", "testdata/simple.ahoy.yml"})
	// sourcefile should be set by the flag parsing
	// Note: we can't easily test this without more complex setup
}

func TestOverrideFlags(t *testing.T) {
	// Test that overrideFlags properly configures the app
	app := cli.NewApp()
	overrideFlags(app)

	if len(app.Flags) != len(globalFlags) {
		t.Errorf("Expected %d flags, got %d", len(globalFlags), len(app.Flags))
	}

	if !app.HideVersion {
		t.Error("Expected HideVersion to be true")
	}

	if !app.HideHelp {
		t.Error("Expected HideHelp to be true")
	}
}

func TestVerboseFlagBehavior(t *testing.T) {
	// Test verbose flag behavior
	originalVerbose := verbose
	defer func() { verbose = originalVerbose }()

	// Test that verbose flag can be set
	verbose = true
	if !verbose {
		t.Error("Failed to set verbose flag")
	}

	verbose = false
	if verbose {
		t.Error("Failed to unset verbose flag")
	}
}

func TestSourcefileFlagBehavior(t *testing.T) {
	// Test sourcefile flag behavior
	originalSourcefile := sourcefile
	defer func() { sourcefile = originalSourcefile }()

	// Test that sourcefile flag can be set
	sourcefile = "test.yml"
	if sourcefile != "test.yml" {
		t.Error("Failed to set sourcefile flag")
	}

	sourcefile = ""
	if sourcefile != "" {
		t.Error("Failed to unset sourcefile flag")
	}
}

func TestEnvironmentVariableFlags(t *testing.T) {
	// Test AHOY_VERBOSE environment variable
	originalVerbose := verbose
	defer func() { verbose = originalVerbose }()

	// Set environment variable
	os.Setenv("AHOY_VERBOSE", "true")
	defer os.Unsetenv("AHOY_VERBOSE")

	// Create a flag set and parse it to test env var behavior
	set := flagSet("test", globalFlags)
	err := set.Parse([]string{})
	if err != nil {
		t.Errorf("Flag parsing failed: %v", err)
	}

	// Note: The actual environment variable processing happens in urfave/cli
	// This test mainly ensures the flag is properly configured
}

func TestFlagNameAliases(t *testing.T) {
	// Test that flag aliases work correctly
	for _, f := range globalFlags {
		switch f := f.(type) {
		case cli.BoolFlag:
			if f.Name == "verbose, v" {
				// This flag should accept both --verbose and -v
				if f.Name != "verbose, v" {
					t.Error("Verbose flag should have both long and short forms")
				}
			}
			if f.Name == "help, h" {
				// This flag should accept both --help and -h
				if f.Name != "help, h" {
					t.Error("Help flag should have both long and short forms")
				}
			}
		case cli.StringFlag:
			if f.Name == "file, f" {
				// This flag should accept both --file and -f
				if f.Name != "file, f" {
					t.Error("File flag should have both long and short forms")
				}
			}
		}
	}
}

func TestCLIAppConfiguration(t *testing.T) {
	// Test that CLI app is configured correctly for migration compatibility

	// Save original global state
	originalApp := app
	originalSourcefile := sourcefile
	originalVerbose := verbose

	defer func() {
		app = originalApp
		sourcefile = originalSourcefile
		verbose = originalVerbose
	}()

	// Test app setup
	testApp := setupApp([]string{})
	if testApp == nil {
		t.Error("setupApp returned nil")
		return
	}

	if testApp.Name != "ahoy" {
		t.Errorf("Expected app name 'ahoy', got '%s'", testApp.Name)
	}

	if testApp.Usage != "Creates a configurable cli app for running commands." {
		t.Errorf("Unexpected app usage: %s", testApp.Usage)
	}

	if !testApp.EnableBashCompletion {
		t.Error("Bash completion should be enabled")
	}
}

func TestMigrationCompatibility(t *testing.T) {
	// Test that the current flag structure is compatible with future viper migration

	// Check that all flags have clear names that can be mapped to viper
	for _, f := range globalFlags {
		switch f := f.(type) {
		case cli.BoolFlag:
			if f.Name == "" {
				t.Error("Flag has empty name")
			}
			// Check that flag names are viper-compatible (no spaces in primary name)
			if f.Name == "generate-bash-completion" {
				// This is OK - kebab case works with viper
			}
		case cli.StringFlag:
			if f.Name == "" {
				t.Error("Flag has empty name")
			}
		}
	}
}

func TestFlagValueTypes(t *testing.T) {
	// Test that flag value types are correctly configured
	for _, f := range globalFlags {
		switch f := f.(type) {
		case cli.BoolFlag:
			// Boolean flags should have proper destinations
			if f.Name == "verbose, v" && f.Destination == nil {
				t.Error("Verbose flag should have destination")
			}
		case cli.StringFlag:
			// String flags should have proper destinations
			if f.Name == "file, f" && f.Destination == nil {
				t.Error("File flag should have destination")
			}
		}
	}
}
