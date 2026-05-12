#!/usr/bin/env bats

# Verbose flag positioning tests.
#
# Cobra's persistent flag handler accepts the verbose flag anywhere relative
# to the subcommand name and its positional arguments. The legacy single-dash
# pre-parser (flag.go) handles the case where the flag appears before the
# subcommand; everything else is delegated to cobra. These tests pin that
# contract and verify that command arguments still reach the underlying
# command without requiring a '--' separator.

@test "Verbose: no flag means no debug output" {
  run ./ahoy -f testdata/simple.ahoy.yml echo hello
  [ $status -eq 0 ]
  [[ "$output" != *"===> Ahoy"* ]]
  [ "$output" = "hello" ]
}

@test "Verbose: -v before command enables debug output" {
  run ./ahoy -v -f testdata/simple.ahoy.yml echo hello
  [ $status -eq 0 ]
  [[ "$output" == *"===> Ahoy echo"* ]]
  [[ "$output" == *"hello"* ]]
}

@test "Verbose: --verbose before command enables debug output" {
  run ./ahoy --verbose -f testdata/simple.ahoy.yml echo hello
  [ $status -eq 0 ]
  [[ "$output" == *"===> Ahoy echo"* ]]
  [[ "$output" == *"hello"* ]]
}

@test "Verbose: -v after command enables debug output" {
  run ./ahoy -f testdata/simple.ahoy.yml echo -v hello
  [ $status -eq 0 ]
  [[ "$output" == *"===> Ahoy echo"* ]]
  [[ "$output" == *"hello"* ]]
}

@test "Verbose: --verbose after command enables debug output" {
  run ./ahoy -f testdata/simple.ahoy.yml echo --verbose hello
  [ $status -eq 0 ]
  [[ "$output" == *"===> Ahoy echo"* ]]
  [[ "$output" == *"hello"* ]]
}

@test "Verbose: -v interspersed between args sets verbose and preserves args" {
  run ./ahoy -f testdata/simple.ahoy.yml echo one -v two
  [ $status -eq 0 ]
  [[ "$output" == *"===> Ahoy echo"* ]]
  [[ "$output" == *"one two"* ]]
}

@test "Args passthru: multiple positional args reach command without '--'" {
  run ./ahoy -f testdata/simple.ahoy.yml echo first second third
  [ $status -eq 0 ]
  [ "$output" = "first second third" ]
}

@test "Args passthru: '--' separator passes -v as a literal argument" {
  run ./ahoy -f testdata/simple.ahoy.yml echo -- -v
  [ $status -eq 0 ]
  [[ "$output" != *"===> Ahoy"* ]]
  [ "$output" = "-v" ]
}

@test "Args passthru: '--' passes verbose-like args through verbatim" {
  run ./ahoy -f testdata/simple.ahoy.yml echo -- hello --verbose world
  [ $status -eq 0 ]
  [[ "$output" != *"===> Ahoy"* ]]
  [ "$output" = "hello --verbose world" ]
}

@test "Verbose: AHOY_VERBOSE env var enables debug output" {
  AHOY_VERBOSE=true run ./ahoy -f testdata/simple.ahoy.yml echo hello
  [ $status -eq 0 ]
  [[ "$output" == *"===> Ahoy echo"* ]]
  [[ "$output" == *"hello"* ]]
}
