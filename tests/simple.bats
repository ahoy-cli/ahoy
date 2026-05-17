#!/usr/bin/env bats

@test "Display help text and fatal error when no arguments are passed." {
  run ./ahoy -f testdata/simple.ahoy.yml
  # Should throw an error.
  [ $status -ne 0 ]
  echo "$output"
  [ "${#lines[@]}" -gt 10 ]
  [ "${lines[$((${#lines[@]}-1))]}" == "[warn] Missing flag or argument." ]
}

@test "Run a simple ahoy command: echo" {
  result="$(./ahoy -f testdata/simple.ahoy.yml echo something)"
  [ "$result" == "something" ]
}

@test "Run a simple ahoy command (ls -a) with an extra parameter (-l)" {
  run ./ahoy -f testdata/simple.ahoy.yml list -- -l
  [ "${#lines[@]}" -gt 13 ]
}
