#!/usr/bin/env bats

# Windows-specific tests for binary execution and Windows issue #147

@test "Binary executes without hanging on Windows" {
  if [[ "$OSTYPE" != "msys" && "$OSTYPE" != "cygwin" ]]; then
    skip "Windows-specific test"
  fi
  
  # Test that the binary doesn't hang and can display version
  timeout 10s ./ahoy.exe --version
  [ $? -eq 0 ]
}

@test "Binary responds to help command on Windows" {
  if [[ "$OSTYPE" != "msys" && "$OSTYPE" != "cygwin" ]]; then
    skip "Windows-specific test"
  fi
  
  run timeout 10s ./ahoy.exe --help
  [ $status -eq 0 ]
  [[ "$output" == *"Creates a configurable cli app"* ]]
}

@test "Binary can execute simple commands on Windows" {
  if [[ "$OSTYPE" != "msys" && "$OSTYPE" != "cygwin" ]]; then
    skip "Windows-specific test"
  fi
  
  run timeout 10s ./ahoy.exe -f testdata/simple.ahoy.yml echo "hello windows"
  [ $status -eq 0 ]
  [[ "$output" == *"hello windows"* ]]
}

@test "Binary can read config files with Windows paths" {
  if [[ "$OSTYPE" != "msys" && "$OSTYPE" != "cygwin" ]]; then
    skip "Windows-specific test"
  fi
  
  # Create a test config with Windows-style paths
  cat > testdata/windows-paths.ahoy.yml << 'EOF'
ahoyapi: v2
commands:
  test-windows-path:
    usage: Test Windows path handling
    cmd: echo "Windows path test"
EOF
  
  run timeout 10s ./ahoy.exe -f testdata/windows-paths.ahoy.yml test-windows-path
  [ $status -eq 0 ]
  [[ "$output" == *"Windows path test"* ]]
  
  # Clean up
  rm -f testdata/windows-paths.ahoy.yml
}

@test "Binary handles Windows environment variables correctly" {
  if [[ "$OSTYPE" != "msys" && "$OSTYPE" != "cygwin" ]]; then
    skip "Windows-specific test"
  fi
  
  # Set a Windows-style environment variable
  export WINDOWS_TEST_VAR="windows_value"
  
  # Create a test config that uses the environment variable
  cat > testdata/windows-env.ahoy.yml << 'EOF'
ahoyapi: v2
commands:
  test-windows-env:
    usage: Test Windows environment variables
    cmd: echo $WINDOWS_TEST_VAR
EOF
  
  run timeout 10s ./ahoy.exe -f testdata/windows-env.ahoy.yml test-windows-env
  [ $status -eq 0 ]
  [[ "$output" == *"windows_value"* ]]
  
  # Clean up
  rm -f testdata/windows-env.ahoy.yml
  unset WINDOWS_TEST_VAR
}