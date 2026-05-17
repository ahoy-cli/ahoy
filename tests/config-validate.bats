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

@test "ahoy config validate shows help" {
  run "$ORIGINAL_DIR/ahoy" config validate --help
  [ "$status" -eq 0 ]
  [[ "$output" =~ "NAME:" ]]
  [[ "$output" =~ "ahoy config validate" ]]
  [[ "$output" =~ "Validate and diagnose an Ahoy configuration file." ]]
}

@test "ahoy config validate warns when no config file exists" {
  run "$ORIGINAL_DIR/ahoy" config validate
  [ "$status" -eq 0 ]
  [[ "$output" =~ "Warning: No .ahoy.yml file found" ]]
  [[ "$output" =~ "ahoy config init" ]]
}

@test "ahoy config validate succeeds with valid config" {
  cat > .ahoy.yml <<'EOF'
ahoyapi: v2
usage: Test configuration
commands:
  test:
    usage: Test command
    cmd: echo "test"
  hello:
    usage: Say hello
    cmd: echo "Hello, World!"
EOF

  run "$ORIGINAL_DIR/ahoy" config validate
  [ "$status" -eq 0 ]
  [[ "$output" =~ "Ahoy Configuration Validator" ]]
  [[ "$output" =~ "API Version: v2" ]]
  [[ "$output" =~ "No validation issues found" ]]
  [[ "$output" =~ "Configuration looks great!" ]]
}

@test "ahoy config validate detects invalid YAML syntax" {
  cat > .ahoy.yml <<'EOF'
ahoyapi: v2
commands:
  test:
    cmd: echo "test"
    invalid_yaml: [unclosed array
EOF

  # Invalid YAML causes a fatal error during config loading before the validate command runs.
  run "$ORIGINAL_DIR/ahoy" config validate
  [ "$status" -eq 1 ]
  [[ "$output" =~ "yaml:" ]] || [[ "$output" =~ "fatal" ]] || [[ "$output" =~ "Invalid" ]]
}

@test "ahoy config validate detects wrong API version" {
  cat > .ahoy.yml <<'EOF'
ahoyapi: v1
usage: Test configuration
commands:
  test:
    usage: Test command
    cmd: echo test
EOF

  run "$ORIGINAL_DIR/ahoy" config validate
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Invalid YAML" ]] || [[ "$output" =~ "YAML syntax errors" ]] || [[ "$output" =~ "Ahoy only supports API version" ]]
}

@test "ahoy config validate checks environment files" {
  cat > .ahoy.yml <<'EOF'
ahoyapi: v2
env:
  - .env
  - .env.local
commands:
  test:
    usage: Test command
    cmd: echo "test"
    env:
      - .env.command
EOF

  echo "TEST=value" > .env

  run "$ORIGINAL_DIR/ahoy" config validate
  [ "$status" -eq 0 ]
  [[ "$output" =~ "Environment Files:" ]]
  [[ "$output" =~ ".env" ]]
  [[ "$output" =~ "missing" ]]
  [[ "$output" =~ "Consider creating missing environment files" ]]
}

@test "ahoy config validate checks import files" {
  cat > .ahoy.yml <<'EOF'
ahoyapi: v2
commands:
  test1:
    usage: Test command 1
    imports:
      - import1.ahoy.yml
      - import2.ahoy.yml
  test2:
    usage: Test command 2
    imports:
      - import3.ahoy.yml
    optional: true
EOF

  cat > import1.ahoy.yml <<'EOF'
ahoyapi: v2
commands:
  imported:
    cmd: echo "imported"
EOF

  run "$ORIGINAL_DIR/ahoy" config validate
  [ "$status" -eq 0 ]
  [[ "$output" =~ "Import Files:" ]]
  [[ "$output" =~ "import1.ahoy.yml" ]]
  [[ "$output" =~ "import2.ahoy.yml" ]]
  [[ "$output" =~ "missing" ]]
  [[ "$output" =~ "Create missing import files" ]]
}

@test "ahoy config validate can be called with -f to specify config file" {
  run "$ORIGINAL_DIR/ahoy" -f "$ORIGINAL_DIR/testdata/simple.ahoy.yml" config validate
  [ "$status" -eq 0 ]
  [[ "$output" =~ "Ahoy Configuration Validator" ]]
  [[ "$output" =~ "Configuration looks great!" ]]
}

@test "ahoy config validate with simulate-version detects unsupported features" {
  cat > .ahoy.yml <<'EOF'
ahoyapi: v2
commands:
  test:
    usage: Test with alias
    cmd: echo test
    aliases:
      - t
EOF

  # command_aliases is a warning (not error), so exit code must be 0.
  run "$ORIGINAL_DIR/ahoy" --simulate-version v2.0.0 config validate
  [ "$status" -eq 0 ]
  [[ "$output" =~ "aliases" ]] || [[ "$output" =~ "warning" ]] || [[ "$output" =~ "Warning" ]]
}

@test "ahoy config validate exits 1 when simulate-version triggers error-severity issue" {
  cat > .ahoy.yml <<'EOF'
ahoyapi: v2
commands:
  fetch:
    usage: Fetch subcommands
    imports:
      - sub.ahoy.yml
    optional: true
EOF

  # optional_imports requires v2.2.0; v2.1.0 triggers an error-severity issue.
  run "$ORIGINAL_DIR/ahoy" --simulate-version v2.1.0 config validate
  [ "$status" -eq 1 ]
  [[ "$output" =~ "optional" ]]
}
