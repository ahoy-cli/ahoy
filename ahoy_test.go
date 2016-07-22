package main

import (
	"github.com/codegangsta/cli"
	"os"
	"path/filepath"
	"testing"
)

func TestOverrideExample(t *testing.T) {
	os.Args = []string{"ahoy", "docker", "override-example"}
	initFlags()
	app = cli.NewApp()
	app.Name = "ahoy"
	overrideFlags(app)
	if sourcefile, err := getConfigPath(sourcefile); err == nil {
		sourcedir = filepath.Dir(sourcefile)
		config, _ := getConfig(sourcefile)
		app.Commands = getCommands(config)
		app.Commands = addDefaultCommands(app.Commands)
		if config.Usage != "" {
			app.Usage = config.Usage
		}
	}

	app.Run(os.Args)
	// Output
	// "Overrode you."
}
