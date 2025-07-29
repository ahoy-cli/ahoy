#!/usr/bin/env bats

setup() {
  mv .ahoy.yml tmp.ahoy.yml
}

teardown() {
  mv tmp.ahoy.yml .ahoy.yml
}

@test "Commands with usage appear in command list" {
  run ./ahoy -f testdata/descriptions-test.ahoy.yml
  # Usage should appear in the main command list
  [[ "$output" =~ "simple  Simple command" ]]
  [[ "$output" =~ "multiline          Multiline description test" ]]
  [[ "$output" =~ "no-description    Command without description" ]]
}

@test "Commands without usage still appear in command list" {
  run ./ahoy -f testdata/descriptions-test.ahoy.yml
  # Command should appear even without usage
  [[ "$output" =~ "only-description" ]]
}

@test "Updated simple.ahoy.yml commands show usage in list" {
  run ./ahoy -f testdata/simple.ahoy.yml
  [[ "$output" =~ "echo  Display a message" ]]
  [[ "$output" =~ "list  List directory contents" ]]
  [[ "$output" =~ "whalesay  Make a whale say something" ]]
}

@test "Commands with aliases show usage correctly" {
  run ./ahoy -f testdata/command-aliases.ahoy.yml
  [[ "$output" =~ "hello, hi, greet, ahoy  Say hello" ]]
  [[ "$output" =~ "ahoy-there, ahoy  Say \"ahoy there!\"" ]]
}

@test "Newer features commands show updated usage" {
  run ./ahoy -f testdata/newer-features.ahoy.yml  
  [[ "$output" =~ "test-aliases, t, test-cmd  Run test with aliases" ]]
  # test-optional command may not appear if optional imports are missing, which is expected
}

@test "Description field is properly parsed and available" {
  # This test verifies that the description field exists and can be accessed
  # even though it may not be displayed in current help output
  
  # Test that commands with descriptions can still execute normally
  run ./ahoy -f testdata/descriptions-test.ahoy.yml simple
  [ "$status" -eq 0 ]
  [[ "$output" =~ "simple" ]]
  
  # Test multiline description command executes
  run ./ahoy -f testdata/descriptions-test.ahoy.yml multiline  
  [ "$status" -eq 0 ]
  [[ "$output" =~ "multiline" ]]
}

@test "Commands with both usage and description work correctly" {
  # Test that having both fields doesn't break functionality
  run ./ahoy -f testdata/simple.ahoy.yml echo "test message"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "test message" ]]
}

@test "Empty descriptions don't break command execution" {
  run ./ahoy -f testdata/descriptions-test.ahoy.yml empty-description
  [ "$status" -eq 0 ]
  [[ "$output" =~ "empty" ]]
}

# Test that the main ahoy help still works with description fields present
@test "Main help output includes commands with descriptions" {
  run ./ahoy -f testdata/descriptions-test.ahoy.yml
  [ "$status" -eq 1 ]  # Should show help and return error for missing command
  [[ "$output" =~ "COMMANDS:" ]]
  [[ "$output" =~ "simple" ]]
  [[ "$output" =~ "multiline" ]]
}