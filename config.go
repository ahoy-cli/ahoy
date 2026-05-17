package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// validateCommandAction is the Cobra handler for the 'ahoy config validate' command.
func validateCommandAction(cmd *cobra.Command, args []string) {
	configFile := AhoyConf.srcFile
	if configFile == "" {
		var err error
		configFile, err = getConfigPath("")
		if err != nil || configFile == "" {
			fmt.Println("Warning: No .ahoy.yml file found")
			fmt.Println("Run 'ahoy config init' to create a new configuration file")
			return
		}
	}

	result := RunConfigValidate(configFile)
	PrintConfigReport(result)

	// Exit with a non-zero status code when the configuration has errors.
	if !result.ConfigValid || result.ValidationResult.HasError {
		os.Exit(1)
	}
}
