#!/usr/bin/env bats

load 'test_helpers/bats-support/load'
load 'test_helpers/bats-assert/load'

@test "test cross compilation command lifecycle" {
  run make clean
  assert_success
  assert_output --partial 'rm -vRf ./builds/ahoy-bin-*'

  run ls ./builds/ahoy-*
  assert_failure
  assert_output --partial 'No such file'

  run make cross
  assert_success
  assert_output --partial 'mv ./builds/ahoy-bin-windows-amd64 ./builds/ahoy-bin-windows-amd64.exe; mv ./builds/ahoy-bin-windows-arm64 ./builds/ahoy-bin-windows-arm64.exe;'

  run ls ./builds/ahoy-*
  assert_output --partial 'ahoy-bin-darwin-amd64'
  assert_output --partial 'ahoy-bin-linux-arm64'
  assert_output --partial 'ahoy-bin-windows-amd64.exe'

  run make clean
  assert_success
  assert_output --partial 'rm -vRf ./builds/ahoy-bin-*'

  run ls ./builds/ahoy-*
  assert_failure
  assert_output --partial 'No such file'
}
