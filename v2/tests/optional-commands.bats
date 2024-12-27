#!/usr/bin/env bats

load 'test_helpers/bats-support/load'
load 'test_helpers/bats-assert/load'

@test "Optional import not found doesn't cause error" {
  # Run ahoy without arguments (which typically lists available commands)
  run ./ahoy -f testdata/optional-command.ahoy.yml

    # Check that the optional command is not listed
  [[ ! "$output" =~ "optional-cmd" ]]

  # Check that the regular command is listed
  [[ "$output" =~ "regular-cmd" ]]

  # Check that a standard "Missing argument" error is shown
  [[ "$output" =~ "Missing flag or argument" ]]

  # Check that the command exited with an error
  [ $status -eq 1 ]

  # Try to run the optional command (it should fail gracefully)
  run ./ahoy -f testdata/optional-command.ahoy.yml optional-cmd
  [ $status -eq 1 ]
  [[ "$output" =~ "Command not found for 'optional-cmd'" ]]

  # Run the regular command (it should work)
  run ./ahoy -f testdata/optional-command.ahoy.yml regular-cmd
  [ $status -eq 0 ]
  [[ "$output" = "This is a regular command" ]]
}

@test "Non-optional command with missing imports causes error" {
  # Run ahoy without arguments
  run ./ahoy -f testdata/non-optional-command.ahoy.yml

  # Check that the command failed
  [ $status -eq 1 ]

  # Check for the appropriate error message
  [[ "$output" =~ "Command [non-optional-cmd] has 'imports' set, but no commands were found" ]]
}
