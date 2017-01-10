#!/usr/bin/env bats

@test "override bash entrypoint to add additional flags" {
  run ./ahoy -f testdata/entrypoint-bash.ahoy.yml echo something
  [ $status -eq 0 ]
  echo "$output"
  [ "${#lines[@]}" -eq 2 ]
  [ "${lines[0]}" == "+ echo something" ]
  [ "${lines[1]}" == "something" ]

}

@test "override bash entrypoint to use php instead" {
  run ./ahoy -f testdata/entrypoint-php.ahoy.yml echo something
  [ $status -eq 0 ]
  echo "$output"
  [ "${#lines[@]}" -eq 1 ]
  [ "${lines[0]}" == "something" ]
}
