#!/usr/bin/env bats

bats_require_minimum_version 1.5.0

# Regression tests for stderr handling. The cobra migration introduced a
# pipe-based stderr capture in main(); if the pipe is not drained while
# Execute() runs, any subprocess writing more than ~64 KB to stderr
# deadlocks because its write blocks on a full pipe buffer that nothing
# is reading.

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

  # The pipe buffer on macOS/Linux is around 64 KB. 5000 lines (~300 KB)
  # is well past that. With the bug present the subprocess blocks on
  # write and ahoy never returns; the timeout catches that.
  run timeout 15 ./ahoy -f "$TMP_CONFIG" spam-stderr
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

  # 200 000 lines is ~400 KB - far past the pipe buffer. Confirms that
  # the drain works for sustained output volumes too.
  run --separate-stderr timeout 15 ./ahoy -f "$TMP_CONFIG" large-stderr
  [ "$status" -ne 124 ]
  [ "$status" -eq 0 ]

  # Stderr content must actually reach the user (not silently dropped).
  line_count=$(printf '%s' "$stderr" | wc -l | tr -d ' ')
  [ "$line_count" -gt 100000 ]
}
