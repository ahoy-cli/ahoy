package main

import (
	"fmt"

	"github.com/urfave/cli"
)

// validateCommandAction performs the config validate command functionality
func validateCommandAction(c *cli.Context) {
	configFile := AhoyConf.srcFile
	if configFile == "" {
		// Try to find a config file
		var err error
		configFile, err = getConfigPath("")
		if err != nil || configFile == "" {
			fmt.Println("Warning: No .ahoy.yml file found")
			fmt.Println("Run 'ahoy config init' to create a new configuration file")
			return
		}
	}

	// Skip validation during config loading for validate command
	options := ValidateOptions{SkipValidation: true}
	result := RunConfigValidate(configFile, options)
	PrintConfigReport(result)
}