#!/usr/bin/env bats

load 'test_helpers/bats-support/load'
load 'test_helpers/bats-assert/load'

@test "Test cross compilation command lifecycle" {
  run make clean
  assert_success
  assert_output --partial 'rm -vRf ./builds/ahoy-bin-*'

  run ls ./builds/ahoy-*
  assert_failure
  assert_output --partial 'No such file'

  run make cross
  assert_success

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
