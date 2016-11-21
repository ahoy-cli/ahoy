#!/usr/bin/env bats

@test "get the version of ahoy with --version" {
  result="$(ahoy -f testdata/simple.ahoy.yml --version)"
  [ "$result" == "master" ]
}

@test "get help instead of running a command with --help" {
  result="$(ahoy -f testdata/simple.ahoy.yml --help echo something)"
  [ "$result" != "something" ]
}
