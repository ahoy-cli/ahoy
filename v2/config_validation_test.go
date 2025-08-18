package main

import (
	"slices"
	"testing"
)

func TestRunConfigValidate_ConfigNotExists(t *testing.T) {
	// Test when config file doesn't exist
	result := RunConfigValidate("nonexistent.ahoy.yml", ValidateOptions{SkipValidation: false})

	if result.ConfigExists {
		t.Error("Expected ConfigExists to be false for nonexistent file")
	}

	if result.ConfigValid {
		t.Error("Expected ConfigValid to be false for nonexistent file")
	}

	if len(result.Recommendations) == 0 {
		t.Error("Expected recommendations for missing config file")
	}

	expectedRec := "Create a .ahoy.yml file using 'ahoy config init'"
	found := slices.Contains(result.Recommendations, expectedRec)
	if !found {
		t.Errorf("Expected recommendation '%s' not found in: %v", expectedRec, result.Recommendations)
	}
}

func TestRunConfigValidate_InvalidYAML(t *testing.T) {
	// Use static test file with invalid YAML
	configFile := "testdata/invalid-yaml.ahoy.yml"

	result := RunConfigValidate(configFile, ValidateOptions{SkipValidation: false})

	if !result.ConfigExists {
		t.Error("Expected ConfigExists to be true for existing file")
	}

	if result.ConfigValid {
		t.Error("Expected ConfigValid to be false for invalid YAML")
	}

	expectedRec := "Fix YAML syntax errors in configuration file"
	found := slices.Contains(result.Recommendations, expectedRec)
	if !found {
		t.Errorf("Expected recommendation '%s' not found in: %v", expectedRec, result.Recommendations)
	}
}

func TestRunConfigValidate_ValidConfig(t *testing.T) {
	// Use static test file with valid YAML
	configFile := "testdata/simple.ahoy.yml"

	result := RunConfigValidate(configFile, ValidateOptions{SkipValidation: false})

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

func TestRunConfigValidate_WithEnvironmentFiles(t *testing.T) {
	// Use static test file that references env files
	configFile := "testdata/with-env-files.ahoy.yml"

	result := RunConfigValidate(configFile, ValidateOptions{SkipValidation: false})

	if len(result.EnvFiles) != 3 {
		t.Errorf("Expected 3 environment files, got %d", len(result.EnvFiles))
	}

	// Check that env files are detected (they won't exist in testdata)
	envFileStatuses := make(map[string]bool)
	for _, envFile := range result.EnvFiles {
		envFileStatuses[envFile.Path] = envFile.Exists
	}

	// Most env files should be missing, but we can check they're detected
	// Check that files are detected in the result
	envPaths := make([]string, 0)
	for _, envFile := range result.EnvFiles {
		envPaths = append(envPaths, envFile.Path)
	}

	// Should detect .env, .env.local, and .env.command
	expectedPaths := map[string]bool{".env": false, ".env.local": false, ".env.command": false}
	for _, path := range envPaths {
		if _, ok := expectedPaths[path]; ok {
			expectedPaths[path] = true
		}
	}

	for path, found := range expectedPaths {
		if !found {
			t.Errorf("Expected to find env file %s in validation result", path)
		}
	}
}

func TestRunConfigValidate_WithImportFiles(t *testing.T) {
	// Use static test file that references import files
	configFile := "testdata/with-imports.ahoy.yml"

	result := RunConfigValidate(configFile, ValidateOptions{SkipValidation: false})

	if len(result.ImportFiles) != 3 {
		t.Errorf("Expected 3 import files, got %d", len(result.ImportFiles))
	}

	// Check import file statuses
	importFileStatuses := make(map[string]ImportFileStatus)
	for _, importFile := range result.ImportFiles {
		importFileStatuses[importFile.Path] = importFile
	}

	// simple.ahoy.yml should exist, missing files should not
	if !importFileStatuses["simple.ahoy.yml"].Exists {
		t.Error("Expected simple.ahoy.yml to exist in testdata")
	}

	if importFileStatuses["missing-import.ahoy.yml"].Exists {
		t.Error("Expected missing-import.ahoy.yml to not exist")
	}

	if importFileStatuses["another-missing.ahoy.yml"].Optional != true {
		t.Error("Expected another-missing.ahoy.yml to be optional")
	}
}

func TestGenerateRecommendations_VersionMismatch(t *testing.T) {
	result := ConfigReport{
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
	found := slices.Contains(recommendations, expectedRec)
	if !found {
		t.Errorf("Expected recommendation '%s' not found in: %v", expectedRec, recommendations)
	}
}

func TestGenerateRecommendations_MissingImportFiles(t *testing.T) {
	result := ConfigReport{
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
	found := slices.Contains(recommendations, expectedRec)
	if !found {
		t.Errorf("Expected recommendation '%s' not found in: %v", expectedRec, recommendations)
	}
}

func TestGenerateRecommendations_MissingEnvFiles(t *testing.T) {
	result := ConfigReport{
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
	found := slices.Contains(recommendations, expectedRec)
	if !found {
		t.Errorf("Expected recommendation '%s' not found in: %v", expectedRec, recommendations)
	}
}

func TestGenerateRecommendations_NewerFeatures(t *testing.T) {
	result := ConfigReport{
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
	found := slices.Contains(recommendations, expectedRec)
	if !found {
		t.Errorf("Expected recommendation '%s' not found in: %v", expectedRec, recommendations)
	}
}

func TestGenerateRecommendations_NoIssues(t *testing.T) {
	result := ConfigReport{
		ValidationResult: ValidationResult{
			Issues: []ValidationIssue{},
		},
		EnvFiles:    []EnvFileStatus{},
		ImportFiles: []ImportFileStatus{},
	}

	recommendations := generateRecommendations(result)

	expectedRec := "Configuration looks good! No issues found."
	found := slices.Contains(recommendations, expectedRec)
	if !found {
		t.Errorf("Expected recommendation '%s' not found in: %v", expectedRec, recommendations)
	}
}

func TestCheckEnvironmentFiles(t *testing.T) {
	// Use testdata directory
	configFile := "testdata/with-env-files.ahoy.yml"

	// We have a test env file in testdata (.env.test)

	config := Config{
		Env: []string{".env.test", ".env.missing"},
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

	// Check that .env.test exists
	found := false
	for _, envFile := range envFiles {
		if envFile.Path == ".env.test" && envFile.Exists {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected .env.test to be found and exist")
	}
}

func TestCheckImportFiles(t *testing.T) {
	// Use testdata directory
	configFile := "testdata/with-imports.ahoy.yml"

	config := Config{
		Commands: map[string]Command{
			"test1": {
				Imports: []string{"simple.ahoy.yml", "missing-import.ahoy.yml"},
			},
			"test2": {
				Imports:  []string{"another-missing.ahoy.yml"},
				Optional: true,
			},
		},
	}

	importFiles := checkImportFiles(config, configFile)

	if len(importFiles) != 3 {
		t.Errorf("Expected 3 import files, got %d", len(importFiles))
	}

	// Check that simple.ahoy.yml exists
	found := false
	for _, importFile := range importFiles {
		if importFile.Path == "simple.ahoy.yml" && importFile.Exists {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected simple.ahoy.yml to be found and exist")
	}

	// Check that another-missing.ahoy.yml is marked as optional
	found = false
	for _, importFile := range importFiles {
		if importFile.Path == "another-missing.ahoy.yml" && importFile.Optional {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected another-missing.ahoy.yml to be marked as optional")
	}
}

func TestConfigReport_Fields(t *testing.T) {
	// Test that ConfigReport has all expected fields
	result := ConfigReport{
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

func TestRunConfigValidate_WrongAPIVersion(t *testing.T) {
	// Use static test file with wrong API version
	configFile := "testdata/wrong-api-version.ahoy.yml"

	result := RunConfigValidate(configFile, ValidateOptions{SkipValidation: false})

	if !result.ConfigExists {
		t.Error("Expected ConfigExists to be true for existing file")
	}

	if result.ConfigValid {
		t.Error("Expected ConfigValid to be false due to wrong API version")
	}

	// API version won't be set if config loading failed due to wrong version
	if result.APIVersion != "" {
		t.Errorf("Expected APIVersion to be empty when config loading fails, got '%s'", result.APIVersion)
	}
}

func TestRunConfigValidate_MissingAPIVersion(t *testing.T) {
	// Use static test file with missing API version
	configFile := "testdata/missing-api-version.ahoy.yml"

	result := RunConfigValidate(configFile, ValidateOptions{SkipValidation: false})

	if !result.ConfigExists {
		t.Error("Expected ConfigExists to be true for existing file")
	}

	if result.ConfigValid {
		t.Error("Expected ConfigValid to be false due to missing API version")
	}

	if result.APIVersion != "" {
		t.Errorf("Expected APIVersion to be empty, got '%s'", result.APIVersion)
	}
}

func TestRunConfigValidate_WithNewerFeatures(t *testing.T) {
	// Use static test file with newer features (aliases, optional imports)
	configFile := "testdata/newer-features.ahoy.yml"

	result := RunConfigValidate(configFile, ValidateOptions{SkipValidation: false})

	if !result.ConfigExists {
		t.Error("Expected ConfigExists to be true for existing file")
	}

	if !result.ConfigValid {
		t.Error("Expected ConfigValid to be true for valid YAML with newer features")
	}

	if result.APIVersion != "v2" {
		t.Errorf("Expected APIVersion to be 'v2', got '%s'", result.APIVersion)
	}

	// Should have validation issues about newer features if simulating older version
	// This test mainly ensures the file loads correctly
}
