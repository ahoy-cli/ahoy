#!/usr/bin/env bats

# Tests for issue #188.
#
# A missing env file has always been skipped, which the layered-override
# pattern relies on. What was missing was any way to tell Ahoy that a file
# is meant to be there: a command whose credentials file was never created
# would run with every variable empty and say nothing about it.
#
# An env entry is now required by default and reported on stderr when absent,
# and 'optional: true' marks the ones that are expected to come and go.

setup() {
  TMP_DIR="$BATS_TEST_TMPDIR"
  TMP_CONFIG="$TMP_DIR/.ahoy.yml"
}

@test "A missing required env file warns and the command still runs" {
  cat > "$TMP_CONFIG" <<'YAML'
ahoyapi: v2
commands:
  deploy:
    env:
      - .env
    cmd: echo "HOST=[$DEPLOY_HOST]"
YAML

  run ./ahoy -f "$TMP_CONFIG" deploy
  [ "$status" -eq 0 ]
  [[ "$output" == *"environment file '.env' not found"* ]]
  [[ "$output" == *"HOST=[]"* ]]
}

@test "The warning names 'optional: true' as the way to silence it" {
  cat > "$TMP_CONFIG" <<'YAML'
ahoyapi: v2
commands:
  deploy:
    env:
      - .env
    cmd: echo ok
YAML

  run ./ahoy -f "$TMP_CONFIG" deploy
  [[ "$output" == *"optional: true"* ]]
}

@test "A missing optional env file is silent" {
  cat > "$TMP_CONFIG" <<'YAML'
ahoyapi: v2
commands:
  deploy:
    env:
      - path: .env.local
        optional: true
    cmd: echo "LOCAL=[$LOCAL_ONLY]"
YAML

  run ./ahoy -f "$TMP_CONFIG" deploy
  [ "$status" -eq 0 ]
  [[ "$output" != *"not found"* ]]
  [ "$output" = "LOCAL=[]" ]
}

@test "The warning goes to stderr, leaving the command's stdout clean" {
  cat > "$TMP_CONFIG" <<'YAML'
ahoyapi: v2
commands:
  deploy:
    env:
      - .env
    cmd: echo "just-the-output"
YAML

  run --separate-stderr ./ahoy -f "$TMP_CONFIG" deploy
  [ "$status" -eq 0 ]
  [ "$output" = "just-the-output" ]
  [[ "$stderr" == *"environment file '.env' not found"* ]]
}

@test "A missing global env file warns too" {
  cat > "$TMP_CONFIG" <<'YAML'
ahoyapi: v2
env:
  - .env.global
commands:
  show:
    cmd: echo done
YAML

  run ./ahoy -f "$TMP_CONFIG" show
  [ "$status" -eq 0 ]
  [[ "$output" == *"environment file '.env.global' not found"* ]]
}

@test "A present required file loads with no warning" {
  printf 'DEPLOY_HOST=prod\n' > "$TMP_DIR/.env"
  cat > "$TMP_CONFIG" <<'YAML'
ahoyapi: v2
commands:
  deploy:
    env:
      - .env
    cmd: echo "HOST=[$DEPLOY_HOST]"
YAML

  run ./ahoy -f "$TMP_CONFIG" deploy
  [ "$status" -eq 0 ]
  [[ "$output" != *"not found"* ]]
  [ "$output" = "HOST=[prod]" ]
}

@test "Plain paths and optional entries mix, and precedence is unchanged" {
  printf 'VALUE=base\nBASE_ONLY=yes\n' > "$TMP_DIR/.env.base"
  printf 'VALUE=local\n' > "$TMP_DIR/.env.local"
  cat > "$TMP_CONFIG" <<'YAML'
ahoyapi: v2
commands:
  show:
    env:
      - .env.base
      - path: .env.local
        optional: true
    cmd: echo "VALUE=[$VALUE] BASE_ONLY=[$BASE_ONLY]"
YAML

  run ./ahoy -f "$TMP_CONFIG" show
  [ "$status" -eq 0 ]
  [[ "$output" != *"not found"* ]]
  # Later files win, and earlier ones still contribute their own keys.
  [ "$output" = "VALUE=[local] BASE_ONLY=[yes]" ]
}

@test "An optional entry that exists is loaded like any other" {
  printf 'LOCAL_ONLY=yes\n' > "$TMP_DIR/.env.local"
  cat > "$TMP_CONFIG" <<'YAML'
ahoyapi: v2
commands:
  show:
    env:
      - path: .env.local
        optional: true
    cmd: echo "LOCAL=[$LOCAL_ONLY]"
YAML

  run ./ahoy -f "$TMP_CONFIG" show
  [ "$output" = "LOCAL=[yes]" ]
}

@test "'optional: false' is required, and warns when absent" {
  cat > "$TMP_CONFIG" <<'YAML'
ahoyapi: v2
commands:
  deploy:
    env:
      - path: .env
        optional: false
    cmd: echo ok
YAML

  run ./ahoy -f "$TMP_CONFIG" deploy
  [ "$status" -eq 0 ]
  [[ "$output" == *"environment file '.env' not found"* ]]
}

@test "An env entry with no path is a config error, not an empty path" {
  cat > "$TMP_CONFIG" <<'YAML'
ahoyapi: v2
commands:
  deploy:
    env:
      - optional: true
    cmd: echo ok
YAML

  run ./ahoy -f "$TMP_CONFIG" deploy
  [ "$status" -ne 0 ]
  [[ "$output" == *"env entry 1: env entry has no file path"* ]]
  # The message must be ours, not the decoder's rendering of the internal
  # struct that mappings are unmarshalled through.
  [[ "$output" != *"struct {"* ]]
}

@test "An empty env path is a config error rather than a file named ''" {
  cat > "$TMP_CONFIG" <<'YAML'
ahoyapi: v2
commands:
  deploy:
    env:
      - ""
    cmd: echo "COMMAND-DID-RUN"
YAML

  run ./ahoy -f "$TMP_CONFIG" deploy
  [ "$status" -ne 0 ]
  [[ "$output" == *"env entry has no file path"* ]]
  [[ "$output" != *"COMMAND-DID-RUN"* ]]
  [[ "$output" != *"environment file '' not found"* ]]
}

@test "A null env entry is a config error and names its position" {
  cat > "$TMP_CONFIG" <<'YAML'
ahoyapi: v2
commands:
  deploy:
    env:
      - .env
      - null
    cmd: echo "COMMAND-DID-RUN"
YAML

  run ./ahoy -f "$TMP_CONFIG" deploy
  [ "$status" -ne 0 ]
  [[ "$output" == *"env entry 2"* ]]
  [[ "$output" != *"COMMAND-DID-RUN"* ]]
}

@test "config validate distinguishes a missing optional file from a missing required one" {
  cat > "$TMP_CONFIG" <<'YAML'
ahoyapi: v2
commands:
  deploy:
    env:
      - .env
      - path: .env.local
        optional: true
    cmd: echo ok
YAML

  run ./ahoy -f "$TMP_CONFIG" config validate
  [[ "$output" == *"❌ .env "* ]]
  [[ "$output" == *"marked optional"* ]]
  # Only the required file is raised as an issue.
  [[ "$output" == *"Environment file '.env' not found"* ]]
  [[ "$output" != *"Environment file '.env.local' not found"* ]]
}
