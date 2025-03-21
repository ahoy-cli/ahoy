#!/usr/bin/env bats

@test "Command-level variables can be defined and used" {
    run ./ahoy -f testdata/env.ahoy.yml test-cmd
    [[ "$output" == "123456789" ]]
}

@test "Environment variables can be overriden" {
    run ./ahoy -f testdata/env.ahoy.yml test-override
    [[ "$output" = "after" ]]
}

@test "Global variables can be defined and used" {
    run ./ahoy -f testdata/env.ahoy.yml test-global
    [[ "$output" = "global" ]]
}

@test "Fail when attempting to load invalid env files" {
    run ./ahoy -f testdata/env.ahoy.yml test-invalid-env
    [ $status -eq 1 ]
}

@test "Existing environment variables are not clobbered by .env file loading" {
    # This variable should be kept and be available to an ahoy command.
    export ENV_CLOBBER_TEST=1234
    run ./ahoy -f testdata/env.ahoy.yml test-keep-established-env-vars
    [[ "$output" = "1234" ]]
}
