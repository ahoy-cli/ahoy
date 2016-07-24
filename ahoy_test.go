package main

import (
	"github.com/codegangsta/cli"
	"gopkg.in/yaml.v2"
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

func TestGetConfig(t *testing.T) {
	test_file, err := os.Create("test_getConfig.yml")

	if err != nil {
		t.Error("Something went wrong creating the test file.")
	}

	expected := Config{
		Usage:   "Test example usage.",
		AhoyAPI: "v2",
		Version: "0.0.0",
		Commands: map[string]Command{
			"test-command": Command{
				Description: "Testing example Command.",
				Usage:       "test-command",
				Cmd:         "echo 'Hello World'",
				Hide:        false,
				Imports: []string{
					"./path/a",
					"./path/b",
				},
			},
		},
	}
	test_yaml, err := yaml.Marshal(expected)

	if err != nil {
		t.Error("Something went wrong mashelling the test object.")
	}

	test_file.Write([]byte(test_yaml))

	config, err := getConfig("test_getConfig.yml")

	if err != nil {
		t.Error("Something went wrong trying to load the config file.")
	}

	if config.Usage != expected.Usage {
		t.Errorf("Expected config.Usage to be %s, but actaul is %s", expected.Usage, config.Usage)
	}

	if config.Commands["test-command"].Cmd != expected.Commands["test-command"].Cmd {
		t.Errorf("Expected config.Commands['test-command'].cmd to be %s, but actaul is %s", expected.Commands["test-command"].Cmd, config.Commands["test-command"].Cmd)
	}

	test_file.Close()
	os.Remove("test_getConfig.yml")
}

func TestGetConfigPath(t *testing.T) {
	// Passinng empty string.
	pwd, _ := os.Getwd()
	expected := pwd + "/.ahoy.yml"
	actual, _ := getConfigPath("")
	if expected != actual {
		t.Errorf("ahoy docker override-example: expected - %s; actual - %s", string(expected), string(actual))
	}

	// TODO: use golang try-catch?
	// Passing bogus path..
	//_, err := getConfigPath("~/bogus/path")
	//if err == nil {
	//t.Error("getConfigPath fails on bogus path.")
	//}

	// Passing known path works as expected
	expected = pwd + "/.ahoy.yml"
	actual, _ = getConfigPath(expected)

	if expected != actual {
		t.Errorf("ahoy docker override-example: expected - %s; actual - %s", string(expected), string(actual))
	}

	// TODO: Passing directory should return default
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
