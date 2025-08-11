package main

import (
	"strings"
	"testing"

	"github.com/urfave/cli"
)

func TestDescriptionParsing(t *testing.T) {
	// Test that descriptions are parsed correctly from YAML
	config, err := getConfig("testdata/descriptions-test.ahoy.yml")
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	tests := []struct {
		command     string
		expectUsage string
		expectDesc  string
	}{
		{
			command:     "simple",
			expectUsage: "Simple command",
			expectDesc:  "A basic command with a simple description.",
		},
		{
			command:     "multiline",
			expectUsage: "Multiline description test",
			expectDesc:  "This command demonstrates multiline descriptions.\n\nIt includes:\n- Multiple paragraphs\n- Bullet points\n- Detailed explanations\n\nUse this to test that YAML multiline strings work correctly.\n",
		},
		{
			command:     "no-description",
			expectUsage: "Command without description",
			expectDesc:  "",
		},
		{
			command:     "only-description",
			expectUsage: "",
			expectDesc:  "This command has only a description but no usage field.",
		},
		{
			command:     "empty-description",
			expectUsage: "Command with empty description",
			expectDesc:  "",
		},
	}

	for _, test := range tests {
		t.Run(test.command, func(t *testing.T) {
			cmd, exists := config.Commands[test.command]
			if !exists {
				t.Fatalf("Command %s not found in config", test.command)
			}

			if cmd.Usage != test.expectUsage {
				t.Errorf("Usage mismatch for %s: expected %q, got %q", test.command, test.expectUsage, cmd.Usage)
			}

			if cmd.Description != test.expectDesc {
				t.Errorf("Description mismatch for %s: expected %q, got %q", test.command, test.expectDesc, cmd.Description)
			}
		})
	}
}

func TestDescriptionInCLICommands(t *testing.T) {
	// Test that descriptions are properly assigned to CLI commands
	config, err := getConfig("testdata/descriptions-test.ahoy.yml")
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	commands := getCommands(config)

	// Create a map for easy lookup
	cmdMap := make(map[string]cli.Command)
	for _, cmd := range commands {
		cmdMap[cmd.Name] = cmd
	}

	tests := []struct {
		command     string
		expectUsage string
		expectDesc  string
	}{
		{
			command:     "simple",
			expectUsage: "Simple command",
			expectDesc:  "A basic command with a simple description.",
		},
		{
			command:     "multiline",
			expectUsage: "Multiline description test",
			expectDesc:  "This command demonstrates multiline descriptions.\n\nIt includes:\n- Multiple paragraphs\n- Bullet points\n- Detailed explanations\n\nUse this to test that YAML multiline strings work correctly.\n",
		},
		{
			command:     "no-description",
			expectUsage: "Command without description",
			expectDesc:  "",
		},
		{
			command:     "only-description",
			expectUsage: "",
			expectDesc:  "This command has only a description but no usage field.",
		},
	}

	for _, test := range tests {
		t.Run(test.command, func(t *testing.T) {
			cmd, exists := cmdMap[test.command]
			if !exists {
				t.Fatalf("CLI command %s not found", test.command)
			}

			if cmd.Usage != test.expectUsage {
				t.Errorf("CLI Usage mismatch for %s: expected %q, got %q", test.command, test.expectUsage, cmd.Usage)
			}

			if cmd.Description != test.expectDesc {
				t.Errorf("CLI Description mismatch for %s: expected %q, got %q", test.command, test.expectDesc, cmd.Description)
			}
		})
	}
}

func TestDescriptionWithExistingCommands(t *testing.T) {
	// Test descriptions in existing test files
	testFiles := []struct {
		file     string
		command  string
		hasUsage bool
		hasDesc  bool
	}{
		{
			file:     "testdata/simple.ahoy.yml",
			command:  "echo",
			hasUsage: true,
			hasDesc:  true,
		},
		{
			file:     "testdata/command-aliases.ahoy.yml",
			command:  "hello",
			hasUsage: true,
			hasDesc:  true,
		},
		{
			file:     "testdata/newer-features.ahoy.yml",
			command:  "test-aliases",
			hasUsage: true,
			hasDesc:  true,
		},
	}

	for _, test := range testFiles {
		t.Run(test.file+"_"+test.command, func(t *testing.T) {
			config, err := getConfig(test.file)
			if err != nil {
				t.Fatalf("Failed to load %s: %v", test.file, err)
			}

			cmd, exists := config.Commands[test.command]
			if !exists {
				t.Fatalf("Command %s not found in %s", test.command, test.file)
			}

			if test.hasUsage && cmd.Usage == "" {
				t.Errorf("Expected %s to have usage field", test.command)
			}

			if test.hasDesc && cmd.Description == "" {
				t.Errorf("Expected %s to have description field", test.command)
			}

			// Verify CLI command assignment
			commands := getCommands(config)
			var cliCmd *cli.Command
			for _, c := range commands {
				if c.Name == test.command {
					cliCmd = &c
					break
				}
			}

			if cliCmd == nil {
				t.Fatalf("CLI command %s not found", test.command)
			}

			if test.hasUsage && cliCmd.Usage != cmd.Usage {
				t.Errorf("CLI Usage not assigned correctly for %s: expected %q, got %q", test.command, cmd.Usage, cliCmd.Usage)
			}

			if test.hasDesc && cliCmd.Description != cmd.Description {
				t.Errorf("CLI Description not assigned correctly for %s: expected %q, got %q", test.command, cmd.Description, cliCmd.Description)
			}
		})
	}
}

func TestMultilineDescriptionFormatting(t *testing.T) {
	// Test that multiline descriptions preserve formatting
	config, err := getConfig("testdata/descriptions-test.ahoy.yml")
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	cmd := config.Commands["multiline"]
	if !strings.Contains(cmd.Description, "Multiple paragraphs") {
		t.Error("Multiline description should contain 'Multiple paragraphs'")
	}

	if !strings.Contains(cmd.Description, "- Bullet points") {
		t.Error("Multiline description should contain bullet points")
	}

	// Count newlines to ensure multiline structure is preserved
	newlineCount := strings.Count(cmd.Description, "\n")
	if newlineCount < 3 {
		t.Errorf("Expected at least 3 newlines in multiline description, got %d", newlineCount)
	}
}
