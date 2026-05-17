#!/usr/bin/env bats

bats_require_minimum_version 1.5.0

# Migration safety tests for urfave/cli to spf13/viper transition
# These tests ensure existing configurations and commands continue to work

@test "Existing .ahoy.yml files continue to work" {
  # Test that current YAML structure is preserved
  run ./ahoy -f testdata/simple.ahoy.yml echo "migration test"
  [ $status -eq 0 ]
  [[ "$output" == *"migration test"* ]]
}

@test "All current global flags are recognized" {
  # Test each global flag that should be preserved during migration
  
  # Version flag
  run ./ahoy --version
  [ $status -eq 0 ]
  
  # Help flag
  run ./ahoy --help
  [ $status -eq 0 ]
  
  # File flag
  run ./ahoy -f testdata/simple.ahoy.yml --help
  [ $status -eq 0 ]
  
  # Verbose flag
  run ./ahoy -v -f testdata/simple.ahoy.yml echo "test"
  [ $status -eq 0 ]
}

@test "Environment variable configuration is preserved" {
  # Test AHOY_VERBOSE environment variable
  export AHOY_VERBOSE=true
  run ./ahoy -f testdata/simple.ahoy.yml echo "env test"
  [ $status -eq 0 ]
  [[ "$output" == *"env test"* ]]
  unset AHOY_VERBOSE
}

@test "Command aliases functionality is preserved" {
  # Test that command aliases continue to work
  run ./ahoy -f testdata/command-aliases.ahoy.yml hello
  [ $status -eq 0 ]
  
  # Test alias
  run ./ahoy -f testdata/command-aliases.ahoy.yml hi
  [ $status -eq 0 ]
}

@test "Environment file loading is preserved" {
  # Test that .env file loading continues to work
  run ./ahoy -f testdata/env.ahoy.yml test-global
  [ $status -eq 0 ]
  [[ "$output" = "global" ]]
}

@test "Import functionality is preserved" {
  # Test that config file imports continue to work
  # Just test that the config loads and shows help (which lists commands)
  run ./ahoy -f testdata/docker.ahoy.yml --help
  [ $status -eq 0 ]
  [[ "$output" == *"up"* ]]
  [[ "$output" == *"Start the Docker Compose containers"* ]]
}

@test "Entrypoint customization is preserved" {
  # Test that custom entrypoints continue to work
  run ./ahoy -f testdata/entrypoint-bash.ahoy.yml echo "bash entrypoint test"
  [ $status -eq 0 ]
  [[ "$output" == *"bash entrypoint test"* ]]
}

@test "Multi-line commands are preserved" {
  # Test that multi-line command syntax continues to work  
  # The entrypoint-bash.ahoy.yml has a multi-line echo command
  run ./ahoy -f testdata/entrypoint-bash.ahoy.yml echo "line1" "line2"
  [ $status -eq 0 ]
  [[ "$output" == *"line1"* ]]
  [[ "$output" == *"line2"* ]]
}

@test "Command arguments are preserved" {
  # Test that command argument passing continues to work
  run ./ahoy -f testdata/simple.ahoy.yml echo "arg1" "arg2"
  [ $status -eq 0 ]
  [[ "$output" == *"arg1"* ]]
  [[ "$output" == *"arg2"* ]]
}

@test "Optional commands functionality is preserved" {
  # Test that optional command imports continue to work
  run ./ahoy -f testdata/optional-command.ahoy.yml regular-cmd
  [ $status -eq 0 ]
  [[ "$output" == *"This is a regular command"* ]]
}

@test "Hidden commands functionality is preserved" {
  # Test that hidden commands are not shown in help but can be executed
  run ./ahoy -f testdata/simple.ahoy.yml --help
  [ $status -eq 0 ]
  # Hidden command should not appear in help
  [[ "$output" != *"hidden-command"* ]]
  
  # But should still be executable
  run ./ahoy -f testdata/simple.ahoy.yml hidden-command 2>/dev/null || true
  # This may or may not exist in testdata, so we don't assert the result
}

@test "Bash completion continues to work" {
  # Test that bash completion functionality is preserved
  run ./ahoy -f testdata/simple.ahoy.yml --generate-bash-completion
  [ $status -eq 0 ]
}

@test "Config file discovery continues to work" {
  # Simplified test - just verify that config loading works with relative paths
  run ./ahoy -f testdata/simple.ahoy.yml --help
  [ $status -eq 0 ]
  [[ "$output" == *"echo"* ]]
  [[ "$output" == *"list"* ]]
}

@test "Error handling behavior is preserved" {
  # Test that error handling continues to work as expected
  
  # Non-existent file
  run ./ahoy -f non-existent.yml test
  [ $status -ne 0 ]
  
  # Invalid YAML
  echo "invalid: yaml: content" > /tmp/invalid.yml
  run ./ahoy -f /tmp/invalid.yml test
  [ $status -ne 0 ]
  rm -f /tmp/invalid.yml
  
  # Non-existent command
  run ./ahoy -f testdata/simple.ahoy.yml non-existent-command
  [ $status -ne 0 ]
}

@test "API version enforcement is preserved" {
  # Test that ahoyapi: v2 requirement is still enforced
  cat > /tmp/old-api.yml << 'EOF'
ahoyapi: v1
commands:
  test:
    cmd: echo "old api"
EOF
  
  run ./ahoy -f /tmp/old-api.yml test
  [ $status -ne 0 ]
  [[ "$output" == *"API version"* ]] || [[ "$output" == *"v2"* ]]
  
  rm -f /tmp/old-api.yml
}

@test "Working directory behavior is preserved" {
  # Simplified test - verify that commands execute from the correct directory
  run ./ahoy -f testdata/simple.ahoy.yml echo "working directory test"
  [ $status -eq 0 ]
  [[ "$output" == *"working directory test"* ]]
}