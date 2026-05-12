package main

import (
	"testing"

	"gopkg.in/yaml.v2"
)

func TestStringArrayUnmarshalSingleString(t *testing.T) {
	// Test unmarshalling a single string value
	yamlData := `env: .env`

	type TestConfig struct {
		Env StringArray `yaml:"env"`
	}

	var config TestConfig
	err := yaml.Unmarshal([]byte(yamlData), &config)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	if len(config.Env) != 1 {
		t.Errorf("Expected 1 element, got %d", len(config.Env))
	}

	if config.Env[0] != ".env" {
		t.Errorf("Expected '.env', got '%s'", config.Env[0])
	}
}

func TestStringArrayUnmarshalStringArray(t *testing.T) {
	// Test unmarshalling an array of strings
	yamlData := `env:
  - .env.base
  - .env.local
  - .env.override`

	type TestConfig struct {
		Env StringArray `yaml:"env"`
	}

	var config TestConfig
	err := yaml.Unmarshal([]byte(yamlData), &config)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	expected := []string{".env.base", ".env.local", ".env.override"}
	if len(config.Env) != len(expected) {
		t.Errorf("Expected %d elements, got %d", len(expected), len(config.Env))
	}

	for i, expectedValue := range expected {
		if config.Env[i] != expectedValue {
			t.Errorf("Expected '%s' at index %d, got '%s'", expectedValue, i, config.Env[i])
		}
	}
}

func TestStringArrayUnmarshalEmptyArray(t *testing.T) {
	// Test unmarshalling an empty array
	yamlData := `env: []`

	type TestConfig struct {
		Env StringArray `yaml:"env"`
	}

	var config TestConfig
	err := yaml.Unmarshal([]byte(yamlData), &config)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	if len(config.Env) != 0 {
		t.Errorf("Expected 0 elements, got %d", len(config.Env))
	}
}

func TestStringArrayUnmarshalNullValue(t *testing.T) {
	// Test unmarshalling a null value
	yamlData := `env: null`

	type TestConfig struct {
		Env StringArray `yaml:"env"`
	}

	var config TestConfig
	err := yaml.Unmarshal([]byte(yamlData), &config)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	if len(config.Env) != 0 {
		t.Errorf("Expected 0 elements, got %d", len(config.Env))
	}
}

func TestStringArrayUnmarshalInvalidType(t *testing.T) {
	// Test unmarshalling a complex invalid type (should return error)
	yamlData := `env:
  key: value`

	type TestConfig struct {
		Env StringArray `yaml:"env"`
	}

	var config TestConfig
	err := yaml.Unmarshal([]byte(yamlData), &config)
	if err == nil {
		t.Error("Expected error when unmarshalling invalid type, got nil")
	}
}

func TestStringArrayLen(t *testing.T) {
	// Test that len() works correctly with StringArray
	sa := StringArray{"one", "two", "three"}
	if len(sa) != 3 {
		t.Errorf("Expected length 3, got %d", len(sa))
	}

	empty := StringArray{}
	if len(empty) != 0 {
		t.Errorf("Expected length 0, got %d", len(empty))
	}
}

func TestStringArrayIndexing(t *testing.T) {
	// Test that indexing works correctly with StringArray
	sa := StringArray{"first", "second", "third"}

	if sa[0] != "first" {
		t.Errorf("Expected 'first', got '%s'", sa[0])
	}

	if sa[1] != "second" {
		t.Errorf("Expected 'second', got '%s'", sa[1])
	}

	if sa[2] != "third" {
		t.Errorf("Expected 'third', got '%s'", sa[2])
	}
}

func TestStringArrayIteration(t *testing.T) {
	// Test that range iteration works correctly with StringArray
	sa := StringArray{"alpha", "beta", "gamma"}
	expected := []string{"alpha", "beta", "gamma"}

	var result []string
	for _, value := range sa {
		result = append(result, value)
	}

	if len(result) != len(expected) {
		t.Errorf("Expected %d items, got %d", len(expected), len(result))
	}

	for i, expectedValue := range expected {
		if result[i] != expectedValue {
			t.Errorf("Expected '%s' at index %d, got '%s'", expectedValue, i, result[i])
		}
	}
}

func TestStringArrayAppend(t *testing.T) {
	// Test that append works correctly with StringArray
	sa := StringArray{"one"}
	sa = append(sa, "two", "three")

	expected := []string{"one", "two", "three"}
	if len(sa) != len(expected) {
		t.Errorf("Expected %d elements, got %d", len(expected), len(sa))
	}

	for i, expectedValue := range expected {
		if sa[i] != expectedValue {
			t.Errorf("Expected '%s' at index %d, got '%s'", expectedValue, i, sa[i])
		}
	}
}

func TestStringArrayInConfigStruct(t *testing.T) {
	// Test that StringArray works correctly in Config struct (like in actual usage)
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
	err := yaml.Unmarshal([]byte(yamlData), &config)
	if err != nil {
		t.Fatalf("Failed to unmarshal config YAML: %v", err)
	}

	// Test global env
	if len(config.Env) != 2 {
		t.Errorf("Expected 2 global env files, got %d", len(config.Env))
	}

	if config.Env[0] != ".env.base" {
		t.Errorf("Expected '.env.base', got '%s'", config.Env[0])
	}

	if config.Env[1] != ".env.local" {
		t.Errorf("Expected '.env.local', got '%s'", config.Env[1])
	}

	// Test command env
	testCmd, exists := config.Commands["test"]
	if !exists {
		t.Fatal("Test command not found in config")
	}

	if len(testCmd.Env) != 1 {
		t.Errorf("Expected 1 command env file, got %d", len(testCmd.Env))
	}

	if testCmd.Env[0] != ".env.test" {
		t.Errorf("Expected '.env.test', got '%s'", testCmd.Env[0])
	}
}

func TestStringArrayBackwardsCompatibility(t *testing.T) {
	// Test that existing single string configs still work
	oldFormatYaml := `ahoyapi: v2
env: .env
commands:
  test:
    usage: Test command
    cmd: echo test
    env: .env.command`

	var config Config
	err := yaml.Unmarshal([]byte(oldFormatYaml), &config)
	if err != nil {
		t.Fatalf("Failed to unmarshal backwards compatible YAML: %v", err)
	}

	// Test global env backwards compatibility
	if len(config.Env) != 1 {
		t.Errorf("Expected 1 global env file for backwards compatibility, got %d", len(config.Env))
	}

	if config.Env[0] != ".env" {
		t.Errorf("Expected '.env', got '%s'", config.Env[0])
	}

	// Test command env backwards compatibility
	testCmd, exists := config.Commands["test"]
	if !exists {
		t.Fatal("Test command not found in config")
	}

	if len(testCmd.Env) != 1 {
		t.Errorf("Expected 1 command env file for backwards compatibility, got %d", len(testCmd.Env))
	}

	if testCmd.Env[0] != ".env.command" {
		t.Errorf("Expected '.env.command', got '%s'", testCmd.Env[0])
	}
}

func TestStringArrayMixedFormats(t *testing.T) {
	// Test mixing single string and array formats in the same config
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
      - .env.array2`

	var config Config
	err := yaml.Unmarshal([]byte(mixedYaml), &config)
	if err != nil {
		t.Fatalf("Failed to unmarshal mixed format YAML: %v", err)
	}

	// Test global env (array format)
	if len(config.Env) != 2 {
		t.Errorf("Expected 2 global env files, got %d", len(config.Env))
	}

	// Test cmd1 (single string format)
	cmd1, exists := config.Commands["cmd1"]
	if !exists {
		t.Fatal("cmd1 not found in config")
	}

	if len(cmd1.Env) != 1 {
		t.Errorf("Expected 1 env file for cmd1, got %d", len(cmd1.Env))
	}

	if cmd1.Env[0] != ".env.single" {
		t.Errorf("Expected '.env.single', got '%s'", cmd1.Env[0])
	}

	// Test cmd2 (array format)
	cmd2, exists := config.Commands["cmd2"]
	if !exists {
		t.Fatal("cmd2 not found in config")
	}

	if len(cmd2.Env) != 2 {
		t.Errorf("Expected 2 env files for cmd2, got %d", len(cmd2.Env))
	}

	if cmd2.Env[0] != ".env.array1" {
		t.Errorf("Expected '.env.array1', got '%s'", cmd2.Env[0])
	}

	if cmd2.Env[1] != ".env.array2" {
		t.Errorf("Expected '.env.array2', got '%s'", cmd2.Env[1])
	}
}
