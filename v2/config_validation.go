package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ValidationIssue represents a configuration validation problem
type ValidationIssue struct {
	Type            string // "version_mismatch", "unknown_field", "missing_file", "syntax_error"
	Severity        string // "error", "warning", "info"
	Message         string
	File            string
	Field           string
	Feature         string
	RequiredVersion string
	CurrentVersion  string
	Suggestion      string
}

// ValidationResult contains all validation issues found
type ValidationResult struct {
	Issues   []ValidationIssue
	HasError bool
}

// FeatureSupport defines which features are supported in which versions
var FeatureSupport = map[string]string{
	"command_aliases":    "v2.1.0",
	"optional_imports":   "v2.2.0",
	"multiple_env_files": "v2.5.0",
	"schema_validation":  "v2.6.0", // This version we're implementing
}

// GetAhoyVersion returns the current Ahoy version
func GetAhoyVersion() string {
	// Allow simulation of older versions for testing
	if simulateVersion != "" {
		return simulateVersion
	}
	if version == "" {
		return "development"
	}
	return version
}

// VersionSupports checks if a given version supports a specific feature
func VersionSupports(currentVersion, feature string) bool {
	requiredVersion, exists := FeatureSupport[feature]
	if !exists {
		return true // Unknown features are assumed supported
	}

	// For development versions, assume all features are supported
	if currentVersion == "development" || currentVersion == "" {
		return true
	}

	return compareVersions(currentVersion, requiredVersion) >= 0
}

// compareVersions compares two semantic version strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	// Strip 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < 3; i++ {
		var p1, p2 int
		if i < len(parts1) {
			p1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			p2, _ = strconv.Atoi(parts2[i])
		}
		if p1 < p2 {
			return -1
		} else if p1 > p2 {
			return 1
		}
	}
	return 0
}

// ValidateConfig performs comprehensive validation of an Ahoy configuration
func ValidateConfig(config Config, configFile string) ValidationResult {
	result := ValidationResult{
		Issues: []ValidationIssue{},
	}

	currentVersion := GetAhoyVersion()

	// Validate API version
	if config.AhoyAPI != "v2" {
		result.Issues = append(result.Issues, ValidationIssue{
			Type:     "version_mismatch",
			Severity: "error",
			Message:  fmt.Sprintf("Unsupported API version '%s'. Only 'v2' is currently supported.", config.AhoyAPI),
			File:     configFile,
			Field:    "ahoyapi",
		})
		result.HasError = true
	}

	// Check for features that might not be supported
	result.Issues = append(result.Issues, validateFeatures(config, configFile, currentVersion)...)

	// Validate commands
	for cmdName, cmd := range config.Commands {
		cmdIssues := validateCommand(cmdName, cmd, configFile, currentVersion)
		result.Issues = append(result.Issues, cmdIssues...)
	}

	// Check for any error-level issues
	for _, issue := range result.Issues {
		if issue.Severity == "error" {
			result.HasError = true
			break
		}
	}

	return result
}

// validateFeatures checks for features that might not be supported in the current version
func validateFeatures(config Config, configFile, currentVersion string) []ValidationIssue {
	var issues []ValidationIssue

	// Check for multiple environment files
	if len(config.Env) > 1 {
		if !VersionSupports(currentVersion, "multiple_env_files") {
			issues = append(issues, ValidationIssue{
				Type:            "version_mismatch",
				Severity:        "warning",
				Message:         "Multiple environment files detected. This feature requires proper support.",
				File:            configFile,
				Field:           "env",
				Feature:         "multiple_env_files",
				RequiredVersion: FeatureSupport["multiple_env_files"],
				CurrentVersion:  currentVersion,
				Suggestion:      "This should work but consider upgrading for better support.",
			})
		}
	}

	return issues
}

// validateCommand validates a single command configuration
func validateCommand(cmdName string, cmd Command, configFile, currentVersion string) []ValidationIssue {
	var issues []ValidationIssue

	// Check for optional imports
	if cmd.Optional && !VersionSupports(currentVersion, "optional_imports") {
		issues = append(issues, ValidationIssue{
			Type:            "version_mismatch",
			Severity:        "error",
			Message:         fmt.Sprintf("Command '%s' uses 'optional: true' which requires Ahoy %s or later", cmdName, FeatureSupport["optional_imports"]),
			File:            configFile,
			Field:           fmt.Sprintf("commands.%s.optional", cmdName),
			Feature:         "optional_imports",
			RequiredVersion: FeatureSupport["optional_imports"],
			CurrentVersion:  currentVersion,
			Suggestion:      "Upgrade Ahoy or remove 'optional: true' from the command",
		})
	}

	// Check for command aliases
	if len(cmd.Aliases) > 0 && !VersionSupports(currentVersion, "command_aliases") {
		issues = append(issues, ValidationIssue{
			Type:            "version_mismatch",
			Severity:        "warning",
			Message:         fmt.Sprintf("Command '%s' uses aliases which require Ahoy %s or later", cmdName, FeatureSupport["command_aliases"]),
			File:            configFile,
			Field:           fmt.Sprintf("commands.%s.aliases", cmdName),
			Feature:         "command_aliases",
			RequiredVersion: FeatureSupport["command_aliases"],
			CurrentVersion:  currentVersion,
			Suggestion:      "Upgrade Ahoy for full alias support",
		})
	}

	// Validate imports exist (if not optional)
	if cmd.Imports != nil {
		for _, importPath := range cmd.Imports {
			issues = append(issues, validateImport(cmdName, importPath, cmd.Optional, configFile, currentVersion)...)
		}
	}

	// Validate environment files exist
	if len(cmd.Env) > 0 {
		for _, envPath := range cmd.Env {
			issues = append(issues, validateEnvFile(cmdName, envPath, configFile)...)
		}
	}

	return issues
}

// validateImport validates a single import file
func validateImport(cmdName, importPath string, optional bool, configFile, currentVersion string) []ValidationIssue {
	var issues []ValidationIssue

	// Make path relative to config file
	configDir := filepath.Dir(configFile)
	fullPath := expandPath(importPath, configDir)

	if !fileExists(fullPath) {
		if optional {
			if !VersionSupports(currentVersion, "optional_imports") {
				issues = append(issues, ValidationIssue{
					Type:            "version_mismatch",
					Severity:        "error",
					Message:         fmt.Sprintf("Import file '%s' not found for command '%s'. This file is marked as optional but your Ahoy version doesn't support optional imports.", importPath, cmdName),
					File:            configFile,
					Field:           fmt.Sprintf("commands.%s.imports", cmdName),
					Feature:         "optional_imports",
					RequiredVersion: FeatureSupport["optional_imports"],
					CurrentVersion:  currentVersion,
					Suggestion:      fmt.Sprintf("Either upgrade Ahoy to %s+, create the missing file '%s', or remove 'optional: true'", FeatureSupport["optional_imports"], importPath),
				})
			} else {
				issues = append(issues, ValidationIssue{
					Type:     "missing_file",
					Severity: "info",
					Message:  fmt.Sprintf("Optional import file '%s' not found for command '%s' (this is OK)", importPath, cmdName),
					File:     configFile,
					Field:    fmt.Sprintf("commands.%s.imports", cmdName),
				})
			}
		} else {
			// For missing required imports, only show as warning since getSubCommands handles this gracefully
			// This maintains backwards compatibility with existing behavior
			issues = append(issues, ValidationIssue{
				Type:       "missing_file",
				Severity:   "warning",
				Message:    fmt.Sprintf("Import file '%s' not found for command '%s' (will be skipped)", importPath, cmdName),
				File:       configFile,
				Field:      fmt.Sprintf("commands.%s.imports", cmdName),
				Suggestion: fmt.Sprintf("Create the file '%s' or mark the import as 'optional: true'", importPath),
			})
		}
	}

	return issues
}

// validateEnvFile validates that environment files exist
func validateEnvFile(cmdName, envPath, configFile string) []ValidationIssue {
	var issues []ValidationIssue

	// Expand path (handles tilde, absolute, and relative paths)
	configDir := filepath.Dir(configFile)
	fullPath := expandPath(envPath, configDir)

	if !fileExists(fullPath) {
		issues = append(issues, ValidationIssue{
			Type:       "missing_file",
			Severity:   "warning",
			Message:    fmt.Sprintf("Environment file '%s' not found for command '%s' (will be ignored)", envPath, cmdName),
			File:       configFile,
			Field:      fmt.Sprintf("commands.%s.env", cmdName),
			Suggestion: fmt.Sprintf("Create the file '%s' or remove it from the configuration", envPath),
		})
	}

	return issues
}

// PrintValidationIssues prints validation issues in a user-friendly format
func PrintValidationIssues(result ValidationResult) {
	if len(result.Issues) == 0 {
		return
	}

	fmt.Fprintf(os.Stderr, "\nConfiguration Validation Issues:\n")
	fmt.Fprintf(os.Stderr, "================================\n\n")

	errorCount := 0
	warningCount := 0
	infoCount := 0

	for _, issue := range result.Issues {
		switch issue.Severity {
		case "error":
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", issue.Message)
			errorCount++
		case "warning":
			fmt.Fprintf(os.Stderr, "WARNING: %s\n", issue.Message)
			warningCount++
		case "info":
			fmt.Fprintf(os.Stderr, "INFO: %s\n", issue.Message)
			infoCount++
		}

		if issue.File != "" {
			fmt.Fprintf(os.Stderr, "File: %s\n", issue.File)
		}
		if issue.Field != "" {
			fmt.Fprintf(os.Stderr, "Field: %s\n", issue.Field)
		}
		if issue.RequiredVersion != "" && issue.CurrentVersion != "" {
			fmt.Fprintf(os.Stderr, "Required Version: %s (current: %s)\n", issue.RequiredVersion, issue.CurrentVersion)
		}
		if issue.Suggestion != "" {
			fmt.Fprintf(os.Stderr, "Suggestion: %s\n", issue.Suggestion)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	fmt.Fprintf(os.Stderr, "Summary: %d error(s), %d warning(s), %d info\n", errorCount, warningCount, infoCount)

	if errorCount > 0 {
		fmt.Fprintf(os.Stderr, "\nRun 'ahoy config validate' for more detailed diagnostics and solutions.\n")
	}
}

// ConfigReport contains comprehensive diagnostic information
type ConfigReport struct {
	ConfigFile       string
	ConfigExists     bool
	ConfigValid      bool
	APIVersion       string
	AhoyVersion      string
	ValidationResult ValidationResult
	EnvFiles         []EnvFileStatus
	ImportFiles      []ImportFileStatus
	Recommendations  []string
}

// EnvFileStatus represents the status of an environment file
type EnvFileStatus struct {
	Path   string
	Exists bool
	Global bool
}

// ImportFileStatus represents the status of an import file
type ImportFileStatus struct {
	Path     string
	Exists   bool
	Optional bool
	Command  string
}

// ValidateOptions contains options for configuration validation
type ValidateOptions struct {
	SkipValidation bool
}

// RunConfigValidate performs comprehensive diagnostics on the Ahoy configuration
func RunConfigValidate(configFile string, options ValidateOptions) ConfigReport {
	result := ConfigReport{
		ConfigFile:      configFile,
		AhoyVersion:     GetAhoyVersion(),
		EnvFiles:        []EnvFileStatus{},
		ImportFiles:     []ImportFileStatus{},
		Recommendations: []string{},
	}

	// Check if config file exists
	result.ConfigExists = fileExists(configFile)
	if !result.ConfigExists {
		result.Recommendations = append(result.Recommendations, "Create a .ahoy.yml file using 'ahoy config init'")
		return result
	}

	// Try to load and parse the config
	config, err := getConfig(configFile, options.SkipValidation)
	if err != nil {
		result.ConfigValid = false
		result.Recommendations = append(result.Recommendations, "Fix YAML syntax errors in configuration file")
		return result
	}

	result.ConfigValid = true
	result.APIVersion = config.AhoyAPI

	// Run validation
	result.ValidationResult = ValidateConfig(config, configFile)

	// Check environment files
	result.EnvFiles = checkEnvironmentFiles(config, configFile)

	// Check import files
	result.ImportFiles = checkImportFiles(config, configFile)

	// Generate recommendations
	result.Recommendations = generateRecommendations(result)

	return result
}

// checkEnvironmentFiles checks the status of all environment files
func checkEnvironmentFiles(config Config, configFile string) []EnvFileStatus {
	var envFiles []EnvFileStatus
	configDir := filepath.Dir(configFile)

	// Check global environment files
	for _, envPath := range config.Env {
		fullPath := filepath.Join(configDir, envPath)
		envFiles = append(envFiles, EnvFileStatus{
			Path:   envPath,
			Exists: fileExists(fullPath),
			Global: true,
		})
	}

	// Check command-specific environment files
	for _, cmd := range config.Commands {
		for _, envPath := range cmd.Env {
			fullPath := filepath.Join(configDir, envPath)
			envFiles = append(envFiles, EnvFileStatus{
				Path:   envPath,
				Exists: fileExists(fullPath),
				Global: false,
			})
		}
	}

	return envFiles
}

// checkImportFiles checks the status of all import files
func checkImportFiles(config Config, configFile string) []ImportFileStatus {
	var importFiles []ImportFileStatus
	configDir := filepath.Dir(configFile)

	for cmdName, cmd := range config.Commands {
		for _, importPath := range cmd.Imports {
			fullPath := expandPath(importPath, configDir)

			importFiles = append(importFiles, ImportFileStatus{
				Path:     importPath,
				Exists:   fileExists(fullPath),
				Optional: cmd.Optional,
				Command:  cmdName,
			})
		}
	}

	return importFiles
}

// generateRecommendations generates actionable recommendations based on the diagnostic results
func generateRecommendations(result ConfigReport) []string {
	var recommendations []string

	// Check for version mismatches
	hasVersionMismatch := false
	for _, issue := range result.ValidationResult.Issues {
		if issue.Type == "version_mismatch" && issue.Severity == "error" {
			hasVersionMismatch = true
			break
		}
	}

	if hasVersionMismatch {
		recommendations = append(recommendations, "Upgrade Ahoy to the latest version for full feature support")
	}

	// Check for missing required files
	hasMissingFiles := false
	for _, importFile := range result.ImportFiles {
		if !importFile.Exists && !importFile.Optional {
			hasMissingFiles = true
			break
		}
	}

	if hasMissingFiles {
		recommendations = append(recommendations, "Create missing import files or mark them as optional")
	}

	// Check for missing environment files
	missingEnvFiles := 0
	for _, envFile := range result.EnvFiles {
		if !envFile.Exists {
			missingEnvFiles++
		}
	}

	if missingEnvFiles > 0 {
		recommendations = append(recommendations, "Consider creating missing environment files or removing them from configuration")
	}

	// Check if using newer features
	usingNewerFeatures := false
	for _, issue := range result.ValidationResult.Issues {
		if issue.Type == "version_mismatch" && issue.Severity == "warning" {
			usingNewerFeatures = true
			break
		}
	}

	if usingNewerFeatures {
		recommendations = append(recommendations, "Consider upgrading to a newer Ahoy version for better support of advanced features")
	}

	// If everything looks good
	if len(result.ValidationResult.Issues) == 0 && len(recommendations) == 0 {
		recommendations = append(recommendations, "Configuration looks good! No issues found.")
	}

	return recommendations
}

// PrintConfigReport prints a comprehensive diagnostic report
func PrintConfigReport(result ConfigReport) {
	fmt.Println("Ahoy Configuration Validator")
	fmt.Println("========================")
	fmt.Println()

	// Basic information
	fmt.Printf("📁 Configuration file: %s ", result.ConfigFile)
	if result.ConfigExists {
		fmt.Println("✅ (found)")
	} else {
		fmt.Println("❌ (not found)")
		fmt.Println()
		fmt.Println("💡 Run 'ahoy config init' to create a new configuration file")
		return
	}

	fmt.Printf("📋 API Version: %s ", result.APIVersion)
	if result.APIVersion == "v2" {
		fmt.Println("✅ (supported)")
	} else {
		fmt.Println("❌ (unsupported)")
	}

	fmt.Printf("🔧 Ahoy Version: %s\n", result.AhoyVersion)

	fmt.Printf("✅ Syntax: ")
	if result.ConfigValid {
		fmt.Println("Valid YAML")
	} else {
		fmt.Println("❌ Invalid YAML")
	}

	fmt.Println()

	// Validation issues
	if len(result.ValidationResult.Issues) > 0 {
		fmt.Println("🔍 Issues Found:")
		fmt.Println()

		for i, issue := range result.ValidationResult.Issues {
			fmt.Printf("%d. ", i+1)
			switch issue.Severity {
			case "error":
				fmt.Printf("❌ %s\n", issue.Message)
			case "warning":
				fmt.Printf("⚠️  %s\n", issue.Message)
			case "info":
				fmt.Printf("ℹ️  %s\n", issue.Message)
			}

			if issue.Field != "" {
				fmt.Printf("   📍 Location: %s\n", issue.Field)
			}
			if issue.RequiredVersion != "" {
				fmt.Printf("   📦 Required Version: %s (current: %s)\n", issue.RequiredVersion, issue.CurrentVersion)
			}
			if issue.Suggestion != "" {
				fmt.Printf("   💡 Fix: %s\n", issue.Suggestion)
			}
			fmt.Println()
		}
	} else {
		fmt.Println("✅ No validation issues found")
		fmt.Println()
	}

	// Environment files status
	if len(result.EnvFiles) > 0 {
		fmt.Println("🌍 Environment Files:")
		for _, envFile := range result.EnvFiles {
			scope := "command-specific"
			if envFile.Global {
				scope = "global"
			}

			if envFile.Exists {
				fmt.Printf("   ✅ %s (%s)\n", envFile.Path, scope)
			} else {
				fmt.Printf("   ❌ %s (%s) - missing\n", envFile.Path, scope)
			}
		}
		fmt.Println()
	}

	// Import files status
	if len(result.ImportFiles) > 0 {
		fmt.Println("📥 Import Files:")
		for _, importFile := range result.ImportFiles {
			status := "required"
			if importFile.Optional {
				status = "optional"
			}

			if importFile.Exists {
				fmt.Printf("   ✅ %s (%s, command: %s)\n", importFile.Path, status, importFile.Command)
			} else {
				if importFile.Optional {
					fmt.Printf("   ⚠️  %s (%s, command: %s) - missing but OK\n", importFile.Path, status, importFile.Command)
				} else {
					fmt.Printf("   ❌ %s (%s, command: %s) - missing\n", importFile.Path, status, importFile.Command)
				}
			}
		}
		fmt.Println()
	}

	// Recommendations
	if len(result.Recommendations) > 0 {
		fmt.Println("💡 Recommendations:")
		for i, rec := range result.Recommendations {
			fmt.Printf("%d. %s\n", i+1, rec)
		}
		fmt.Println()
	}

	// Summary
	if result.ValidationResult.HasError {
		fmt.Println("❌ Configuration has errors that need to be fixed")
	} else if len(result.ValidationResult.Issues) > 0 {
		fmt.Println("⚠️  Configuration has warnings but should work")
	} else {
		fmt.Println("✅ Configuration looks great!")
	}
}
