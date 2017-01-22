#!/usr/bin/env bats

@test "A missing command doesn't cause a panic." {
  run ./ahoy -f testdata/missing-cmd.ahoy.yml missing-completely
  [ $status -ne 0 ]
  echo "${lines[@]}"
  [ "${lines[0]}" != "panic: runtime error: invalid memory address or nil pointer dereference" ]
}
