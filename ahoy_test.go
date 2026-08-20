package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestOverrideExample(t *testing.T) {
	// Override a command with the same command from another imported command file.
	expected := "Overrode you.\n"
	actual, _ := appRun([]string{"ahoy", "-f", "testdata/override-base.ahoy.yml", "docker", "override-example"})
	if expected != actual {
		t.Errorf("ahoy docker override-example: expected - %s; actual - %s", expected, actual)
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

	commands := (&appState{}).getCommands(config)

	if len(commands) != 1 {
		t.Error("Expect that getCommands can get one command if passed config with one command.")
	}
}

func TestGetSubCommand(t *testing.T) {
	// Each scenario uses its own appState, no global save/restore needed.
	state := &appState{}

	// When empty return empty list of commands.
	actual := state.getSubCommands([]string{})
	if len(actual) != 0 {
		t.Error("Expect that getSubCommands([]string) returns []Command{}")
	}

	// List of bogus or empty strings returns empty list of commands.
	actual = state.getSubCommands([]string{
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

	actual = state.getSubCommands([]string{
		"./testing/a.ahoy.yml",
		"./testing/b.ahoy.yml",
	})

	if len(actual) != 1 {
		t.Error("Sourcedir:", state.srcDir)
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

	// Fresh state so importVisited doesn't carry over.
	state2 := &appState{}
	actual = state2.getSubCommands([]string{
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

	if _, err = testFile.Write(testYaml); err != nil {
		t.Fatal("Something went wrong writing the test yaml file.")
	}

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
	// Passing an empty string (no sourcefile set) finds .ahoy.yml in cwd.
	pwd, _ := os.Getwd()
	expected := filepath.Join(pwd, ".ahoy.yml")
	actual, _ := (&appState{}).getConfigPath()
	if expected != actual {
		t.Errorf("ahoy docker override-example: expected - %s; actual - %s", expected, actual)
	}

	// Passing known path works as expected.
	expected = filepath.Join(pwd, ".ahoy.yml")
	actual, _ = (&appState{sourcefile: expected}).getConfigPath()

	if expected != actual {
		t.Errorf("ahoy docker override-example: expected - %s; actual - %s", expected, actual)
	}

	// TODO: Passing directory should return default
}

func TestGetConfigPathErrorOnBogusPath(t *testing.T) {
	// Test getting a bogus config path.
	_, err := (&appState{sourcefile: "~/bogus/path"}).getConfigPath()
	if err == nil {
		t.Error("getConfigPath did not fail when passed a bogus path.")
	}
}

func TestGetConfigPathReturnsErrNoConfigWhenNotFound(t *testing.T) {
	// Change to a temp directory outside the repo tree so the walk-up never
	// finds a .ahoy.yml and getConfigPath must return errNoConfig.
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})

	_, gotErr := (&appState{}).getConfigPath()
	if !errors.Is(gotErr, errNoConfig) {
		t.Errorf("expected errNoConfig, got %v", gotErr)
	}
}

func appRun(args []string) (string, error) {
	stdout := os.Stdout
	stderr := os.Stderr

	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", pipeErr)
	}
	defer r.Close()

	rErr, wErr, pipeErr := os.Pipe()
	if pipeErr != nil {
		w.Close()
		return "", fmt.Errorf("failed to create stderr pipe: %w", pipeErr)
	}
	defer rErr.Close()

	os.Stdout = w
	os.Stderr = wErr
	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
	}()

	cmd := newAppState().setupApp(args[1:])
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
		if strings.HasPrefix(arg, "--file=") || strings.HasPrefix(arg, "-f=") {
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
	execErr := cmd.Execute()

	w.Close()
	wErr.Close()
	out, _ := io.ReadAll(r)
	errOut, _ := io.ReadAll(rErr)

	// Report a failure if either the command itself returned an error or
	// anything was written to stderr.
	if execErr != nil && len(errOut) > 0 {
		return string(out), fmt.Errorf("%w: %s", execErr, errOut)
	}
	if execErr != nil {
		return string(out), execErr
	}
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

func TestMultiBranchAndCircularImports(t *testing.T) {
	err := os.MkdirAll("test_imports", 0755)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("test_imports")

	// Create a shared command file
	sharedYml := `
ahoyapi: v2
commands:
  shared-cmd:
    cmd: echo "shared"
`
	if err := os.WriteFile("test_imports/shared.yml", []byte(sharedYml), 0644); err != nil {
		t.Fatal(err)
	}

	// Create branch A which imports shared
	branchAYml := `
ahoyapi: v2
commands:
  import-a:
    imports:
      - shared.yml
`
	if err := os.WriteFile("test_imports/branchA.yml", []byte(branchAYml), 0644); err != nil {
		t.Fatal(err)
	}

	// Create branch B which imports shared
	branchBYml := `
ahoyapi: v2
commands:
  import-b:
    imports:
      - shared.yml
`
	if err := os.WriteFile("test_imports/branchB.yml", []byte(branchBYml), 0644); err != nil {
		t.Fatal(err)
	}

	// Create circular A which imports circular B
	circularAYml := `
ahoyapi: v2
commands:
  import-circ-a:
    optional: true
    imports:
      - circularB.yml
  circ-a:
    cmd: echo "circ-a"
`
	if err := os.WriteFile("test_imports/circularA.yml", []byte(circularAYml), 0644); err != nil {
		t.Fatal(err)
	}

	// Create circular B which imports circular A
	circularBYml := `
ahoyapi: v2
commands:
  import-circ-b:
    optional: true
    imports:
      - circularA.yml
  circ-b:
    cmd: echo "circ-b"
`
	if err := os.WriteFile("test_imports/circularB.yml", []byte(circularBYml), 0644); err != nil {
		t.Fatal(err)
	}

	origLogOutput := log.Writer()
	t.Cleanup(func() {
		log.SetOutput(origLogOutput)
	})

	// Test multi-branch imports. Both branchA and branchB should successfully resolve shared.yml.
	state := &appState{
		srcDir:        "test_imports",
		importVisited: map[string]bool{normalizePath("test_imports/root.yml"): true},
	}

	commands := state.getSubCommands([]string{
		"branchA.yml",
		"branchB.yml",
	})

	foundSharedCmdInA := false
	foundSharedCmdInB := false
	for _, cmd := range commands {
		if cmd.Name() == "import-a" {
			for _, sub := range cmd.Commands() {
				if sub.Name() == "shared-cmd" {
					foundSharedCmdInA = true
				}
			}
		}
		if cmd.Name() == "import-b" {
			for _, sub := range cmd.Commands() {
				if sub.Name() == "shared-cmd" {
					foundSharedCmdInB = true
				}
			}
		}
	}
	if !foundSharedCmdInA || !foundSharedCmdInB {
		t.Errorf("Expected to find 'shared-cmd' in both import-a (found: %v) and import-b (found: %v)", foundSharedCmdInA, foundSharedCmdInB)
	}

	// Test circular imports to make sure they are caught and do not stack overflow.
	circState := &appState{
		srcDir:        "test_imports",
		importVisited: map[string]bool{normalizePath("test_imports/root.yml"): true},
	}

	// Capturing log/stdout to verify circular import warning is printed.
	// The original log output is restored by the t.Cleanup above.
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)

	circularCmds := circState.getSubCommands([]string{
		"circularA.yml",
	})

	// Since circularA imports circularB, and circularB imports circularA:
	// Path: circularA -> circularB -> circularA (detected circular, skipped)
	// We should still get the commands from circularA and circularB once.
	hasCircA := false
	hasCircB := false
	for _, cmd := range circularCmds {
		if cmd.Name() == "circ-a" {
			hasCircA = true
		}
		if cmd.Name() == "import-circ-a" {
			for _, sub := range cmd.Commands() {
				if sub.Name() == "circ-b" {
					hasCircB = true
				}
			}
		}
	}
	if !hasCircA || !hasCircB {
		t.Errorf("Expected to load circ-a (found: %v) and circ-b (found: %v)", hasCircA, hasCircB)
	}

	logStr := logBuf.String()
	if !strings.Contains(logStr, "Circular import detected") {
		t.Errorf("Expected log to contain 'Circular import detected', got: %q", logStr)
	}
}
