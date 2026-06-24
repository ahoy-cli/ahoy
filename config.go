package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// validateCommandAction is the Cobra handler for the 'ahoy config validate' command.
func (s *appState) validateCommandAction(cmd *cobra.Command, args []string) {
	configFile := s.srcFile
	if configFile == "" {
		fmt.Println("Warning: No .ahoy.yml file found")
		fmt.Println("Run 'ahoy config init' to create a new configuration file")
		return
	}

	result := RunConfigValidate(configFile)
	PrintConfigReport(result)

	// Exit with a non-zero status code when the configuration has errors.
	if !result.ConfigValid || result.ValidationResult.HasError {
		os.Exit(1)
	}
}
