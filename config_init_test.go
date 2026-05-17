package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestRunConfigInit_NewDirectory(t *testing.T) {
	// Network-dependent - tested in BATS integration tests.
	t.Skip("Network-dependent test - tested in BATS integration tests")
}

func TestRunConfigInit_ExistingFileForce(t *testing.T) {
	// Network-dependent - tested in BATS integration tests.
	t.Skip("Network-dependent test - tested in BATS integration tests")
}

func TestInitArgs_Structure(t *testing.T) {
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

func TestDownloadFile_InvalidURL(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yml")

	err := downloadFile("http://invalid-url-that-does-not-exist.local", testFile)
	if err == nil {
		t.Error("Expected error when downloading from invalid URL")
	}
	if fileExists(testFile) {
		t.Error("File should not be created when download fails")
	}
}

func TestDownloadFile_200Response(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ahoyapi: v2\ncommands: {}\n"))
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yml")

	if err := downloadFile(srv.URL+"/test.yml", testFile); err != nil {
		t.Fatalf("Expected no error for 200 response, got: %v", err)
	}
	if !fileExists(testFile) {
		t.Error("File should exist after successful download.")
	}
}

func TestDownloadFile_404Response(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yml")

	err := downloadFile(srv.URL+"/missing.yml", testFile)
	if err == nil {
		t.Error("Expected error when server returns 404.")
	}
	if fileExists(testFile) {
		t.Error("File should not be created when download returns 404.")
	}
}

func TestDownloadFile_InvalidScheme(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yml")

	for _, badURL := range []string{
		"file:///etc/passwd",
		"ftp://example.com/file.yml",
		"not-a-url",
		"",
	} {
		err := downloadFile(badURL, testFile)
		if err == nil {
			t.Errorf("Expected error for URL %q, got nil.", badURL)
		}
		if fileExists(testFile) {
			t.Errorf("File should not be created for invalid URL %q.", badURL)
		}
	}
}

func TestFileExists_Helper(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if fileExists(testFile) {
		t.Error("fileExists should return false for non-existent file")
	}

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if !fileExists(testFile) {
		t.Error("fileExists should return true for existing file")
	}

	if fileExists(tmpDir) {
		t.Error("fileExists should return false for directories")
	}
}
