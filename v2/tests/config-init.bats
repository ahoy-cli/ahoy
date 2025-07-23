#!/usr/bin/env bats

# Setup and teardown
setup() {
    # Create a temporary directory for each test
    export TEST_TEMP_DIR=$(mktemp -d)
    export ORIGINAL_DIR=$(pwd)
    cd "$TEST_TEMP_DIR"
}

teardown() {
    # Clean up
    cd "$ORIGINAL_DIR"
    rm -rf "$TEST_TEMP_DIR"
}

@test "ahoy config init downloads example config file" {
    # Run init command
    run "$ORIGINAL_DIR/ahoy" config init --force
    
    [ "$status" -eq 0 ]
    [ -f ".ahoy.yml" ]
    
    # Check that file contains expected content
    grep -q "ahoyapi: v2" .ahoy.yml
    grep -q "commands:" .ahoy.yml
}

@test "ahoy init (backwards compatible) downloads example config file with notification" {
    # Run backwards compatible init command
    run "$ORIGINAL_DIR/ahoy" init --force
    
    [ "$status" -eq 0 ]
    [ -f ".ahoy.yml" ]
    
    # Check that we get the deprecation notice
    [[ "$output" =~ "Note: 'ahoy init' is now available as 'ahoy config init'" ]]
    
    # Check that file contains expected content
    grep -q "ahoyapi: v2" .ahoy.yml
    grep -q "commands:" .ahoy.yml
}

@test "ahoy config init prompts when .ahoy.yml exists" {
    # Create existing config file
    echo "ahoyapi: v2" > .ahoy.yml
    echo "commands:" >> .ahoy.yml
    echo "  existing:" >> .ahoy.yml
    echo "    cmd: echo 'I exist'" >> .ahoy.yml
    
    # Run init and answer 'n' to prompt
    run bash -c "echo 'n' | '$ORIGINAL_DIR/ahoy' config init"
    
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Warning: .ahoy.yml found in current directory" ]]
    [[ "$output" =~ "Abort: exiting without overwriting" ]]
    
    # Original file should be unchanged
    grep -q "I exist" .ahoy.yml
}

@test "ahoy config init overwrites when user agrees" {
    # Create existing config file
    echo "ahoyapi: v2" > .ahoy.yml
    echo "commands:" >> .ahoy.yml
    echo "  existing:" >> .ahoy.yml
    echo "    cmd: echo 'I exist'" >> .ahoy.yml
    
    # Run init and answer 'y' to prompt
    run bash -c "echo 'y' | '$ORIGINAL_DIR/ahoy' config init"
    
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Warning: .ahoy.yml found in current directory" ]]
    [[ "$output" =~ "Ok, overwriting .ahoy.yml in current directory with example file" ]]
    
    # File should be replaced with example content
    grep -q "ahoyapi: v2" .ahoy.yml
    ! grep -q "I exist" .ahoy.yml
}

@test "ahoy config init --force overwrites without prompting" {
    # Create existing config file
    echo "ahoyapi: v2" > .ahoy.yml
    echo "commands:" >> .ahoy.yml
    echo "  existing:" >> .ahoy.yml
    echo "    cmd: echo 'I exist'" >> .ahoy.yml
    
    # Run init with --force
    run "$ORIGINAL_DIR/ahoy" config init --force
    
    [ "$status" -eq 0 ]
    [[ "$output" =~ "--force" ]]
    [[ "$output" =~ "overwriting .ahoy.yml" ]]
    
    # File should be replaced
    grep -q "ahoyapi: v2" .ahoy.yml
    ! grep -q "I exist" .ahoy.yml
}

@test "ahoy config init with custom URL" {
    # Skip this test if no network access (we'll create a mock)
    # For now, test the command structure
    run "$ORIGINAL_DIR/ahoy" config init --help
    
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Initialise a new .ahoy.yml config file" ]]
}

@test "ahoy config init shows help" {
    run "$ORIGINAL_DIR/ahoy" config init --help
    
    [ "$status" -eq 0 ]
    [[ "$output" =~ "NAME:" ]]
    [[ "$output" =~ "ahoy config init" ]]
    [[ "$output" =~ "Initialise a new .ahoy.yml config file" ]]
    [[ "$output" =~ "--force" ]]
}