package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// writeCompletionFixture writes an ahoy file with a plain command, a hidden
// command and a command that imports subcommands, and returns its path.
func writeCompletionFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		".ahoy.yml": `ahoyapi: v2
commands:
  leaf:
    usage: A plain command
    cmd: echo leaf
    aliases: ["lf"]
  secret:
    hide: true
    cmd: echo secret
  group:
    usage: A command with subcommands
    imports:
      - group.ahoy.yml
`,
		"group.ahoy.yml": `ahoyapi: v2
commands:
  one:
    cmd: echo one
  two:
    cmd: echo two
`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(dir, ".ahoy.yml")
}

func TestInitFlagsTrailingCompletionFlag(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantFlag    bool
		wantCmdArgs []string
	}{
		{"before the command", []string{"--generate-bash-completion"}, true, []string{}},
		{"after a command", []string{"leaf", "--generate-bash-completion"}, true, []string{"leaf"}},
		{"after a partial word", []string{"leaf", "x", "--generate-bash-completion"}, true, []string{"leaf", "x"}},
		{"after the file flag", []string{"-f", "a.yml", "group", "--generate-bash-completion"}, true, []string{"group"}},
		{"not last belongs to the command", []string{"leaf", "--generate-bash-completion", "x"}, false, []string{"leaf", "--generate-bash-completion", "x"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newAppState()
			s.initFlags(tt.args)
			if s.bashCompletionFlagSet != tt.wantFlag {
				t.Errorf("bashCompletionFlagSet = %v, want %v", s.bashCompletionFlagSet, tt.wantFlag)
			}
			if !reflect.DeepEqual(s.commandArgs, tt.wantCmdArgs) {
				t.Errorf("commandArgs = %q, want %q", s.commandArgs, tt.wantCmdArgs)
			}
		})
	}
}

func TestPrintLegacyCompletions(t *testing.T) {
	rootCmd := newAppState().setupApp([]string{"-f", writeCompletionFixture(t)})

	t.Run("root lists visible commands", func(t *testing.T) {
		var out bytes.Buffer
		printLegacyCompletions(&out, rootCmd, nil)
		got := strings.Fields(out.String())
		for _, want := range []string{"leaf", "group"} {
			if !contains(got, want) {
				t.Errorf("expected %q in %q", want, got)
			}
		}
		if contains(got, "secret") {
			t.Errorf("hidden command listed: %q", got)
		}
	})

	tests := []struct {
		name string
		args []string
		want string
	}{
		{"plain command lists nothing", []string{"leaf"}, ""},
		{"plain command with a partial word lists nothing", []string{"leaf", "x"}, ""},
		{"alias resolves", []string{"lf"}, ""},
		{"imports list subcommands", []string{"group"}, "one\ntwo\n"},
		{"imports with a partial word list subcommands", []string{"group", "o"}, "one\ntwo\n"},
		{"unknown command lists nothing", []string{"nope"}, ""},
		{"imports with unresolved words list nothing", []string{"group", "typo", "o"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			printLegacyCompletions(&out, rootCmd, tt.args)
			if out.String() != tt.want {
				t.Errorf("got %q, want %q", out.String(), tt.want)
			}
		})
	}
}

func TestBashCompleteOnlyAddsAliases(t *testing.T) {
	s := newAppState()
	rootCmd := s.setupApp([]string{"-f", writeCompletionFixture(t)})

	tests := []struct {
		name       string
		args       []string
		toComplete string
		want       []string
	}{
		{"empty prefix", nil, "", []string{"lf"}},
		{"matching prefix", nil, "l", []string{"lf"}},
		{"non-matching prefix", nil, "g", []string{}},
		{"past the first argument", []string{"leaf"}, "", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := s.bashComplete(rootCmd, tt.args, tt.toComplete)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func contains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}
