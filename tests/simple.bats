#!/usr/bin/env bats

@test "run a simple ahoy command: echo" {
  result="$(ahoy -f testdata/simple.ahoy.yml echo something)"
  [ "$result" == "something" ]
}
