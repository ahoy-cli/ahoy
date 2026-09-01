#!/usr/bin/env bats

bats_require_minimum_version 1.5.0

# Regression tests for stderr handling. Subprocesses must keep direct access
# to the stderr inherited by Ahoy so output streams live and cannot block on
# an internal capture pipe.

# timeout(1) is a GNU coreutils command. On macOS it may be absent or only
# available as gtimeout (when coreutils is installed via Homebrew). This
# helper resolves whichever variant is present so the tests are portable.
timeout_cmd() {
  if command -v timeout >/dev/null 2>&1; then
    timeout "$@"
  elif command -v gtimeout >/dev/null 2>&1; then
    gtimeout "$@"
  else
    echo "Neither 'timeout' nor 'gtimeout' found; skipping test" >&2
    skip
  fi
}

setup() {
  TMP_CONFIG="$(mktemp -t ahoy-stderr-XXXXXX).yml"
}

teardown() {
  rm -f "$TMP_CONFIG"
}

@test "Subprocess writing large stderr output does not deadlock ahoy" {
  cat > "$TMP_CONFIG" <<'EOF'
ahoyapi: v2
commands:
  spam-stderr:
    cmd: yes "stderr line padding to fill the pipe buffer faster" | head -n 5000 1>&2
EOF

  # 5000 lines (~300 KB) is well beyond typical pipe buffer capacity.
  # The timeout catches any internal stderr capture that is not drained.
  run timeout_cmd 15 ./ahoy -f "$TMP_CONFIG" spam-stderr
  # `timeout` exits 124 when it had to kill the command.
  [ "$status" -ne 124 ]
  [ "$status" -eq 0 ]
}

@test "Subprocess stderr is streamed live, not buffered until exit" {
  cat > "$TMP_CONFIG" <<'EOF'
ahoyapi: v2
commands:
  large-stderr:
    cmd: yes "x" | head -n 200000 1>&2
EOF

  # 200 000 lines is ~400 KB, confirming sustained stderr output remains
  # live rather than buffered until the command exits.
  run --separate-stderr timeout_cmd 15 ./ahoy -f "$TMP_CONFIG" large-stderr
  [ "$status" -ne 124 ]
  [ "$status" -eq 0 ]

  # Stderr content must actually reach the user (not silently dropped).
  line_count=$(printf '%s' "$stderr" | wc -l | tr -d ' ')
  [ "$line_count" -gt 100000 ]
}

@test "Subprocess inherits ahoy's stderr descriptor, not an internal pipe" {
  # Regression test for issue #180. Ahoy must not interpose anything on
  # stderr: a child that draws a TUI there (gum, anything on Bubble Tea)
  # cannot query the terminal size through a pipe and renders nothing.
  # Here ahoy's own stderr is a regular file, so the child must see a
  # regular file too - seeing a pipe means ahoy inserted one.
  TMP_CONFIG="$BATS_TEST_TMPDIR/fd.ahoy.yml"
  cat > "$TMP_CONFIG" <<'YAML'
ahoyapi: v2
commands:
  fd-kind:
    usage: "report what kind of descriptor is on fd 2"
    cmd: if test -p /dev/fd/2; then echo PIPE; else echo NOT_A_PIPE; fi
YAML

  # Redirect explicitly rather than via `run`, whose own capture would
  # otherwise decide what the child finds on fd 2.
  ./ahoy -f "$TMP_CONFIG" fd-kind \
    >"$BATS_TEST_TMPDIR/stdout.log" 2>"$BATS_TEST_TMPDIR/stderr.log"
  [ "$(cat "$BATS_TEST_TMPDIR/stdout.log")" = "NOT_A_PIPE" ]
}
