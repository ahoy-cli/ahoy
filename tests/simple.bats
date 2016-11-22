#!/usr/bin/env bats

@test "display help text when no arguments are passed." {
  run ./ahoy -f testdata/simple.ahoy.yml
  # Should throw an error.
  [ $status -ne 0 ]
  echo "$output"
  [ "${#lines[@]}" -gt 10 ]
}

@test "run a simple ahoy command: echo" {
  result="$(./ahoy -f testdata/simple.ahoy.yml echo something)"
  [ "$result" == "something" ]
}
