package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"gopkg.in/yaml.v2"
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
	AhoyConf.srcDir = ""

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

	yamlConfigA := `
ahoyapi: v2
commands:
  test-command:
    description: Testing example Command.
    usage: test-command a
    cmd: echo "test"
    hide: false
`
	yamlConfigB := `
ahoyapi: v2
commands:
  test-command:
    description: Testing example Command.
    usage: test-command b
    cmd: echo "test"
    hide: false
`
	_, err = file1.Write([]byte(yamlConfigA))
	if err != nil {
		t.Error("Error writing to file1.")
	}

	_, err = file2.Write([]byte(yamlConfigB))
	if err != nil {
		t.Error("Error writing to file2.")
	}

	actual = getSubCommands([]string{
		"./testing/a.ahoy.yml",
		"./testing/b.ahoy.yml",
	})

	if len(actual) != 1 {
		t.Error("Sourcedir:", AhoyConf.srcDir)
		t.Error("Failed: expect that two commands with the same name get merged into one.", actual)
	}

	if len(actual) > 0 && actual[0].Usage != "test-command b" {
		t.Error("Failed: expect that when multiple commands are merged, last one wins.", actual)
	}

	// Test commands with different names do not get merged.
	file3, err := os.Create("testing/c.ahoy.yml")
	if err != nil {
		t.Error("Something went wrong with the file creation - file3.")
	}

	//logger("fatal", "test")
	yamlConfigC := `
ahoyapi: v2
commands:
  test-new-command:
    description: Testing new example Command.
    usage: test-new-command a
    cmd: "echo new a.ahoy.yml"
    hide: false
`
	_, err = file3.Write([]byte(yamlConfigC))
	if err != nil {
		t.Error("Error writing to file3.")
	}

	actual = getSubCommands([]string{
		"./testing/a.ahoy.yml",
		"./testing/b.ahoy.yml",
		"./testing/c.ahoy.yml",
	})

	if len(actual) != 2 {
		fmt.Printf("x = %#v \n", actual)
		t.Error("Failed: expect unique commands to be captured separately.", "commands found", actual)
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

func TestGetConfigPathErrorOnBogusPath(t *testing.T) {
	_, err := getConfigPath("~/bogus/path")
	if err == nil {
		t.Error("getConfigPath did not fail when passed a bogus path.")
	}
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
