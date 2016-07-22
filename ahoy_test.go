package main

import (
	"github.com/codegangsta/cli"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestOverrideExample(t *testing.T) {

	expected := "Overrode you.\n"
	actual, _ := appRun([]string{"ahoy", "docker", "override-example"})

	if expected != actual {
		t.Errorf("ahoy docker override-example: expected - %s; actual - %s", string(expected), string(actual))
	}
}

func appRun(args []string) (string, error) {
	std_out := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	app = cli.NewApp()
	app.Name = "ahoy"

	if sourcefile, err := getConfigPath(sourcefile); err == nil {
		sourcedir = filepath.Dir(sourcefile)
		config, _ := getConfig(sourcefile)
		app.Commands = getCommands(config)
	}

	app.Run(args)

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = std_out
	return string(out), nil
}
