#!/usr/bin/env bats

load 'test_helpers/bats-support/load'
load 'test_helpers/bats-assert/load'

setup() {
    # Create temporary directory for test files
    export AHOY_TEST_DIR=$(mktemp -d)
    export ORIGINAL_DIR=$(pwd)
    cd "$AHOY_TEST_DIR"
    
    # Ensure ahoy binary is available
    if [[ ! -f "$ORIGINAL_DIR/ahoy" ]]; then
        cd "$ORIGINAL_DIR"
        make
        cd "$AHOY_TEST_DIR"
    fi
    export AHOY_BIN="$ORIGINAL_DIR/ahoy"
}

teardown() {
    # Clean up temporary directory
    if [[ -n "$AHOY_TEST_DIR" && -d "$AHOY_TEST_DIR" ]]; then
        rm -rf "$AHOY_TEST_DIR"
    fi
}

@test "Single env file syntax (backwards compatibility)" {
    # Create a simple .env file
    cat > .env << 'EOF'
SINGLE_ENV_VAR=test_value
GLOBAL_VAR=from_single
EOF

    # Create .ahoy.yml with single env file syntax
    cat > .ahoy.yml << 'EOF'
ahoyapi: v2
env: .env
commands:
  test-env:
    usage: Test environment variables
    cmd: echo "SINGLE_ENV_VAR=$SINGLE_ENV_VAR GLOBAL_VAR=$GLOBAL_VAR"
EOF

    run "$AHOY_BIN" test-env
    assert_success
    assert_output --partial "SINGLE_ENV_VAR=test_value"
    assert_output --partial "GLOBAL_VAR=from_single"
}

@test "Multiple env files array syntax" {
    # Create multiple env files
    cat > .env.base << 'EOF'
BASE_VAR=base_value
OVERRIDE_VAR=from_base
EOF

    cat > .env.local << 'EOF'
LOCAL_VAR=local_value
OVERRIDE_VAR=from_local
EOF

    # Create .ahoy.yml with multiple env files
    cat > .ahoy.yml << 'EOF'
ahoyapi: v2
env:
  - .env.base
  - .env.local
commands:
  test-env:
    usage: Test environment variables
    cmd: echo "BASE_VAR=$BASE_VAR LOCAL_VAR=$LOCAL_VAR OVERRIDE_VAR=$OVERRIDE_VAR"
EOF

    run "$AHOY_BIN" test-env
    assert_success
    assert_output --partial "BASE_VAR=base_value"
    assert_output --partial "LOCAL_VAR=local_value"
    # Later files should override earlier ones
    assert_output --partial "OVERRIDE_VAR=from_local"
}

@test "Environment variable precedence with multiple files" {
    # Create base environment file
    cat > .env.defaults << 'EOF'
VAR_A=default_a
VAR_B=default_b
VAR_C=default_c
EOF

    # Create override environment file
    cat > .env.overrides << 'EOF'
VAR_B=override_b
VAR_C=override_c
EOF

    # Create .ahoy.yml with multiple env files
    cat > .ahoy.yml << 'EOF'
ahoyapi: v2
env:
  - .env.defaults
  - .env.overrides
commands:
  test-precedence:
    usage: Test environment variable precedence
    cmd: echo "A=$VAR_A B=$VAR_B C=$VAR_C"
EOF

    run "$AHOY_BIN" test-precedence
    assert_success
    assert_output --partial "A=default_a"
    assert_output --partial "B=override_b"
    assert_output --partial "C=override_c"
}

@test "Command-level multiple env files" {
    # Create global env file
    cat > .env.global << 'EOF'
GLOBAL_VAR=global_value
OVERRIDE_VAR=from_global
EOF

    # Create command-specific env files
    cat > .env.cmd1 << 'EOF'
CMD1_VAR=cmd1_value
OVERRIDE_VAR=from_cmd1
EOF

    cat > .env.cmd2 << 'EOF'
CMD2_VAR=cmd2_value
OVERRIDE_VAR=from_cmd2
EOF

    # Create .ahoy.yml with global and command-level env files
    cat > .ahoy.yml << 'EOF'
ahoyapi: v2
env: .env.global
commands:
  test-cmd1:
    usage: Test command 1 env vars
    env:
      - .env.cmd1
    cmd: echo "GLOBAL=$GLOBAL_VAR CMD1=$CMD1_VAR OVERRIDE=$OVERRIDE_VAR"
  test-cmd2:
    usage: Test command 2 env vars
    env:
      - .env.cmd2
    cmd: echo "GLOBAL=$GLOBAL_VAR CMD2=$CMD2_VAR OVERRIDE=$OVERRIDE_VAR"
EOF

    # Test command 1
    run "$AHOY_BIN" test-cmd1
    assert_success
    assert_output --partial "GLOBAL=global_value"
    assert_output --partial "CMD1=cmd1_value"
    assert_output --partial "OVERRIDE=from_cmd1"

    # Test command 2
    run "$AHOY_BIN" test-cmd2
    assert_success
    assert_output --partial "GLOBAL=global_value"
    assert_output --partial "CMD2=cmd2_value"
    assert_output --partial "OVERRIDE=from_cmd2"
}

@test "Mixed global array and command single env file" {
    # Create global env files
    cat > .env.global1 << 'EOF'
GLOBAL1=value1
MIXED_VAR=from_global1
EOF

    cat > .env.global2 << 'EOF'
GLOBAL2=value2
MIXED_VAR=from_global2
EOF

    # Create command env file
    cat > .env.command << 'EOF'
COMMAND_VAR=command_value
MIXED_VAR=from_command
EOF

    # Create .ahoy.yml with array global env and single command env
    cat > .ahoy.yml << 'EOF'
ahoyapi: v2
env:
  - .env.global1
  - .env.global2
commands:
  test-mixed:
    usage: Test mixed env file configurations
    env: .env.command
    cmd: echo "G1=$GLOBAL1 G2=$GLOBAL2 CMD=$COMMAND_VAR MIXED=$MIXED_VAR"
EOF

    run "$AHOY_BIN" test-mixed
    assert_success
    assert_output --partial "G1=value1"
    assert_output --partial "G2=value2"
    assert_output --partial "CMD=command_value"
    # Command env should override global env
    assert_output --partial "MIXED=from_command"
}

@test "Non-existent env files are gracefully handled" {
    # Create .ahoy.yml with non-existent env files
    cat > .ahoy.yml << 'EOF'
ahoyapi: v2
env:
  - .env.missing
  - .env.also-missing
commands:
  test-missing:
    usage: Test missing env files
    cmd: echo "TEST=success"
EOF

    run "$AHOY_BIN" test-missing
    assert_success
    assert_output --partial "TEST=success"
}

@test "Empty env array is handled correctly" {
    # Create .ahoy.yml with empty env array
    cat > .ahoy.yml << 'EOF'
ahoyapi: v2
env: []
commands:
  test-empty:
    usage: Test empty env array
    cmd: echo "EMPTY_TEST=success"
EOF

    run "$AHOY_BIN" test-empty
    assert_success
    assert_output --partial "EMPTY_TEST=success"
}

@test "Environment files with comments and empty lines" {
    # Create env file with comments and empty lines
    cat > .env.complex << 'EOF'
# This is a comment
VALID_VAR=valid_value

# Another comment with empty line above
ANOTHER_VAR=another_value
# Comment at end
EOF

    # Create .ahoy.yml
    cat > .ahoy.yml << 'EOF'
ahoyapi: v2
env: .env.complex
commands:
  test-complex:
    usage: Test complex env file
    cmd: echo "VALID=$VALID_VAR ANOTHER=$ANOTHER_VAR"
EOF

    run "$AHOY_BIN" test-complex
    assert_success
    assert_output --partial "VALID=valid_value"
    assert_output --partial "ANOTHER=another_value"
}

@test "Relative paths work correctly with multiple env files" {
    # Create subdirectory structure
    mkdir -p config
    
    # Create env files in subdirectory
    cat > config/.env.sub << 'EOF'
SUBDIR_VAR=subdir_value
EOF

    cat > .env.root << 'EOF'
ROOT_VAR=root_value
EOF

    # Create .ahoy.yml with relative paths
    cat > .ahoy.yml << 'EOF'
ahoyapi: v2
env:
  - .env.root
  - config/.env.sub
commands:
  test-paths:
    usage: Test relative paths
    cmd: echo "ROOT=$ROOT_VAR SUBDIR=$SUBDIR_VAR"
EOF

    run "$AHOY_BIN" test-paths
    assert_success
    assert_output --partial "ROOT=root_value"
    assert_output --partial "SUBDIR=subdir_value"
}

@test "Large number of environment files" {
    # Create multiple env files
    for i in {1..5}; do
        cat > ".env.file$i" << EOF
VAR_$i=value_$i
EOF
    done

    # Create .ahoy.yml with all env files
    cat > .ahoy.yml << 'EOF'
ahoyapi: v2
env:
  - .env.file1
  - .env.file2
  - .env.file3
  - .env.file4
  - .env.file5
commands:
  test-many:
    usage: Test many env files
    cmd: echo "1=$VAR_1 2=$VAR_2 3=$VAR_3 4=$VAR_4 5=$VAR_5"
EOF

    run "$AHOY_BIN" test-many
    assert_success
    assert_output --partial "1=value_1"
    assert_output --partial "2=value_2"
    assert_output --partial "3=value_3"
    assert_output --partial "4=value_4"
    assert_output --partial "5=value_5"
}