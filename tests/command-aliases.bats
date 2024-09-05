#!/usr/bin/env bats

load 'test_helpers/bats-support/load'
load 'test_helpers/bats-assert/load'

@test "Command aliases work correctly" {
  run ./ahoy -f testdata/command-aliases.ahoy.yml hello
  [[ "$output" =~ "Hello, World!" ]]
  [ "$status" -eq 0 ]

  run ./ahoy -f testdata/command-aliases.ahoy.yml hi
  [[ "$output" =~ "Hello, World!" ]]
  [ "$status" -eq 0 ]

  run ./ahoy -f testdata/command-aliases.ahoy.yml greet
  [[ "$output" =~ "Hello, World!" ]]
  [ "$status" -eq 0 ]

  run ./ahoy -f testdata/command-aliases.ahoy.yml
  [[ "$output" =~ "hello, hi, greet" ]]
  [[ "$output" =~ "Aliases: hi, greet" ]]

  # Should exit with error as no command was supplied.
  [ "$status" -eq 1 ]
}

@test "A cow says ahoy" {
  run ./ahoy -f testdata/command-aliases.ahoy.yml cowsay
  [[ "$output" =~ "ahoy there!" ]]
  [ "$status" -eq 0 ]

  run ./ahoy -f testdata/command-aliases.ahoy.yml ahoy
  [[ "$output" =~ "ahoy there!" ]]
  [ "$status" -eq 0 ]
}

@test "Multiple conflicting aliases means the last one loaded takes precedence" {
  run ./ahoy -f testdata/command-aliases.ahoy.yml ahoy
  [[ "$output" =~ "ahoy there!" ]]
  [ "$status" -eq 0 ]
}
