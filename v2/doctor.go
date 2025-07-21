package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// DoctorResult contains comprehensive diagnostic information
type DoctorResult struct {
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

// RunDoctor performs comprehensive diagnostics on the Ahoy configuration
func RunDoctor(configFile string) DoctorResult {
	result := DoctorResult{
		ConfigFile:      configFile,
		AhoyVersion:     GetAhoyVersion(),
		EnvFiles:        []EnvFileStatus{},
		ImportFiles:     []ImportFileStatus{},
		Recommendations: []string{},
	}

	// Check if config file exists
	result.ConfigExists = fileExists(configFile)
	if !result.ConfigExists {
		result.Recommendations = append(result.Recommendations, "Create a .ahoy.yml file using 'ahoy init'")
		return result
	}

	// Try to load and parse the config
	config, err := getConfig(configFile)
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
			fullPath := importPath
			if !strings.HasPrefix(importPath, "/") && !strings.HasPrefix(importPath, "~") {
				fullPath = filepath.Join(configDir, importPath)
			}

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
func generateRecommendations(result DoctorResult) []string {
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

// PrintDoctorReport prints a comprehensive diagnostic report
func PrintDoctorReport(result DoctorResult) {
	fmt.Println("Ahoy Configuration Doctor")
	fmt.Println("========================")
	fmt.Println()

	// Basic information
	fmt.Printf("📁 Configuration file: %s ", result.ConfigFile)
	if result.ConfigExists {
		fmt.Println("✅ (found)")
	} else {
		fmt.Println("❌ (not found)")
		fmt.Println()
		fmt.Println("💡 Run 'ahoy init' to create a new configuration file")
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
