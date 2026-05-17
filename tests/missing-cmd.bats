#!/usr/bin/env bats

@test "A missing command thows error, but doesn't cause a panic." {
  run ./ahoy -f testdata/missing-cmd.ahoy.yml missing-completely
  [ $status -ne 0 ]
  echo "${lines[@]}"
  [ "${lines[0]}" != "panic: runtime error: invalid memory address or nil pointer dereference" ]
  [ "${lines[0]}" == "[fatal] Command [missing-completely] has neither 'cmd' or 'imports' set. Check your yaml file." ]
}

@test "An empty imports throws err, but doesn't cause a panic." {
  run ./ahoy -f testdata/empty-imports.ahoy.yml empty-imports
  [ $status -ne 0 ]
  echo "${lines[@]}"
  [ "${lines[0]}" != "panic: runtime error: invalid memory address or nil pointer dereference" ]
  [ "${lines[0]}" == "[fatal] Command [empty-imports] has 'imports' set, but it is empty. Check your yaml file." ]
}

@test "An missing import throws err, but doesn't cause a panic." {
  run ./ahoy -f testdata/missing-imports.ahoy.yml missing-imports
  [ $status -ne 0 ]
  echo "${lines[@]}"
  [ "${lines[0]}" != "panic: runtime error: invalid memory address or nil pointer dereference" ]
  # Error message now includes enhanced diagnostic info with missing file names.
  [[ "$output" =~ "Command [missing-imports] has 'imports' set, but no commands were found." ]]
}
