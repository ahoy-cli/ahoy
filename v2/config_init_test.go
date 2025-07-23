package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunConfigInit_NewDirectory(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Skip network test in unit tests - we'll test this in BATS
	t.Skip("Network-dependent test - will be tested in BATS integration tests")
}

func TestRunConfigInit_ExistingFileForce(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create existing .ahoy.yml file
	existingContent := "ahoyapi: v2\ncommands:\n  test:\n    cmd: echo 'existing'"
	err = os.WriteFile(".ahoy.yml", []byte(existingContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing .ahoy.yml: %v", err)
	}

	// Skip network test in unit tests - we'll test this in BATS
	t.Skip("Network-dependent test - will be tested in BATS integration tests")
}

func TestInitArgs_Structure(t *testing.T) {
	// Test that InitArgs struct has expected fields
	args := InitArgs{
		Force: true,
		URL:   "https://example.com/test.yml",
	}

	if !args.Force {
		t.Error("Force field not working")
	}

	if args.URL != "https://example.com/test.yml" {
		t.Error("URL field not working")
	}
}

func TestFileExists_Helper(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// File should not exist initially
	if fileExists(testFile) {
		t.Error("fileExists should return false for non-existent file")
	}

	// Create the file
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// File should exist now
	if !fileExists(testFile) {
		t.Error("fileExists should return true for existing file")
	}

	// Directory should not be considered as existing file
	if fileExists(tmpDir) {
		t.Error("fileExists should return false for directories")
	}
}
