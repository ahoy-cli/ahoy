package main

import (
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

func TestEnvFilesUnmarshalMappingWithoutPathIsAnError(t *testing.T) {
	var config envTestConfig
	err := yaml.Unmarshal([]byte("env:\n  - optional: true\n"), &config)
	if err == nil {
		t.Fatal("Expected an error for a mapping with no 'path' key, got nil")
	}
}

func TestEnvFilesUnmarshalUnknownMappingKeyIsAnError(t *testing.T) {
	// Previously `env: {key: value}` was a type error. It must stay an error
	// now that mappings are meaningful, because it still names no file.
	var config envTestConfig
	err := yaml.Unmarshal([]byte("env:\n  key: value\n"), &config)
	if err == nil {
		t.Fatal("Expected an error for a mapping with no 'path' key, got nil")
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
