#!/usr/bin/env bats

@test "display help text and fatal error when no arguments are passed." {
  run ./ahoy -f testdata/simple.ahoy.yml
  # Should throw an error.
  [ $status -ne 0 ]
  echo "$output"
  [ "${#lines[@]}" -gt 10 ]
  [ "${lines[-1]}" == "[fatal] Missing flag or argument." ]
}

@test "run a simple ahoy command: echo" {
  result="$(./ahoy -f testdata/simple.ahoy.yml echo something)"
  [ "$result" == "something" ]
}

@test "run a simple ahoy command (ls -a) with an extra parameter (-l)" {
  run ./ahoy -f testdata/simple.ahoy.yml list -- -l
  [ "${#lines[@]}" -gt 13 ]
}

@test "override an ahoy command with another command" {
  result="$(./ahoy -f testdata/override-base.ahoy.yml docker override-example)"
  [ "$result" == "Overrode you." ]
}
