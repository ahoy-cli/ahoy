#!/usr/bin/env bats

@test "get the version of ahoy with --version" {
  run ./ahoy -f testdata/simple.ahoy.yml --version
  [ $status -eq 0 ]
  [ $(expr "$output" : "^v?[0-9.]\.[0-9.]\.[0-9.](\S*)?") -eq 0 ]
}

@test "get help instead of running a command with --help" {
  result="$(./ahoy -f testdata/simple.ahoy.yml --help echo something)"
  [ "$result" != "something" ]
}
