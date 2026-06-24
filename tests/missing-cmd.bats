#!/usr/bin/env bats

@test "A missing command thows error, but doesn't cause a panic." {
  run ./ahoy -f testdata/missing-cmd.ahoy.yml missing-completely
  [ $status -ne 0 ]
  echo "${lines[@]}"
  [ "${lines[0]}" != "panic: runtime error: invalid memory address or nil pointer dereference" ]
  [[ "$output" =~ "Command [missing-completely] has neither 'cmd' or 'imports' set. Check your yaml file." ]]
}

@test "An empty imports throws err, but doesn't cause a panic." {
  run ./ahoy -f testdata/empty-imports.ahoy.yml empty-imports
  [ $status -ne 0 ]
  echo "${lines[@]}"
  [ "${lines[0]}" != "panic: runtime error: invalid memory address or nil pointer dereference" ]
  [[ "$output" =~ "Command [empty-imports] has 'imports' set, but it is empty. Check your yaml file." ]]
}

@test "Circular imports are detected and skipped without a stack overflow." {
  local tmpdir
  tmpdir=$(mktemp -d)
  cat > "$tmpdir/a.ahoy.yml" <<'EOF'
ahoyapi: v2
commands:
  from-a:
    imports:
      - b.ahoy.yml
EOF
  cat > "$tmpdir/b.ahoy.yml" <<'EOF'
ahoyapi: v2
commands:
  from-b:
    imports:
      - a.ahoy.yml
EOF
  run ./ahoy -f "$tmpdir/a.ahoy.yml" list
  rm -rf "$tmpdir"
  # Circular imports result in a broken config (empty non-optional import group),
  # so exit status is non-zero, but the process must not crash or stack overflow.
  [ $status -ne 0 ]
  [ "${lines[0]}" != "panic: runtime error: invalid memory address or nil pointer dereference" ]
  [[ "$output" =~ "Circular import detected" ]]
}

@test "An missing import throws err, but doesn't cause a panic." {
  run ./ahoy -f testdata/missing-imports.ahoy.yml missing-imports
  [ $status -ne 0 ]
  echo "${lines[@]}"
  [ "${lines[0]}" != "panic: runtime error: invalid memory address or nil pointer dereference" ]
  # Error message now includes enhanced diagnostic info with missing file names.
  [[ "$output" =~ "Command [missing-imports] has 'imports' set, but no commands were found." ]]
}
