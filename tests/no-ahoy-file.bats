#!/usr/bin/env bats

setup() {
  mv .ahoy.yml tmp.ahoy.yml
}

teardown() {
  mv tmp.ahoy.yml .ahoy.yml
}

@test "run ahoy withough a command and without a .ahoy.yml file" {
  result="$(./ahoy)"
  echo "$result"
  [ "$result" -eq 4 ]
}

@test "run an ahoy command without a .ahoy.yml file" {
  result="$(./ahoy something)"
  [ "$result" -eq 4 ]
}

@test "run ahoy init without a .ahoy.yml file" {
  result="$(./ahoy init)"
  [ "$result" -eq 4 ]
}
