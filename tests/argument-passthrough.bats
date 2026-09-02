#!/usr/bin/env bats

# Regression tests for issue #182.
#
# A command defined in .ahoy.yml owns every argument after its own name.
# v3.0.0 let cobra parse those arguments first, and pflag's unknown-flag
# handling dropped each flag AND the token following it - so a wrapper like
# `cmd: docker compose exec app $@` silently ran a different command than
# the one the user typed. These tests pin the v2 contract: verbatim, in
# order, no exceptions.

setup() {
  TMP_CONFIG="$BATS_TEST_TMPDIR/.ahoy.yml"
  cat > "$TMP_CONFIG" <<'YAML'
ahoyapi: v2
commands:
  echoargs:
    usage: Echo the arguments it receives
    cmd: echo "$@"
YAML
}

@test "Unknown long flag and its value survive" {
  run ./ahoy -f "$TMP_CONFIG" echoargs AAA --log-junit x.xml --configuration y.dist BBB
  [ "$status" -eq 0 ]
  [ "$output" = "AAA --log-junit x.xml --configuration y.dist BBB" ]
}

@test "Unknown long flag in '=' form survives" {
  run ./ahoy -f "$TMP_CONFIG" echoargs AAA --log-junit=x.xml BBB
  [ "$status" -eq 0 ]
  [ "$output" = "AAA --log-junit=x.xml BBB" ]
}

@test "Unknown short flags survive" {
  run ./ahoy -f "$TMP_CONFIG" echoargs AAA -R -p BBB
  [ "$status" -eq 0 ]
  [ "$output" = "AAA -R -p BBB" ]
}

@test "A quoted flag value containing spaces survives as one argument" {
  run ./ahoy -f "$TMP_CONFIG" echoargs START --filter='My Test Name' -R END
  [ "$status" -eq 0 ]
  [ "$output" = "START --filter=My Test Name -R END" ]
}

@test "A trailing positional argument after a short flag is not swallowed" {
  # pflag treated the token after an unknown flag as that flag's value, so
  # the flag, its value AND the trailing argument all disappeared.
  run ./ahoy -f "$TMP_CONFIG" echoargs CMD -p /tmp/logs END
  [ "$status" -eq 0 ]
  [ "$output" = "CMD -p /tmp/logs END" ]
}

@test "Flags a command hard-codes for its own tooling survive" {
  # The reported real-world breakage: options with no user input involved.
  run ./ahoy -f "$TMP_CONFIG" echoargs chown -R www-data:www-data .
  [ "$status" -eq 0 ]
  [ "$output" = "chown -R www-data:www-data ." ]
}

@test "Ahoy's own flags before the command name are still ahoy's" {
  run ./ahoy -v -f "$TMP_CONFIG" echoargs hello
  [ "$status" -eq 0 ]
  [[ "$output" == *"===> Ahoy echoargs"* ]]
  [[ "$output" == *"hello"* ]]
}

@test "Ahoy's own flags after the command name belong to the command" {
  run ./ahoy -f "$TMP_CONFIG" echoargs -f other.yml --verbose -v
  [ "$status" -eq 0 ]
  [[ "$output" != *"===> Ahoy"* ]]
  [ "$output" = "-f other.yml --verbose -v" ]
}

@test "--version after the command name reaches the command" {
  # The wrapped tool's own --version must win, or `ahoy node --version`
  # reports ahoy's version instead of Node's.
  run ./ahoy -f "$TMP_CONFIG" echoargs --version
  [ "$status" -eq 0 ]
  [ "$output" = "--version" ]
}

@test "--version before the command name is still ahoy's" {
  run ./ahoy --version
  [ "$status" -eq 0 ]
  [[ "$output" != *"--version"* ]]
}

@test "Every global flag form is consumed before the command's arguments begin" {
  # commandArgs is sliced at the first non-flag token, so each spelling of a
  # global flag has to be counted correctly - miscount by one and the config
  # path leaks into the command's arguments, or an argument goes missing.
  for flag in "-f" "-file" "--file"; do
    run ./ahoy "$flag" "$TMP_CONFIG" echoargs hello
    [ "$status" -eq 0 ]
    [ "$output" = "hello" ]

    run ./ahoy "$flag=$TMP_CONFIG" echoargs hello
    [ "$status" -eq 0 ]
    [ "$output" = "hello" ]
  done
}

@test "A global flag in '=' form still leaves the command's arguments intact" {
  run ./ahoy -f="$TMP_CONFIG" --verbose=true echoargs one --two three
  [ "$status" -eq 0 ]
  [[ "$output" == *"===> Ahoy echoargs"* ]]
  [[ "$output" == *"one --two three"* ]]
}

@test "A leading '--' is consumed by ahoy" {
  # The habit from issue #100. Ahoy no longer parses a command's arguments,
  # so this is only a readability convention now, but scripts still use it.
  run ./ahoy -f "$TMP_CONFIG" echoargs -- --build
  [ "$status" -eq 0 ]
  [ "$output" = "--build" ]
}

@test "A later '--' is the command's own data" {
  # Issue #186: kubectl exec, npm run, cargo run and git checkout all need
  # a '--' of their own, and stripping it silently changed the request.
  run ./ahoy -f "$TMP_CONFIG" echoargs exec mypod -- ls /app
  [ "$status" -eq 0 ]
  [ "$output" = "exec mypod -- ls /app" ]
}

@test "A leading '--' is consumed and a later one is kept" {
  run ./ahoy -f "$TMP_CONFIG" echoargs -- run test -- --watch
  [ "$status" -eq 0 ]
  [ "$output" = "run test -- --watch" ]
}

@test "Only the first '--' is consumed, not every leading one" {
  run ./ahoy -f "$TMP_CONFIG" echoargs -- -- a
  [ "$status" -eq 0 ]
  [ "$output" = "-- a" ]
}

@test "A '--' as the only argument reaches the command as nothing" {
  run ./ahoy -f "$TMP_CONFIG" echoargs --
  [ "$status" -eq 0 ]
  [ "$output" = "" ]
}

@test "A command that is only a '--' plus data keeps the data intact" {
  run ./ahoy -f "$TMP_CONFIG" echoargs checkout -- deploy
  [ "$status" -eq 0 ]
  [ "$output" = "checkout -- deploy" ]
}

@test "Imported subcommands get the same verbatim pass-through" {
  cat > "$BATS_TEST_TMPDIR/sub.ahoy.yml" <<'YAML'
ahoyapi: v2
commands:
  leaf:
    usage: Echo the arguments it receives
    cmd: echo "$@"
YAML
  cat > "$TMP_CONFIG" <<'YAML'
ahoyapi: v2
commands:
  group:
    usage: A group of imported commands
    imports:
      - sub.ahoy.yml
YAML

  run ./ahoy -f "$TMP_CONFIG" group leaf AAA --log-junit x.xml BBB
  [ "$status" -eq 0 ]
  [ "$output" = "AAA --log-junit x.xml BBB" ]
}
