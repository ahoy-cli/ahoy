package main

import (
	"errors"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

type envTestConfig struct {
	Env EnvFiles `yaml:"env"`
}

func parseEnv(t *testing.T, yamlData string) EnvFiles {
	t.Helper()

	var config envTestConfig
	if err := yaml.Unmarshal([]byte(yamlData), &config); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}
	return config.Env
}

// The shorthand and list forms below predate the mapping form and must keep
// parsing exactly as they did, so every existing .ahoy.yml still works.

func TestEnvFilesUnmarshalSingleString(t *testing.T) {
	env := parseEnv(t, `env: .env`)

	if len(env) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(env))
	}
	if env[0].Path != ".env" {
		t.Errorf("Expected path '.env', got %q", env[0].Path)
	}
	if env[0].Optional {
		t.Error("A plain path must default to required")
	}
}

func TestEnvFilesUnmarshalStringArray(t *testing.T) {
	env := parseEnv(t, `
env:
  - .env.base
  - .env.local
`)

	if len(env) != 2 {
		t.Fatalf("Expected 2 elements, got %d", len(env))
	}
	for i, want := range []string{".env.base", ".env.local"} {
		if env[i].Path != want {
			t.Errorf("Element %d: expected %q, got %q", i, want, env[i].Path)
		}
		if env[i].Optional {
			t.Errorf("Element %d: a plain path must default to required", i)
		}
	}
}

func TestEnvFilesUnmarshalEmptyArray(t *testing.T) {
	if env := parseEnv(t, `env: []`); len(env) != 0 {
		t.Errorf("Expected 0 elements, got %d", len(env))
	}
}

func TestEnvFilesUnmarshalNullValue(t *testing.T) {
	if env := parseEnv(t, `env: null`); len(env) != 0 {
		t.Errorf("Expected 0 elements, got %d", len(env))
	}
}

// The mapping form, and mixtures of it with plain paths.

func TestEnvFilesUnmarshalOptionalEntry(t *testing.T) {
	env := parseEnv(t, `
env:
  - .env
  - path: .env.local
    optional: true
`)

	if len(env) != 2 {
		t.Fatalf("Expected 2 elements, got %d", len(env))
	}
	if env[0].Path != ".env" || env[0].Optional {
		t.Errorf("Expected required '.env', got %+v", env[0])
	}
	if env[1].Path != ".env.local" || !env[1].Optional {
		t.Errorf("Expected optional '.env.local', got %+v", env[1])
	}
}

func TestEnvFilesUnmarshalMappingWithoutOptionalIsRequired(t *testing.T) {
	env := parseEnv(t, `
env:
  - path: .env
`)

	if len(env) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(env))
	}
	if env[0].Optional {
		t.Error("A mapping without 'optional' must default to required")
	}
}

func TestEnvFilesUnmarshalOptionalFalseIsRequired(t *testing.T) {
	env := parseEnv(t, `
env:
  - path: .env
    optional: false
`)

	if env[0].Optional {
		t.Error("'optional: false' must mean required")
	}
}

func TestEnvFilesUnmarshalSingleMapping(t *testing.T) {
	env := parseEnv(t, `
env:
  path: .env.local
  optional: true
`)

	if len(env) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(env))
	}
	if env[0].Path != ".env.local" || !env[0].Optional {
		t.Errorf("Expected optional '.env.local', got %+v", env[0])
	}
}

// Malformed input must be rejected rather than silently producing an entry
// with an empty path, which would read as "load the config directory".

// An entry that names no file must be rejected outright. An empty path
// resolves to the config directory, which previously produced a warning about
// a missing file called ” and then carried on.
func TestEnvFilesUnmarshalEntriesWithoutAPathAreErrors(t *testing.T) {
	cases := map[string]string{
		"empty string":              `env: ""`,
		"null":                      "env:\n  - null",
		"empty string in a list":    "env:\n  - \"\"",
		"empty string before valid": "env:\n  - \"\"\n  - .env",
		"mapping with no path":      "env:\n  - optional: true",
		"mapping with empty path":   "env:\n  path: \"\"",
		"mapping with unknown key":  "env:\n  key: value",
	}

	for name, doc := range cases {
		t.Run(name, func(t *testing.T) {
			var config envTestConfig
			err := yaml.Unmarshal([]byte(doc), &config)
			if err == nil {
				t.Fatalf("Expected an error, got none (parsed as %+v)", config.Env)
			}
			if !errors.Is(err, errEnvFileNoPath) {
				t.Errorf("Expected errEnvFileNoPath, got: %v", err)
			}
		})
	}
}

func TestEnvFilesUnmarshalErrorNamesTheOffendingEntry(t *testing.T) {
	var config envTestConfig
	err := yaml.Unmarshal([]byte("env:\n  - .env\n  - null\n"), &config)
	if err == nil {
		t.Fatal("Expected an error for a null entry, got nil")
	}
	if !strings.Contains(err.Error(), "env entry 2") {
		t.Errorf("Expected the error to name entry 2, got: %v", err)
	}
}

func TestEnvFilesUnmarshalWrongShapeIsAnError(t *testing.T) {
	// A nested list is neither a path nor a mapping. The message must be ours
	// rather than the decoder's, which names the internal struct we unmarshal
	// mappings through.
	var config envTestConfig
	err := yaml.Unmarshal([]byte("env:\n  - - nested\n"), &config)
	if err == nil {
		t.Fatal("Expected an error for a nested list, got nil")
	}
	if !errors.Is(err, errEnvFileShape) {
		t.Errorf("Expected errEnvFileShape, got: %v", err)
	}
	if strings.Contains(err.Error(), "struct {") {
		t.Errorf("Error leaks the internal struct type: %v", err)
	}
}

func TestEnvFilesLen(t *testing.T) {
	env := EnvFiles{{Path: "one"}, {Path: "two"}, {Path: "three"}}
	if len(env) != 3 {
		t.Errorf("Expected length 3, got %d", len(env))
	}

	if len(EnvFiles{}) != 0 {
		t.Error("Expected empty EnvFiles to have length 0")
	}
}

// The tests below parse through the real Config and Command structs rather
// than a local stand-in, so they also cover the `yaml:"env"` wiring on both.
// Carried over from the StringArray tests this type replaced.

func TestEnvFilesInConfigStruct(t *testing.T) {
	yamlData := `ahoyapi: v2
env:
  - .env.base
  - .env.local
commands:
  test:
    usage: Test command
    cmd: echo test
    env: .env.test`

	var config Config
	if err := yaml.Unmarshal([]byte(yamlData), &config); err != nil {
		t.Fatalf("Failed to unmarshal config YAML: %v", err)
	}

	if len(config.Env) != 2 {
		t.Fatalf("Expected 2 global env files, got %d", len(config.Env))
	}
	if config.Env[0].Path != ".env.base" {
		t.Errorf("Expected '.env.base', got %q", config.Env[0].Path)
	}
	if config.Env[1].Path != ".env.local" {
		t.Errorf("Expected '.env.local', got %q", config.Env[1].Path)
	}

	testCmd, exists := config.Commands["test"]
	if !exists {
		t.Fatal("Test command not found in config")
	}
	if len(testCmd.Env) != 1 {
		t.Fatalf("Expected 1 command env file, got %d", len(testCmd.Env))
	}
	if testCmd.Env[0].Path != ".env.test" {
		t.Errorf("Expected '.env.test', got %q", testCmd.Env[0].Path)
	}
}

func TestEnvFilesBackwardsCompatibility(t *testing.T) {
	// The single-string shorthand, at both levels, exactly as configs
	// written before the mapping form existed use it.
	oldFormatYaml := `ahoyapi: v2
env: .env
commands:
  test:
    usage: Test command
    cmd: echo test
    env: .env.command`

	var config Config
	if err := yaml.Unmarshal([]byte(oldFormatYaml), &config); err != nil {
		t.Fatalf("Failed to unmarshal backwards compatible YAML: %v", err)
	}

	if len(config.Env) != 1 {
		t.Fatalf("Expected 1 global env file, got %d", len(config.Env))
	}
	if config.Env[0].Path != ".env" || config.Env[0].Optional {
		t.Errorf("Expected required '.env', got %+v", config.Env[0])
	}

	testCmd, exists := config.Commands["test"]
	if !exists {
		t.Fatal("Test command not found in config")
	}
	if len(testCmd.Env) != 1 {
		t.Fatalf("Expected 1 command env file, got %d", len(testCmd.Env))
	}
	if testCmd.Env[0].Path != ".env.command" || testCmd.Env[0].Optional {
		t.Errorf("Expected required '.env.command', got %+v", testCmd.Env[0])
	}
}

func TestEnvFilesMixedFormats(t *testing.T) {
	// All three spellings in one config: a list globally, the single-string
	// shorthand on one command, and a list mixing plain paths with the
	// mapping form on another.
	mixedYaml := `ahoyapi: v2
env:
  - .env.global1
  - .env.global2
commands:
  cmd1:
    usage: Command with single env
    cmd: echo cmd1
    env: .env.single
  cmd2:
    usage: Command with array env
    cmd: echo cmd2
    env:
      - .env.array1
      - path: .env.array2
        optional: true`

	var config Config
	if err := yaml.Unmarshal([]byte(mixedYaml), &config); err != nil {
		t.Fatalf("Failed to unmarshal mixed format YAML: %v", err)
	}

	if len(config.Env) != 2 {
		t.Errorf("Expected 2 global env files, got %d", len(config.Env))
	}

	cmd1, exists := config.Commands["cmd1"]
	if !exists {
		t.Fatal("cmd1 not found in config")
	}
	if len(cmd1.Env) != 1 {
		t.Fatalf("Expected 1 env file for cmd1, got %d", len(cmd1.Env))
	}
	if cmd1.Env[0].Path != ".env.single" {
		t.Errorf("Expected '.env.single', got %q", cmd1.Env[0].Path)
	}

	cmd2, exists := config.Commands["cmd2"]
	if !exists {
		t.Fatal("cmd2 not found in config")
	}
	if len(cmd2.Env) != 2 {
		t.Fatalf("Expected 2 env files for cmd2, got %d", len(cmd2.Env))
	}
	if cmd2.Env[0].Path != ".env.array1" || cmd2.Env[0].Optional {
		t.Errorf("Expected required '.env.array1', got %+v", cmd2.Env[0])
	}
	if cmd2.Env[1].Path != ".env.array2" || !cmd2.Env[1].Optional {
		t.Errorf("Expected optional '.env.array2', got %+v", cmd2.Env[1])
	}
}
