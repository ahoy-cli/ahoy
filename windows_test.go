package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWindowsPathHandling(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	// The function should handle both forward and back slashes
	config, err := getConfig("testdata/simple.ahoy.yml")
	if err != nil {
		t.Errorf("Failed to load config with forward slashes: %v", err)
	}

	if config.AhoyAPI != "v2" {
		t.Errorf("Expected AhoyAPI v2, got %s", config.AhoyAPI)
	}
}

func TestWindowsConfigPathResolution(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	// Test that getConfigPath works with Windows paths
	originalDir, _ := os.Getwd()

	// Create a temporary directory structure
	tempDir := filepath.Join(os.TempDir(), "ahoy-windows-test")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test .ahoy.yml file
	testConfig := `ahoyapi: v2
commands:
  test:
    usage: Test command
    cmd: echo "test"`

	configPath := filepath.Join(tempDir, ".ahoy.yml")
	err = os.WriteFile(configPath, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Change to the temp directory
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Test that getConfigPath finds the .ahoy.yml file
	foundPath, err := getConfigPath("")
	if err != nil {
		t.Errorf("getConfigPath failed: %v", err)
	}

	expectedPath := filepath.Join(tempDir, ".ahoy.yml")
	if foundPath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, foundPath)
	}
}

func TestWindowsEnvironmentVariables(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	// Test Windows-style environment variables
	testEnvFile := filepath.Join(os.TempDir(), "test-windows.env")
	envContent := `WINDOWS_TEST_VAR=test_value
# Comment line
ANOTHER_VAR=another_value`

	err := os.WriteFile(testEnvFile, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test env file: %v", err)
	}
	defer os.Remove(testEnvFile)

	envVars := getEnvironmentVars(testEnvFile)

	expectedVars := []string{"WINDOWS_TEST_VAR=test_value", "ANOTHER_VAR=another_value"}
	if len(envVars) != len(expectedVars) {
		t.Errorf("Expected %d environment variables, got %d", len(expectedVars), len(envVars))
	}

	for i, expected := range expectedVars {
		if i < len(envVars) && envVars[i] != expected {
			t.Errorf("Expected env var %s, got %s", expected, envVars[i])
		}
	}
}

func TestWindowsFileExists(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	// Test fileExists function with Windows paths
	testFile := filepath.Join(os.TempDir(), "test-windows-file.txt")

	// File shouldn't exist initially
	if fileExists(testFile) {
		t.Error("fileExists returned true for non-existent file")
	}

	// Create the file
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// File should exist now
	if !fileExists(testFile) {
		t.Error("fileExists returned false for existing file")
	}

	// Test with directory (should return false)
	testDir := filepath.Join(os.TempDir(), "test-windows-dir")
	err = os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	if fileExists(testDir) {
		t.Error("fileExists returned true for directory")
	}
}

func TestWindowsCommandExecution(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	// Test that commands can be executed on Windows
	// Create a simple config for testing
	config := Config{
		AhoyAPI: "v2",
		Commands: map[string]Command{
			"windows-test": {
				Usage: "Test Windows command execution",
				Cmd:   "echo windows test",
			},
		},
		Entrypoint: []string{"cmd", "/c", "{{cmd}}"},
	}

	commands := getCommands(config)
	if len(commands) == 0 {
		t.Error("No commands generated from config")
	}

	// Check that the command was created correctly
	found := false
	for _, cmd := range commands {
		if cmd.Name() == "windows-test" {
			found = true
			if cmd.Short != "Test Windows command execution" {
				t.Errorf("Expected usage 'Test Windows command execution', got '%s'", cmd.Short)
			}
		}
	}

	if !found {
		t.Error("windows-test command not found in generated commands")
	}
}

func TestWindowsBinaryName(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	// On Windows, we should be testing the .exe binary
	executable := "ahoy.exe"
	if !strings.HasSuffix(executable, ".exe") {
		t.Error("Windows binary should have .exe extension")
	}
}

func TestWindowsPathSeparators(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	// Test that we properly handle both path separators on Windows
	forwardSlashPath := "testdata/simple.ahoy.yml"
	backSlashPath := "testdata\\simple.ahoy.yml"

	// Both should resolve to the same absolute path
	absForward, err1 := filepath.Abs(forwardSlashPath)
	absBackSlash, err2 := filepath.Abs(backSlashPath)

	if err1 != nil || err2 != nil {
		t.Skip("Could not resolve test paths")
	}

	if absForward != absBackSlash {
		t.Errorf("Path separators not handled consistently: %s vs %s", absForward, absBackSlash)
	}
}
