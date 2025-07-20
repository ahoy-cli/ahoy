#!/usr/bin/env bats

setup() {
  mv .ahoy.yml tmp.ahoy.yml
}

teardown() {
  mv tmp.ahoy.yml .ahoy.yml
  rm -rf wget-lo*
}

@test "Run ahoy without a command and without an .ahoy.yml file" {
  run ./ahoy
  [ $status -eq 1 ]
  # Use array length to get last elements instead of negative indexing
  [ "${lines[$((${#lines[@]}-2))]}" == "[error] No .ahoy.yml found. You can use 'ahoy init' to download an example." ]
  [ "${lines[$((${#lines[@]}-1))]}" == "[warn] Missing flag or argument." ]
}

@test "Run an ahoy command without an .ahoy.yml file" {
  run ./ahoy something
  [ "$output" == "[fatal] Command not found for 'something'" ]
}

@test "Run ahoy init without an .ahoy.yml file" {
  run ./ahoy init
  [ "${lines[$((${#lines[@]}-1))]}" == "Example .ahoy.yml downloaded to the current directory. You can customize it to suit your needs!" ]
}

@test "Run ahoy init with an existing .ahoy.yml file in the current directory" {
  cp tmp.ahoy.yml .ahoy.yml
  run ./ahoy init --force
  [ "${lines[0]}" == "Warning: '--force' parameter passed, overwriting .ahoy.yml in current directory." ]
  [ "${lines[$((${#lines[@]}-1))]}" == "Example .ahoy.yml downloaded to the current directory. You can customize it to suit your needs!" ]
  rm .ahoy.yml
}
