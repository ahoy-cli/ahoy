package main

import (
	"github.com/codegangsta/cli"
	"path/filepath"
	"testing"
)

func TestOverrideExample(t *testing.T) {
	app = cli.NewApp()
	app.Name = "ahoy"
	if sourcefile, err := getConfigPath(sourcefile); err == nil {
		sourcedir = filepath.Dir(sourcefile)
		config, _ := getConfig(sourcefile)
		app.Commands = getCommands(config)
	}

	app.Run([]string{"ahoy", "docker", "override-example"})
	// Output
	// "Overrode you."
}
