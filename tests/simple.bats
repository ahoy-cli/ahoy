#!/usr/bin/env bats

@test "display help text when no arguments are passed." {
  run ./ahoy -f testdata/simple.ahoy.yml
  # Should throw an error.
  [ "${#lines[@]}" -gt 14 ]
  [ $status -eq 1 ]
}

@test "run a simple ahoy command: echo" {
  result="$(./ahoy -f testdata/simple.ahoy.yml echo something)"
  [ "$result" == "something" ]
}
