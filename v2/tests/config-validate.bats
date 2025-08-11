#!/usr/bin/env bats

# Setup and teardown
setup() {
    # Create a temporary directory for each test
    export TEST_TEMP_DIR=$(mktemp -d)
    export ORIGINAL_DIR=$(pwd)
    cd "$TEST_TEMP_DIR"
}

teardown() {
    # Clean up
    cd "$ORIGINAL_DIR"
    rm -rf "$TEST_TEMP_DIR"
}

@test "ahoy config validate shows help" {
    run "$ORIGINAL_DIR/ahoy" config validate --help
    
    [ "$status" -eq 0 ]
    [[ "$output" =~ "NAME:" ]]
    [[ "$output" =~ "ahoy config validate" ]]
    [[ "$output" =~ "Diagnose configuration issues" ]]
}

@test "ahoy config validate warns when no config file exists" {
    run "$ORIGINAL_DIR/ahoy" config validate
    
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Warning: No .ahoy.yml file found" ]]
    [[ "$output" =~ "Run 'ahoy config init'" ]]
}

@test "ahoy config validate succeeds with valid config" {
    # Create a valid config file
    cat > .ahoy.yml << EOF
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
    [[ "$output" =~ "Configuration file: $TEST_TEMP_DIR/.ahoy.yml ✅ (found)" ]]
    [[ "$output" =~ "API Version: v2 ✅ (supported)" ]]
    [[ "$output" =~ "✅ Syntax: Valid YAML" ]]
    [[ "$output" =~ "✅ No validation issues found" ]]
    [[ "$output" =~ "Configuration looks good!" ]]
    [[ "$output" =~ "✅ Configuration looks great!" ]]
}

@test "ahoy config validate detects invalid YAML syntax" {
    # Create a config file with invalid YAML
    cat > .ahoy.yml << EOF
ahoyapi: v2
commands:
  test:
    cmd: echo "test"
    invalid_yaml: [unclosed array
EOF

    run "$ORIGINAL_DIR/ahoy" config validate
    
    [ "$status" -eq 1 ]
    [[ "$output" =~ "yaml:" ]] || [[ "$output" =~ "YAML" ]]
    [[ "$output" =~ "did not find expected" ]] || [[ "$output" =~ "syntax" ]]
}

@test "ahoy config validate detects missing API version" {
    # Create a config file without ahoyapi
    cat > .ahoy.yml << 'EOF'
usage: "Test configuration"
commands:
  test:
    usage: "Test command"
    cmd: "echo test"
EOF

    run "$ORIGINAL_DIR/ahoy" config validate
    
    [ "$status" -eq 1 ]
    [[ "$output" =~ "API version" ]] || [[ "$output" =~ "Ahoy only supports API version" ]]
}

@test "ahoy config validate detects wrong API version" {
    # Create a config file with wrong API version
    cat > .ahoy.yml << 'EOF'
ahoyapi: v1
usage: "Test configuration"
commands:
  test:
    usage: "Test command"
    cmd: "echo test"
EOF

    run "$ORIGINAL_DIR/ahoy" config validate
    
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Ahoy only supports API version 'v2', but 'v1' given" ]]
}

@test "ahoy config validate checks environment files" {
    # Create a config file that references missing env files
    cat > .ahoy.yml << EOF
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

    # Create only one of the env files
    echo "TEST=value" > .env

    run "$ORIGINAL_DIR/ahoy" config validate
    
    [ "$status" -eq 0 ]
    [[ "$output" =~ "🌍 Environment Files:" ]]
    [[ "$output" =~ ".env" ]]
    [[ "$output" =~ "missing" ]]
    [[ "$output" =~ "Consider creating missing environment files" ]]
}

@test "ahoy config validate checks import files" {
    # Create a config file that references missing import files
    cat > .ahoy.yml << EOF
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

    # Create only one of the import files
    cat > import1.ahoy.yml << EOF
ahoyapi: v2
commands:
  imported:
    cmd: echo "imported"
EOF

    run "$ORIGINAL_DIR/ahoy" config validate
    
    [ "$status" -eq 0 ]
    [[ "$output" =~ "📥 Import Files:" ]]
    [[ "$output" =~ "import1.ahoy.yml" ]]
    [[ "$output" =~ "import2.ahoy.yml" ]]
    [[ "$output" =~ "missing" ]]
    [[ "$output" =~ "Create missing import files" ]]
}

@test "ahoy config validate with complex broken configuration" {
    # Create a seriously broken config file
    cat > .ahoy.yml << 'EOF'
ahoyapi: v1
env: ".env.missing"
commands:
  broken-command:
    usage: "Broken command"
  import-test:
    imports:
      - "missing-import.ahoy.yml"
EOF

    run "$ORIGINAL_DIR/ahoy" config validate
    
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Ahoy only supports API version 'v2', but 'v1' given" ]]
    # Should show API version error
}

@test "ahoy config validate runs from subdirectory" {
    # Create config in current directory
    cat > .ahoy.yml << EOF
ahoyapi: v2
commands:
  test:
    usage: Test command
    cmd: echo "test"
EOF

    # Run from subdirectory
    mkdir subdir
    cd subdir
    
    run "$ORIGINAL_DIR/ahoy" config validate
    
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Configuration file: $TEST_TEMP_DIR/.ahoy.yml ✅ (found)" ]]
    [[ "$output" =~ "✅ Configuration looks great!" ]]
}