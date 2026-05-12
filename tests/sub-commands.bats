#!/usr/bin/env bats

load 'test_helpers/bats-support/load'
load 'test_helpers/bats-assert/load'

setup() {
  # Create a temporary directory for our test files
  TEST_DIR="$(mktemp -d)"
  
  # Create main .ahoy.yml file
  cat > "${TEST_DIR}/.ahoy.yml" <<EOF
ahoyapi: v2
commands:
  main-cmd:
    usage: Main command
    imports:
      - sub1.yml
      - sub2.yml
      - non-existent.yml
EOF

  # Create sub1.yml
  cat > "${TEST_DIR}/sub1.yml" <<EOF
ahoyapi: v2
commands:
  sub1-cmd:
    usage: Subcommand 1
    cmd: echo "Subcommand 1"
EOF

  # Create sub2.yml
  cat > "${TEST_DIR}/sub2.yml" <<EOF
ahoyapi: v2
commands:
  sub2-cmd:
    usage: Subcommand 2
    cmd: echo "Subcommand 2"
EOF

  # Copy freshly built ahoy binary to the test directory
  cp ahoy "${TEST_DIR}"

  # Change to the test directory
  cd "${TEST_DIR}"
}

teardown() {
  # Remove the temporary directory
  rm -rf "${TEST_DIR}"
}

@test "getSubCommands() loads existing subcommands" {
  run ./ahoy

  # Check that main-cmd is listed
  [[ "$output" =~ "Main command" ]]

  # Check that the command executed with an expected error
  [ "$status" -eq 1 ]

  # Check that sub1-cmd and sub2-cmd are listed under main-cmd
  run ./ahoy main-cmd --help
  [ "$status" -eq 0 ]
  [[ "$output" =~ "sub1-cmd" ]]
  [[ "$output" =~ "sub2-cmd" ]]

  # Check that non-existent.yml is ignored without error
  [[ ! "$output" =~ "non-existent.yml" ]]

  # Check that the command executed successfully
  [ "$status" -eq 0 ]
}

@test "getSubCommands() executes subcommands correctly" {
  # Test sub1-cmd
  run ./ahoy main-cmd sub1-cmd
  [ "$status" -eq 0 ]
  [ "$output" = "Subcommand 1" ]

  # Test sub2-cmd
  run ./ahoy main-cmd sub2-cmd
  [ "$status" -eq 0 ]
  [ "$output" = "Subcommand 2" ]
}

@test "getSubCommands() handles empty imports" {
  # Create an empty.ahoy.yml with empty imports
  cat > empty.ahoy.yml <<EOF
ahoyapi: v2
commands:
  empty-import:
    usage: Command with empty imports
    imports: []
EOF

  run ./ahoy -f empty.ahoy.yml
  [[ "$output" =~ "empty-import" ]]

  # Check that the command executed with an expected error
  [ "$status" -eq 1 ]

  run ./ahoy -f empty.ahoy.yml empty-import
  
  [[ "$output" =~ "empty-import" ]]
  
  [[ "$output" =~ "but it is empty" ]]
  
  # Check that the command executed with an expected error
  [ "$status" -eq 1 ]
}
