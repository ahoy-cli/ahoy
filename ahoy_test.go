package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"testing"
)

func TestOverrideExample(t *testing.T) {
	expected := "Overrode you.\n"
	actual, _ := appRun([]string{"ahoy", "-f", "testdata/override-base.ahoy.yml", "docker", "override-example"})
	if expected != actual {
		t.Errorf("ahoy docker override-example: expected - %s; actual - %s", string(expected), string(actual))
	}
}

func TestGetCommands(t *testing.T) {
	// Get Command with no sub Commands.
	config := Config{
		Usage:   "Test getSubCommands Usage.",
		AhoyAPI: "v2",
		Commands: map[string]Command{
			"test-command": Command{
				Description: "Testing example Command.",
				Usage:       "test-command a",
				Cmd:         "echo a.ahoy.yml",
				Hide:        false,
			},
		},
	}

	commands := getCommands(config)

	if len(commands) != 1 {
		t.Error("Expect that getCommands can get one command if passed config with one command.")
	}
}

func TestGetSubCommand(t *testing.T) {
	// Since we're not running the app directly, sourcedir doesn't get reset, so
	// we need to reset it ourselves. TODO: Remove these globals somehow.
	sourcedir = ""

	// When empty return empty list of commands.

	actual := getSubCommands([]string{})

	if len(actual) != 0 {
		t.Error("Expect that getSubCommands([]string) returns []Command{}")
	}

	// List of bogus or empty strings returns empty list of commands.
	actual = getSubCommands([]string{
		"./testing/bogus1.ahoy.yml",
		"./testing/private.ahoy.yml",
	})

	if len(actual) != 0 {
		t.Error("Expect that getSubCommands([]string) returns []Command{}")
	}

	// Commands with same name are merged, last one wins.
	err := os.MkdirAll("testing", 0755)
	if err != nil {
		t.Error("Something went wrong creating the 'testing' directory")
	}

	file1, err := os.Create("testing/a.ahoy.yml")
	if err != nil {
		t.Error("Something went wrong with the file creation - file1.")
	}

	file2, err := os.Create("testing/b.ahoy.yml")
	if err != nil {
		t.Error("Something went wrong with the file creation - file2.")
	}

	config := Config{
		Usage:   "Test getSubCommands Usage.",
		AhoyAPI: "v2",
		Commands: map[string]Command{
			"test-command": Command{
				Description: "Testing example Command.",
				Usage:       "test-command a",
				Cmd:         "echo a.ahoy.yml",
				Hide:        false,
			},
		},
	}

	yamlConfig, err := yaml.Marshal(config)
	if err != nil {
		t.Error("Error marshalling config for file1")
	}

	_, err = file1.Write([]byte(yamlConfig))

	if err != nil {
		t.Error("Error writing to file1.")
	}

	command := config.Commands["test-command"]
	command.Usage = "testing-command b"
	config.Commands["test-command"] = command

	yamlConfig, err = yaml.Marshal(config)
	if err != nil {
		t.Error("Error marshalling config for file2")
	}

	_, err = file2.Write([]byte(yamlConfig))

	if err != nil {
		t.Error("Error writing to file2.")
	}

	actual = getSubCommands([]string{
		"./testing/a.ahoy.yml",
		"./testing/b.ahoy.yml",
	})

	if len(actual) != 1 {
		t.Error("Sourcedir:", sourcedir)
		t.Error("Failed: expect that two commands with the same name get merged into one.", actual)
	}

	if len(actual) > 0 && actual[0].Usage != "testing-command b" {
		t.Error("Failed: expect that when multiple commands are merged, last one wins.")
	}

	// Test commands with different names do not get merged.
	file3, err := os.Create("testing/c.ahoy.yml")
	if err != nil {
		t.Error("Something went wrong with the file creation - file3.")
	}

	config.Commands["testing-new-command"] = Command{
		Description: "Testing new example Command.",
		Usage:       "test-new-command a",
		Cmd:         "echo new a.ahoy.yml",
		Hide:        false,
	}

	yamlConfig, err = yaml.Marshal(config)
	if err != nil {
		t.Error("Error marshalling config for file3")
	}

	_, err = file3.Write([]byte(yamlConfig))

	if err != nil {
		t.Error("Error writing to file3.")
	}

	actual = getSubCommands([]string{
		"./testing/a.ahoy.yml",
		"./testing/b.ahoy.yml",
		"./testing/c.ahoy.yml",
	})

	if len(actual) != 2 {
		t.Error("Failed: expect unique commands to be captured separately.")
	}

	file1.Close()
	file2.Close()
	file3.Close()
	os.RemoveAll("testing")
}

func TestGetConfig(t *testing.T) {
	testFile, err := os.Create("test_getConfig.yml")

	if err != nil {
		t.Error("Something went wrong creating the test file.")
	}

	expected := Config{
		Usage:   "Test example usage.",
		AhoyAPI: "v2",
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
	testYaml, err := yaml.Marshal(expected)

	if err != nil {
		t.Error("Something went wrong mashelling the test object.")
	}

	testFile.Write([]byte(testYaml))

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

	testFile.Close()
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

	// Passing known path works as expected
	expected = pwd + "/.ahoy.yml"
	actual, _ = getConfigPath(expected)

	if expected != actual {
		t.Errorf("ahoy docker override-example: expected - %s; actual - %s", string(expected), string(actual))
	}

	// TODO: Passing directory should return default
}

func TestGetConfigPathPanicOnBogusPath(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("getConfigPath did not fail when passed a bogus path.")
		}
	}()

	getConfigPath("~/bogus/path")
}

func appRun(args []string) (string, error) {
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	setupApp(args[1:])
	app.Run(args)

	w.Close()
	//@aashil thinks this reads from the command line
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	return string(out), nil
}
