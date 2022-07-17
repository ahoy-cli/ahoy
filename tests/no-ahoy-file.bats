#!/usr/bin/env bats

setup() {
  mv .ahoy.yml tmp.ahoy.yml
}

teardown() {
  mv tmp.ahoy.yml .ahoy.yml
}

@test "run ahoy without a command and without a .ahoy.yml file" {
  run ./ahoy
  [ $status -eq 1 ]
  [ "${lines[-2]}" == "[error] No .ahoy.yml found. You can use 'ahoy init' to download an example." ]
  [ "${lines[-1]}" == "[fatal] Missing flag or argument." ]
}

@test "run an ahoy command without a .ahoy.yml file" {
  run ./ahoy something
  [ "$output" == "[fatal] Command not found for 'something'" ]
}

@test "run ahoy init without a .ahoy.yml file" {
  run ./ahoy init
  [ "${lines[-1]}" == "Example .ahoy.yml downloaded to the current directory. You can customize it to suit your needs!" ]
}

@test "run ahoy init with a existing .ahoy.yml file in the current directory" {
  cp tmp.ahoy.yml .ahoy.yml
  run ./ahoy init
  [ "${lines[-1]}" == "Warning: .ahoy.yml found in current directory." ]
  rm .ahoy.yml
}
