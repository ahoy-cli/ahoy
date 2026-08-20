#!/usr/bin/env bats

setup() {
  # Create a temporary directory for isolation to prevent finding parent .ahoy.yml files
  TEST_DIR=$(mktemp -d)
  
  # Copy the ahoy binary to the temporary directory
  # We expect the binary to be in the current directory (where bats is invoked)
  if [ -f "./ahoy" ]; then
    cp ./ahoy "$TEST_DIR/"
  elif [ -f "../ahoy" ]; then
    # Handle case where bats might be run from tests dir
    cp "../ahoy" "$TEST_DIR/"
  else
    echo "Error: ahoy binary not found" >&2
    exit 1
  fi
  
  # Save original directory and switch to temp dir
  export ORIG_DIR=$(pwd)
  cd "$TEST_DIR"
}

teardown() {
  cd "$ORIG_DIR"
  rm -rf "$TEST_DIR"
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
  echo "ahoyapi: v2" > .ahoy.yml
  run ./ahoy init --force
  [ "${lines[0]}" == "Warning: '--force' parameter passed, overwriting .ahoy.yml in current directory." ]
  [ "${lines[$((${#lines[@]}-1))]}" == "Example .ahoy.yml downloaded to the current directory. You can customize it to suit your needs!" ]
  rm .ahoy.yml
}
