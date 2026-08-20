#!/usr/bin/env bats

# Comprehensive CLI flag combination tests for urfave/cli compatibility

@test "Short and long version flags work identically" {
  run ./ahoy --version
  version_long=$status
  version_long_output="$output"
  
  run ./ahoy -version
  version_short=$status
  version_short_output="$output"
  
  [ $version_long -eq $version_short ]
  [ "$version_long_output" = "$version_short_output" ]
}

@test "Short and long help flags work identically" {
  run ./ahoy --help
  help_long=$status
  help_long_output="$output"
  
  run ./ahoy -h
  help_short=$status
  help_short_output="$output"
  
  [ $help_long -eq $help_short ]
  [[ "$help_long_output" == *"Creates a configurable cli app"* ]]
  [[ "$help_short_output" == *"Creates a configurable cli app"* ]]
}

@test "Short and long file flags work identically" {
  run ./ahoy -f testdata/simple.ahoy.yml echo "test"
  file_short=$status
  file_short_output="$output"
  
  run ./ahoy --file testdata/simple.ahoy.yml echo "test"
  file_long=$status
  file_long_output="$output"
  
  [ $file_short -eq $file_long ]
  [ "$file_short_output" = "$file_long_output" ]
}

@test "Short and long verbose flags work identically" {
  run ./ahoy -v -f testdata/simple.ahoy.yml echo "test"
  verbose_short=$status
  
  run ./ahoy --verbose -f testdata/simple.ahoy.yml echo "test"
  verbose_long=$status
  
  [ $verbose_short -eq $verbose_long ]
}

@test "Multiple flag combinations work correctly" {
  # Test various combinations of flags
  run ./ahoy -v -f testdata/simple.ahoy.yml --help
  [ $status -eq 0 ]
  
  run ./ahoy --verbose --file testdata/simple.ahoy.yml --help
  [ $status -eq 0 ]
  
  run ./ahoy -f testdata/simple.ahoy.yml -v --help
  [ $status -eq 0 ]
}

@test "Flag order doesn't matter" {
  run ./ahoy -f testdata/simple.ahoy.yml -v echo "test"
  order1=$status
  order1_output="$output"
  
  run ./ahoy -v -f testdata/simple.ahoy.yml echo "test"
  order2=$status
  order2_output="$output"
  
  [ $order1 -eq $order2 ]
  [[ "$order1_output" == *"test"* ]]
  [[ "$order2_output" == *"test"* ]]
}

@test "Invalid flag combinations are handled gracefully" {
  # Test with non-existent file
  run ./ahoy -f non-existent.yml echo "test"
  [ $status -ne 0 ]
  [[ "$output" == *"error"* ]] || [[ "$output" == *"fatal"* ]]
  
  # An invalid flag should show help and exit 1 (failure, not success).
  run ./ahoy --invalid-flag
  [ $status -eq 1 ]
  [[ "$output" == *"flag provided but not defined"* ]]
}

@test "Environment variable flags work correctly" {
  # Test AHOY_VERBOSE environment variable
  export AHOY_VERBOSE=true
  run ./ahoy -f testdata/simple.ahoy.yml echo "test"
  [ $status -eq 0 ]
  
  unset AHOY_VERBOSE
}

@test "Boolean flags can be specified without values" {
  # Test that boolean flags work without explicit values
  run ./ahoy --verbose -f testdata/simple.ahoy.yml echo "test"
  [ $status -eq 0 ]
  
  run ./ahoy -v -f testdata/simple.ahoy.yml echo "test"
  [ $status -eq 0 ]
}

@test "Flags with commands work correctly" {
  # Test flags when used with specific commands
  run ./ahoy -f testdata/simple.ahoy.yml -v echo "test with flags"
  [ $status -eq 0 ]
  [[ "$output" == *"test with flags"* ]]
}

@test "Double dash separator works correctly" {
  # Test that -- separates ahoy flags from command arguments
  run ./ahoy -f testdata/simple.ahoy.yml echo -- --help
  [ $status -eq 0 ]
  [[ "$output" == *"--help"* ]]
}