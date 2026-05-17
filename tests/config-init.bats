#!/usr/bin/env bats

setup() {
  export TEST_TEMP_DIR=$(mktemp -d)
  export ORIGINAL_DIR=$(pwd)
  cd "$TEST_TEMP_DIR"
}

teardown() {
  cd "$ORIGINAL_DIR"
  rm -rf "$TEST_TEMP_DIR"
}

@test "ahoy config init shows help" {
  run "$ORIGINAL_DIR/ahoy" config init --help
  [ "$status" -eq 0 ]
  [[ "$output" =~ "NAME:" ]]
  [[ "$output" =~ "ahoy config init" ]]
  [[ "$output" =~ "Initialise a new .ahoy.yml config file" ]]
  [[ "$output" =~ "--force" ]]
}

@test "ahoy config init downloads example config file" {
  run "$ORIGINAL_DIR/ahoy" config init --force
  [ "$status" -eq 0 ]
  [ -f ".ahoy.yml" ]
  grep -q "ahoyapi: v2" .ahoy.yml
  grep -q "commands:" .ahoy.yml
}

@test "ahoy config init prompts when .ahoy.yml exists" {
  cat > .ahoy.yml <<'EOF'
ahoyapi: v2
commands:
  existing:
    cmd: echo 'I exist'
EOF

  run bash -c "echo 'n' | '$ORIGINAL_DIR/ahoy' config init"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "Warning: .ahoy.yml found in current directory" ]]
  [[ "$output" =~ "Abort: exiting without overwriting" ]]
  grep -q "I exist" .ahoy.yml
}

@test "ahoy config init overwrites when user agrees" {
  cat > .ahoy.yml <<'EOF'
ahoyapi: v2
commands:
  existing:
    cmd: echo 'I exist'
EOF

  run bash -c "echo 'y' | '$ORIGINAL_DIR/ahoy' config init"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "Warning: .ahoy.yml found in current directory" ]]
  [[ "$output" =~ "Ok, overwriting .ahoy.yml in current directory with example file" ]]
  grep -q "ahoyapi: v2" .ahoy.yml
}

@test "ahoy config init --force overwrites without prompting" {
  cat > .ahoy.yml <<'EOF'
ahoyapi: v2
commands:
  existing:
    cmd: echo 'I exist'
EOF

  run "$ORIGINAL_DIR/ahoy" config init --force
  [ "$status" -eq 0 ]
  [[ "$output" =~ "force" ]]
  [[ "$output" =~ "overwriting .ahoy.yml" ]]
  grep -q "ahoyapi: v2" .ahoy.yml
}

@test "ahoy init (deprecated) shows deprecation notice" {
  run "$ORIGINAL_DIR/ahoy" init --force
  [ "$status" -eq 0 ]
  [[ "$output" =~ "deprecated" ]]
  [[ "$output" =~ "ahoy config init" ]]
  [ -f ".ahoy.yml" ]
  grep -q "ahoyapi: v2" .ahoy.yml
}
