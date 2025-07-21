package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunDoctor_ConfigNotExists(t *testing.T) {
	// Test when config file doesn't exist
	result := RunDoctor("nonexistent.ahoy.yml")

	if result.ConfigExists {
		t.Error("Expected ConfigExists to be false for nonexistent file")
	}

	if result.ConfigValid {
		t.Error("Expected ConfigValid to be false for nonexistent file")
	}

	if len(result.Recommendations) == 0 {
		t.Error("Expected recommendations for missing config file")
	}

	expectedRec := "Create a .ahoy.yml file using 'ahoy init'"
	found := false
	for _, rec := range result.Recommendations {
		if rec == expectedRec {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected recommendation '%s' not found in: %v", expectedRec, result.Recommendations)
	}
}

func TestRunDoctor_InvalidYAML(t *testing.T) {
	// Create a temporary file with invalid YAML
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid.ahoy.yml")

	invalidYAML := `
ahoyapi: v2
commands:
  test:
    cmd: echo "test"
    invalid_yaml: [unclosed array
`

	err := os.WriteFile(configFile, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result := RunDoctor(configFile)

	if !result.ConfigExists {
		t.Error("Expected ConfigExists to be true for existing file")
	}

	if result.ConfigValid {
		t.Error("Expected ConfigValid to be false for invalid YAML")
	}

	expectedRec := "Fix YAML syntax errors in configuration file"
	found := false
	for _, rec := range result.Recommendations {
		if rec == expectedRec {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected recommendation '%s' not found in: %v", expectedRec, result.Recommendations)
	}
}

func TestRunDoctor_ValidConfig(t *testing.T) {
	// Create a temporary file with valid YAML
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "valid.ahoy.yml")

	validYAML := `
ahoyapi: v2
usage: Test configuration
commands:
  test:
    description: Test command
    cmd: echo "test"
`

	err := os.WriteFile(configFile, []byte(validYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result := RunDoctor(configFile)

	if !result.ConfigExists {
		t.Error("Expected ConfigExists to be true for existing file")
	}

	if !result.ConfigValid {
		t.Error("Expected ConfigValid to be true for valid YAML")
	}

	if result.APIVersion != "v2" {
		t.Errorf("Expected APIVersion to be 'v2', got '%s'", result.APIVersion)
	}

	if result.AhoyVersion == "" {
		t.Error("Expected AhoyVersion to be set")
	}
}

func TestRunDoctor_WithEnvironmentFiles(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test.ahoy.yml")

	// Create some environment files
	envFile1 := filepath.Join(tmpDir, ".env")

	err := os.WriteFile(envFile1, []byte("TEST=value"), 0644)
	if err != nil {
		t.Fatalf("Failed to create env file: %v", err)
	}
	// Don't create .env.local to test missing file detection
	configYAML := `
ahoyapi: v2
env:
  - .env
  - .env.local
commands:
  test:
    description: Test command
    cmd: echo "test"
    env:
      - .env.command
`

	err = os.WriteFile(configFile, []byte(configYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result := RunDoctor(configFile)

	if len(result.EnvFiles) != 3 {
		t.Errorf("Expected 3 environment files, got %d", len(result.EnvFiles))
	}

	// Check that .env exists and .env.local doesn't
	envFileStatuses := make(map[string]bool)
	for _, envFile := range result.EnvFiles {
		envFileStatuses[envFile.Path] = envFile.Exists
	}

	if !envFileStatuses[".env"] {
		t.Error("Expected .env to exist")
	}

	if envFileStatuses[".env.local"] {
		t.Error("Expected .env.local to not exist")
	}

	if envFileStatuses[".env.command"] {
		t.Error("Expected .env.command to not exist")
	}
}

func TestRunDoctor_WithImportFiles(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test.ahoy.yml")

	// Create one import file but not the other
	importFile1 := filepath.Join(tmpDir, "import1.ahoy.yml")
	err := os.WriteFile(importFile1, []byte("ahoyapi: v2\ncommands: {}"), 0644)
	if err != nil {
		t.Fatalf("Failed to create import file: %v", err)
	}

	configYAML := `
ahoyapi: v2
commands:
  test1:
    description: Test command 1
    cmd: echo "test1"
    imports:
      - import1.ahoy.yml
      - import2.ahoy.yml
  test2:
    description: Test command 2
    cmd: echo "test2"
    imports:
      - import3.ahoy.yml
    optional: true
`

	err = os.WriteFile(configFile, []byte(configYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result := RunDoctor(configFile)

	if len(result.ImportFiles) != 3 {
		t.Errorf("Expected 3 import files, got %d", len(result.ImportFiles))
	}

	// Check import file statuses
	importFileStatuses := make(map[string]ImportFileStatus)
	for _, importFile := range result.ImportFiles {
		importFileStatuses[importFile.Path] = importFile
	}

	if !importFileStatuses["import1.ahoy.yml"].Exists {
		t.Error("Expected import1.ahoy.yml to exist")
	}

	if importFileStatuses["import2.ahoy.yml"].Exists {
		t.Error("Expected import2.ahoy.yml to not exist")
	}

	if importFileStatuses["import3.ahoy.yml"].Optional != true {
		t.Error("Expected import3.ahoy.yml to be optional")
	}
}

func TestGenerateRecommendations_VersionMismatch(t *testing.T) {
	result := DoctorResult{
		ValidationResult: ValidationResult{
			Issues: []ValidationIssue{
				{
					Type:     "version_mismatch",
					Severity: "error",
					Message:  "Version mismatch",
				},
			},
		},
	}

	recommendations := generateRecommendations(result)

	expectedRec := "Upgrade Ahoy to the latest version for full feature support"
	found := false
	for _, rec := range recommendations {
		if rec == expectedRec {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected recommendation '%s' not found in: %v", expectedRec, recommendations)
	}
}

func TestGenerateRecommendations_MissingImportFiles(t *testing.T) {
	result := DoctorResult{
		ImportFiles: []ImportFileStatus{
			{
				Path:     "missing.ahoy.yml",
				Exists:   false,
				Optional: false,
			},
		},
		ValidationResult: ValidationResult{
			Issues: []ValidationIssue{},
		},
	}

	recommendations := generateRecommendations(result)

	expectedRec := "Create missing import files or mark them as optional"
	found := false
	for _, rec := range recommendations {
		if rec == expectedRec {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected recommendation '%s' not found in: %v", expectedRec, recommendations)
	}
}

func TestGenerateRecommendations_MissingEnvFiles(t *testing.T) {
	result := DoctorResult{
		EnvFiles: []EnvFileStatus{
			{
				Path:   ".env",
				Exists: false,
			},
		},
		ValidationResult: ValidationResult{
			Issues: []ValidationIssue{},
		},
	}

	recommendations := generateRecommendations(result)

	expectedRec := "Consider creating missing environment files or removing them from configuration"
	found := false
	for _, rec := range recommendations {
		if rec == expectedRec {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected recommendation '%s' not found in: %v", expectedRec, recommendations)
	}
}

func TestGenerateRecommendations_NewerFeatures(t *testing.T) {
	result := DoctorResult{
		ValidationResult: ValidationResult{
			Issues: []ValidationIssue{
				{
					Type:     "version_mismatch",
					Severity: "warning",
					Message:  "Using newer features",
				},
			},
		},
	}

	recommendations := generateRecommendations(result)

	expectedRec := "Consider upgrading to a newer Ahoy version for better support of advanced features"
	found := false
	for _, rec := range recommendations {
		if rec == expectedRec {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected recommendation '%s' not found in: %v", expectedRec, recommendations)
	}
}

func TestGenerateRecommendations_NoIssues(t *testing.T) {
	result := DoctorResult{
		ValidationResult: ValidationResult{
			Issues: []ValidationIssue{},
		},
		EnvFiles:    []EnvFileStatus{},
		ImportFiles: []ImportFileStatus{},
	}

	recommendations := generateRecommendations(result)

	expectedRec := "Configuration looks good! No issues found."
	found := false
	for _, rec := range recommendations {
		if rec == expectedRec {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected recommendation '%s' not found in: %v", expectedRec, recommendations)
	}
}

func TestCheckEnvironmentFiles(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test.ahoy.yml")

	// Create one env file but not the other
	envFile1 := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envFile1, []byte("TEST=value"), 0644)
	if err != nil {
		t.Fatalf("Failed to create env file: %v", err)
	}

	config := Config{
		Env: []string{".env", ".env.missing"},
		Commands: map[string]Command{
			"test": {
				Env: []string{".env.command"},
			},
		},
	}

	envFiles := checkEnvironmentFiles(config, configFile)

	if len(envFiles) != 3 {
		t.Errorf("Expected 3 environment files, got %d", len(envFiles))
	}

	// Check global env files
	globalCount := 0
	for _, envFile := range envFiles {
		if envFile.Global {
			globalCount++
		}
	}

	if globalCount != 2 {
		t.Errorf("Expected 2 global environment files, got %d", globalCount)
	}

	// Check that .env exists
	found := false
	for _, envFile := range envFiles {
		if envFile.Path == ".env" && envFile.Exists {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected .env to be found and exist")
	}
}

func TestCheckImportFiles(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test.ahoy.yml")

	// Create one import file but not the other
	importFile1 := filepath.Join(tmpDir, "import1.ahoy.yml")
	err := os.WriteFile(importFile1, []byte("ahoyapi: v2"), 0644)
	if err != nil {
		t.Fatalf("Failed to create import file: %v", err)
	}

	config := Config{
		Commands: map[string]Command{
			"test1": {
				Imports: []string{"import1.ahoy.yml", "import2.ahoy.yml"},
			},
			"test2": {
				Imports:  []string{"import3.ahoy.yml"},
				Optional: true,
			},
		},
	}

	importFiles := checkImportFiles(config, configFile)

	if len(importFiles) != 3 {
		t.Errorf("Expected 3 import files, got %d", len(importFiles))
	}

	// Check that import1.ahoy.yml exists
	found := false
	for _, importFile := range importFiles {
		if importFile.Path == "import1.ahoy.yml" && importFile.Exists {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected import1.ahoy.yml to be found and exist")
	}

	// Check that import3.ahoy.yml is marked as optional
	found = false
	for _, importFile := range importFiles {
		if importFile.Path == "import3.ahoy.yml" && importFile.Optional {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected import3.ahoy.yml to be marked as optional")
	}
}

func TestDoctorResult_Fields(t *testing.T) {
	// Test that DoctorResult has all expected fields
	result := DoctorResult{
		ConfigFile:       "test.ahoy.yml",
		ConfigExists:     true,
		ConfigValid:      true,
		APIVersion:       "v2",
		AhoyVersion:      "v2.3.0",
		ValidationResult: ValidationResult{},
		EnvFiles:         []EnvFileStatus{},
		ImportFiles:      []ImportFileStatus{},
		Recommendations:  []string{},
	}

	if result.ConfigFile != "test.ahoy.yml" {
		t.Error("ConfigFile field not working")
	}

	if !result.ConfigExists {
		t.Error("ConfigExists field not working")
	}

	if !result.ConfigValid {
		t.Error("ConfigValid field not working")
	}

	if result.APIVersion != "v2" {
		t.Error("APIVersion field not working")
	}

	if result.AhoyVersion != "v2.3.0" {
		t.Error("AhoyVersion field not working")
	}
}

func TestEnvFileStatus_Fields(t *testing.T) {
	// Test that EnvFileStatus has all expected fields
	envFile := EnvFileStatus{
		Path:   ".env",
		Exists: true,
		Global: true,
	}

	if envFile.Path != ".env" {
		t.Error("Path field not working")
	}

	if !envFile.Exists {
		t.Error("Exists field not working")
	}

	if !envFile.Global {
		t.Error("Global field not working")
	}
}

func TestImportFileStatus_Fields(t *testing.T) {
	// Test that ImportFileStatus has all expected fields
	importFile := ImportFileStatus{
		Path:     "import.ahoy.yml",
		Exists:   true,
		Optional: false,
		Command:  "test",
	}

	if importFile.Path != "import.ahoy.yml" {
		t.Error("Path field not working")
	}

	if !importFile.Exists {
		t.Error("Exists field not working")
	}

	if importFile.Optional {
		t.Error("Optional field not working")
	}

	if importFile.Command != "test" {
		t.Error("Command field not working")
	}
}
