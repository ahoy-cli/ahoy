#!/usr/bin/env bats

setup() {
  mv .ahoy.yml tmp.ahoy.yml
}

teardown() {
  mv tmp.ahoy.yml .ahoy.yml
}

@test "run ahoy init without a .ahoy.yml file" {
  result="$(./ahoy init)"
  [ "$result" -eq 4 ]
}
