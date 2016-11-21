#!/usr/bin/env bats

@test "get the version of ahoy with --version" {
  run ./ahoy -f testdata/simple.ahoy.yml --version
  [ $status -eq 0 ]
  [ "$result" == "2.0.0-alpha-23-g0479480" ]
}

@test "get help instead of running a command with --help" {
  result="$(./ahoy -f testdata/simple.ahoy.yml --help echo something)"
  [ "$result" != "something" ]
}
