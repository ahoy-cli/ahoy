package main

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
)

func TestFlagParsing(t *testing.T) {
	cmd := newAppState().setupApp([]string{})
	if cmd == nil {
		t.Error("setupApp returned nil")
		return
	}

	requiredFlags := map[string]bool{
		"verbose": false,
		"file":    false,
		"help":    false,
		"version": false,
	}

	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if _, ok := requiredFlags[f.Name]; ok {
			requiredFlags[f.Name] = true
		}
	})

	for name, found := range requiredFlags {
		if !found {
			t.Errorf("Required flag '%s' not found", name)
		}
	}
}

func TestInitFlags(t *testing.T) {
	// Test with empty flags - srcDir should be reset to empty string.
	s := newAppState()
	s.srcDir = "/some/stale/path"
	s.initFlags([]string{})
	if s.srcDir != "" {
		t.Error("Expected srcDir to be reset to empty string")
	}

	// Test that -f sets sourcefile.
	s2 := newAppState()
	s2.initFlags([]string{"-f", "testdata/simple.ahoy.yml"})
	if s2.sourcefile != "testdata/simple.ahoy.yml" {
		t.Errorf("Expected sourcefile to be 'testdata/simple.ahoy.yml', got '%s'", s2.sourcefile)
	}
}

func TestVerboseFlagBehavior(t *testing.T) {
	s := newAppState()
	s.verbose = true
	if !s.verbose {
		t.Error("Failed to set verbose flag")
	}

	s.verbose = false
	if s.verbose {
		t.Error("Failed to unset verbose flag")
	}
}

func TestSourcefileFlagBehavior(t *testing.T) {
	s := newAppState()
	s.sourcefile = "test.yml"
	if s.sourcefile != "test.yml" {
		t.Error("Failed to set sourcefile flag")
	}

	s.sourcefile = ""
	if s.sourcefile != "" {
		t.Error("Failed to unset sourcefile flag")
	}
}

func TestEnvironmentVariableFlags(t *testing.T) {
	defer func() {
		os.Unsetenv("AHOY_VERBOSE")
		os.Unsetenv("AHOY_FILE")
	}()

	t.Run("AHOY_VERBOSE sets verbose when no flag given", func(t *testing.T) {
		os.Setenv("AHOY_VERBOSE", "true")
		s := newAppState()
		s.initFlags([]string{})
		if !s.verbose {
			t.Error("Expected verbose to be true via AHOY_VERBOSE env var.")
		}
	})

	t.Run("explicit -v flag takes precedence over AHOY_VERBOSE=false", func(t *testing.T) {
		os.Unsetenv("AHOY_VERBOSE")
		s := newAppState()
		s.initFlags([]string{"-v"})
		if !s.verbose {
			t.Error("Expected verbose to be true via -v flag.")
		}
	})

	t.Run("AHOY_FILE sets sourcefile when no flag given", func(t *testing.T) {
		os.Setenv("AHOY_FILE", "custom.ahoy.yml")
		s := newAppState()
		s.initFlags([]string{})
		if s.sourcefile != "custom.ahoy.yml" {
			t.Errorf("Expected sourcefile 'custom.ahoy.yml', got '%s'.", s.sourcefile)
		}
	})

	t.Run("explicit -f flag takes precedence over AHOY_FILE", func(t *testing.T) {
		os.Setenv("AHOY_FILE", "env.ahoy.yml")
		s := newAppState()
		s.initFlags([]string{"-f", "explicit.ahoy.yml"})
		if s.sourcefile != "explicit.ahoy.yml" {
			t.Errorf("Expected sourcefile 'explicit.ahoy.yml' from flag, got '%s'.", s.sourcefile)
		}
	})
}

func TestFlagNameAliases(t *testing.T) {
	cmd := newAppState().setupApp([]string{})
	if cmd == nil {
		t.Error("setupApp returned nil")
		return
	}

	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("Verbose flag not found")
	} else if verboseFlag.Shorthand != "v" {
		t.Errorf("Expected verbose flag shorthand 'v', got '%s'", verboseFlag.Shorthand)
	}

	fileFlag := cmd.PersistentFlags().Lookup("file")
	if fileFlag == nil {
		t.Error("File flag not found")
	} else if fileFlag.Shorthand != "f" {
		t.Errorf("Expected file flag shorthand 'f', got '%s'", fileFlag.Shorthand)
	}
}

func TestCLIAppConfiguration(t *testing.T) {
	testCmd := newAppState().setupApp([]string{})
	if testCmd == nil {
		t.Error("setupApp returned nil")
		return
	}

	if testCmd.Use != "ahoy" {
		t.Errorf("Expected command name 'ahoy', got '%s'", testCmd.Use)
	}

	if testCmd.Short != "Creates a configurable cli app for running commands." {
		t.Errorf("Unexpected command description: %s", testCmd.Short)
	}

	if testCmd.ValidArgsFunction == nil {
		t.Error("Bash completion function should be set")
	}
}

func TestMigrationCompatibility(t *testing.T) {
	cmd := newAppState().setupApp([]string{})
	if cmd == nil {
		t.Error("setupApp returned nil")
		return
	}

	for _, name := range []string{"verbose", "file"} {
		if cmd.PersistentFlags().Lookup(name) == nil {
			t.Errorf("Expected persistent flag '%s' to be registered on root command.", name)
		}
	}
}

func TestFlagValueTypes(t *testing.T) {
	cmd := newAppState().setupApp([]string{})
	if cmd == nil {
		t.Error("setupApp returned nil")
		return
	}

	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("Verbose flag not found")
	} else if verboseFlag.Value.Type() != "bool" {
		t.Errorf("Expected verbose flag type 'bool', got '%s'", verboseFlag.Value.Type())
	}

	fileFlag := cmd.PersistentFlags().Lookup("file")
	if fileFlag == nil {
		t.Error("File flag not found")
	} else if fileFlag.Value.Type() != "string" {
		t.Errorf("Expected file flag type 'string', got '%s'", fileFlag.Value.Type())
	}
}
