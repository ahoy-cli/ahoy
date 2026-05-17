package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestOverrideExample(t *testing.T) {
	// Override a command with the same command from another imported command file.
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
			"test-command": {
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
	err := os.MkdirAll("testing", 0o755)
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

	if len(actual) > 0 && actual[0].Short != "test-command b" {
		t.Error("Failed: expect that when multiple commands are merged, last one wins.", actual)
	}

	// Test commands with different names do not get merged.
	file3, err := os.Create("testing/c.ahoy.yml")
	if err != nil {
		t.Error("Something went wrong with the file creation - file3.")
	}

	// logger("fatal", "test")
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
	// Get a config file.
	testFile, err := os.Create("test_getConfig.yml")
	if err != nil {
		t.Error("Something went wrong creating the test file.")
	}

	expected := Config{
		Usage:   "Test example usage.",
		AhoyAPI: "v2",
		Commands: map[string]Command{
			"test-command": {
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
		t.Error("Something went wrong marshalling the test object.")
	}

	testFile.Write([]byte(testYaml))

	config, err := getConfig("test_getConfig.yml")
	if err != nil {
		t.Error("Something went wrong trying to load the config file.")
	}

	if config.Usage != expected.Usage {
		t.Errorf("Expected config.Usage to be %s, but actual is %s", expected.Usage, config.Usage)
	}

	if config.Commands["test-command"].Cmd != expected.Commands["test-command"].Cmd {
		t.Errorf("Expected config.Commands['test-command'].cmd to be %s, but actual is %s", expected.Commands["test-command"].Cmd, config.Commands["test-command"].Cmd)
	}

	testFile.Close()
	os.Remove("test_getConfig.yml")
}

func TestGetConfigPath(t *testing.T) {
	// Passing an empty string.
	pwd, _ := os.Getwd()
	expected := filepath.Join(pwd, ".ahoy.yml")
	actual, _ := getConfigPath("")
	if expected != actual {
		t.Errorf("ahoy docker override-example: expected - %s; actual - %s", string(expected), string(actual))
	}

	// Passing known path works as expected
	expected = filepath.Join(pwd, ".ahoy.yml")
	actual, _ = getConfigPath(expected)

	if expected != actual {
		t.Errorf("ahoy docker override-example: expected - %s; actual - %s", string(expected), string(actual))
	}

	// TODO: Passing directory should return default
}

func TestGetConfigPathErrorOnBogusPath(t *testing.T) {
	// Test getting a bogus config path.
	_, err := getConfigPath("~/bogus/path")
	if err == nil {
		t.Error("getConfigPath did not fail when passed a bogus path.")
	}
}

func appRun(args []string) (string, error) {
	stdout := os.Stdout
	stderr := os.Stderr
	r, w, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = wErr

	cmd := setupApp(args[1:])
	// Don't call SetArgs again - setupApp already parsed the flags
	// Just set the args to the command args (after flags)

	// Find where the command starts (after all flags)
	cmdArgs := []string{}
	skipNext := false
	for i, arg := range args[1:] {
		if skipNext {
			skipNext = false
			continue
		}
		if arg == "-f" || arg == "--file" {
			skipNext = true
			continue
		}
		if arg == "-v" || arg == "--verbose" {
			continue
		}
		// This is a command or command argument
		cmdArgs = append(cmdArgs, args[1+i:]...)
		break
	}

	cmd.SetArgs(cmdArgs)
	cmd.Execute()

	w.Close()
	wErr.Close()
	out, _ := io.ReadAll(r)
	errOut, _ := io.ReadAll(rErr)
	os.Stdout = stdout
	os.Stderr = stderr

	// If there was an error output, include it
	if len(errOut) > 0 {
		return string(out), fmt.Errorf("%s", errOut)
	}
	return string(out), nil
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		path     string
		baseDir  string
		expected string
	}{
		// Absolute paths returned as-is.
		{"/absolute/path", "/base", "/absolute/path"},
		// Tilde expanded to home directory.
		{"~/mydir", "/base", filepath.Join(home, "mydir")},
		{"~/.ahoy.yml", "/base", filepath.Join(home, ".ahoy.yml")},
		// Relative paths joined with base directory.
		{"relative/path", "/base", filepath.Join("/base", "relative/path")},
		{".env", "/some/dir", filepath.Join("/some/dir", ".env")},
	}

	for _, tt := range tests {
		result := expandPath(tt.path, tt.baseDir)
		if result != tt.expected {
			t.Errorf("expandPath(%q, %q) = %q, want %q", tt.path, tt.baseDir, result, tt.expected)
		}
	}
}
