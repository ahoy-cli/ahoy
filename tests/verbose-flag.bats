#!/usr/bin/env bats

# Verbose flag positioning tests.
#
# Ahoy's own flags are only recognised BEFORE the command name, where the
# pre-parser in flag.go reads them. Everything from the command name onwards
# belongs to the command and is passed through exactly as typed - including
# tokens that look like ahoy's own flags. This is the v2 contract, restored
# after v3.0.0 briefly claimed flags anywhere and silently ate arguments
# meant for the wrapped command (issue #182).

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

@test "Verbose: -v after command belongs to the command, not to ahoy" {
  run ./ahoy -f testdata/simple.ahoy.yml echo -v hello
  [ $status -eq 0 ]
  [[ "$output" != *"===> Ahoy"* ]]
  [ "$output" = "-v hello" ]
}

@test "Verbose: --verbose after command belongs to the command, not to ahoy" {
  run ./ahoy -f testdata/simple.ahoy.yml echo --verbose hello
  [ $status -eq 0 ]
  [[ "$output" != *"===> Ahoy"* ]]
  [ "$output" = "--verbose hello" ]
}

@test "Verbose: -v between args is preserved in place" {
  run ./ahoy -f testdata/simple.ahoy.yml echo one -v two
  [ $status -eq 0 ]
  [[ "$output" != *"===> Ahoy"* ]]
  [ "$output" = "one -v two" ]
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
