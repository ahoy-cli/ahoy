package main

import (
	"github.com/codegangsta/cli"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestOverrideExample(t *testing.T) {
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

	app.Run([]string{"ahoy", "docker", "override-example"})

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = std_out

	expected := "Overrode you.\n"
	actual := string(out)

	if expected != actual {
		t.Errorf("ahoy docker override-example: expected - %s; actual - %s", string(expected), string(actual))
	}
}
