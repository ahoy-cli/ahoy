#!/usr/bin/env bats

setup() {
  mv .ahoy.yml tmp.ahoy.yml
}

teardown() {
  mv tmp.ahoy.yml .ahoy.yml
}

@test "Commands with usage appear in command list" {
  run ./ahoy -f testdata/descriptions-test.ahoy.yml
  # Usage should appear in the main command list.
  [[ "$output" =~ "simple" ]] && [[ "$output" =~ "Simple command" ]]
  [[ "$output" =~ "multiline" ]] && [[ "$output" =~ "Multiline description test" ]]
  [[ "$output" =~ "no-description" ]] && [[ "$output" =~ "Command without description" ]]
}

@test "Commands without usage still appear in command list" {
  run ./ahoy -f testdata/descriptions-test.ahoy.yml
  # Command should appear even without usage.
  [[ "$output" =~ "only-description" ]]
}

@test "Descriptions are not shown in the main command list" {
  run ./ahoy -f testdata/descriptions-test.ahoy.yml
  # The long description text should not appear in the top-level listing.
  [[ ! "$output" =~ "A basic command with a simple description" ]]
  [[ ! "$output" =~ "Multiple paragraphs" ]]
  # But the hint to use --help should appear.
  [[ "$output" =~ "Use 'ahoy <command> --help' for detailed information about a command" ]]
}

@test "Updated simple.ahoy.yml commands show usage in list" {
  run ./ahoy -f testdata/simple.ahoy.yml
  [[ "$output" =~ "echo" ]] && [[ "$output" =~ "Display a message" ]]
  [[ "$output" =~ "list" ]] && [[ "$output" =~ "List directory contents" ]]
  [[ "$output" =~ "whalesay" ]] && [[ "$output" =~ "Make a whale say something" ]]
}

@test "Commands with aliases show usage correctly" {
  run ./ahoy -f testdata/command-aliases.ahoy.yml
  [[ "$output" =~ "hello, hi, greet, ahoy" ]] && [[ "$output" =~ "Say hello" ]]
  [[ "$output" =~ "ahoy-there, ahoy" ]] && [[ "$output" =~ "Say \"ahoy there!\"" ]]
}

@test "Newer features commands show updated usage" {
  run ./ahoy -f testdata/newer-features.ahoy.yml
  [[ "$output" =~ "test-aliases, t, test-cmd" ]] && [[ "$output" =~ "Run test with aliases" ]]
  # test-optional command may not appear if optional imports are missing, which is expected.
}

@test "Description field is properly parsed and available" {
  # Test that commands with descriptions can still execute normally.
  run ./ahoy -f testdata/descriptions-test.ahoy.yml simple
  [ "$status" -eq 0 ]
  [[ "$output" =~ "simple" ]]

  # Test multiline description command executes.
  run ./ahoy -f testdata/descriptions-test.ahoy.yml multiline
  [ "$status" -eq 0 ]
  [[ "$output" =~ "multiline" ]]
}

@test "Per-command help shows description" {
  # A command with a simple description should show it in DESCRIPTION section.
  run ./ahoy -f testdata/descriptions-test.ahoy.yml simple --help
  [ "$status" -eq 0 ]
  [[ "$output" =~ "DESCRIPTION:" ]]
  [[ "$output" =~ "A basic command with a simple description" ]]
}

@test "Per-command help shows multiline description" {
  run ./ahoy -f testdata/descriptions-test.ahoy.yml multiline --help
  [ "$status" -eq 0 ]
  [[ "$output" =~ "DESCRIPTION:" ]]
  [[ "$output" =~ "Multiple paragraphs" ]]
  [[ "$output" =~ "Bullet points" ]]
}

@test "Per-command help without description omits DESCRIPTION section" {
  run ./ahoy -f testdata/descriptions-test.ahoy.yml no-description --help
  [ "$status" -eq 0 ]
  [[ ! "$output" =~ "DESCRIPTION:" ]]
  [[ "$output" =~ "NAME:" ]]
  [[ "$output" =~ "USAGE:" ]]
}

@test "Per-command help shows aliases" {
  run ./ahoy -f testdata/command-aliases.ahoy.yml hello --help
  [ "$status" -eq 0 ]
  [[ "$output" =~ "ALIASES:" ]]
  [[ "$output" =~ "hi, greet, ahoy" ]]
}

@test "Commands with both usage and description work correctly" {
  # Test that having both fields doesn't break functionality.
  run ./ahoy -f testdata/simple.ahoy.yml echo "test message"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "test message" ]]
}

@test "Empty descriptions don't break command execution" {
  run ./ahoy -f testdata/descriptions-test.ahoy.yml empty-description
  [ "$status" -eq 0 ]
  [[ "$output" =~ "empty" ]]
}

@test "Per-command --help shows help and does not execute the command" {
  # Passing --help to a command should show help, not run the underlying cmd.
  run ./ahoy -f testdata/simple.ahoy.yml echo --help
  [ "$status" -eq 0 ]
  [[ "$output" =~ "NAME:" ]]
  [[ "$output" =~ "USAGE:" ]]
  [[ ! "$output" =~ "test message" ]]
}

@test "Help flag after -- separator is passed through to the command" {
  # Arguments after -- should be passed to the underlying command, not intercepted.
  run ./ahoy -f testdata/simple.ahoy.yml echo -- --help
  [ "$status" -eq 0 ]
  [[ "$output" =~ "--help" ]]
}

# Test that the main ahoy help still works with description fields present.
@test "Main help output includes commands with descriptions" {
  run ./ahoy -f testdata/descriptions-test.ahoy.yml
  [ "$status" -eq 1 ]  # Should show help and return error for missing command.
  [[ "$output" =~ "COMMANDS:" ]]
  [[ "$output" =~ "simple" ]]
  [[ "$output" =~ "multiline" ]]
}
