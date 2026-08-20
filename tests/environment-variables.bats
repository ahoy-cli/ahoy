#!/usr/bin/env bats

@test "Command-level variables can be defined and used" {
    run ./ahoy -f testdata/env.ahoy.yml test-cmd
    [[ "$output" == "123456789" ]]
}

@test "Environment variables can be overriden" {
    run ./ahoy -f testdata/env.ahoy.yml test-override
    [[ "$output" = "after" ]]
}

@test "Command level environment variables can be overriden by another env file" {
    run ./ahoy -f testdata/env.ahoy.yml test-multiple-cmd-overridden
    [[ "$output" = "after" ]]
}

@test "Global variables can be defined and used" {
    run ./ahoy -f testdata/env.ahoy.yml test-global
    [[ "$output" = "global" ]]
}

@test "Command level environment variable loading tolerates missing env file." {
    run ./ahoy -f testdata/env.ahoy.yml test-multiple-with-missing-cmd
    [[ "$output" = "123456789" ]]
}

@test "Command level environment variable loading tolerates missing env file and can be overriden by another env file" {
    run ./ahoy -f testdata/env.ahoy.yml test-multiple-with-missing-cmd-and-override
    [[ "$output" = "987654321" ]]
}

@test "Allow non-existent env files" {
    run ./ahoy -f testdata/env.ahoy.yml test-nonexistent-env
    [ $status -eq 0 ]
}

@test "Multiple global env files can be defined" {
    run ./ahoy -f testdata/env-multiple.ahoy.yml test-global
    [[ "$output" = "global" ]]

    run ./ahoy -f testdata/env-multiple.ahoy.yml test-command
    [[ "$output" = "123456789" ]]

    run ./ahoy -f testdata/env-multiple.ahoy.yml test-overridden
    [[ "$output" = "after" ]]
}

@test "Multiple command env files can be defined" {
    run ./ahoy -f testdata/env-multiple.ahoy.yml test-cmd-multiple-1
    echo $output
    [[ "$output" = "unique" ]]

    run ./ahoy -f testdata/env-multiple.ahoy.yml test-cmd-multiple-2
    echo $output
    [[ "$output" = "local2" ]]
}

@test "Existing environment variables are not clobbered by .env file loading" {
    # This variable should be kept and be available to an ahoy command.
    export ENV_CLOBBER_TEST=1234
    run ./ahoy -f testdata/env.ahoy.yml test-keep-established-env-vars
    [[ "$output" = "1234" ]]
}

@test "Global variables can be overridden by another global env file" {
    run ./ahoy -f testdata/env-multiple-global.ahoy.yml test-global-multiple
    [[ "$output" = "global-two" ]]
}

@test "Global variables can be overridden by command env file" {
    run ./ahoy -f testdata/env-multiple-global.ahoy.yml test-cmd-multiple-override
    [[ "$output" = "999" ]]
}

@test "AHOY_CMD is set to the path of the running ahoy binary" {
    run ./ahoy -f testdata/ahoy-self-env.ahoy.yml show-ahoy-cmd
    [ $status -eq 0 ]
    [[ "$output" != "" ]]
}

@test "AHOY_CMD points to an executable file" {
    run ./ahoy -f testdata/ahoy-self-env.ahoy.yml ahoy-cmd-is-executable
    [ $status -eq 0 ]
}

@test "AHOY_CMD is not set to a literal 'ahoy' string but resolves to the binary path" {
    run ./ahoy -f testdata/ahoy-self-env.ahoy.yml show-ahoy-cmd
    [ $status -eq 0 ]
    [[ "$output" != "ahoy" ]]
}

@test "AHOY_COMMAND_NAME is set to the name of the running command" {
    run ./ahoy -f testdata/ahoy-self-env.ahoy.yml show-ahoy-command-name
    [ $status -eq 0 ]
    [[ "$output" = "show-ahoy-command-name" ]]
}

@test "AHOY_COMMAND_NAME reflects the correct command name for each command" {
    run ./ahoy -f testdata/ahoy-self-env.ahoy.yml show-both
    [ $status -eq 0 ]
    [[ "$output" == *"show-both" ]]
}

@test "AHOY_CMD and AHOY_COMMAND_NAME are both set in the same subprocess" {
    run ./ahoy -f testdata/ahoy-self-env.ahoy.yml show-both
    [ $status -eq 0 ]
    [[ "$output" != "" ]]
    # Output should contain two space-separated values.
    word_count=$(echo "$output" | wc -w | tr -d ' ')
    [[ "$word_count" -eq 2 ]]
}
