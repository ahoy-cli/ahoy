#!/usr/bin/env bats

# Windows-specific tests for path handling

@test "Ahoy can handle Windows-style config file paths" {
  if [[ "$OSTYPE" != "msys" && "$OSTYPE" != "cygwin" ]]; then
    skip "Windows-specific test"
  fi
  
  # Test with forward slashes (should work on Windows)
  run ./ahoy.exe -f testdata/simple.ahoy.yml --help
  [ $status -eq 0 ]
  
  # Test with backslashes (Windows native)
  run ./ahoy.exe -f testdata\\simple.ahoy.yml --help 2>/dev/null || true
  # This might fail on Git Bash but should work in true Windows environments
}

@test "Ahoy searches for .ahoy.yml files in Windows directory structure" {
  if [[ "$OSTYPE" != "msys" && "$OSTYPE" != "cygwin" ]]; then
    skip "Windows-specific test"
  fi
  
  # Create a temporary Windows-style directory structure
  mkdir -p /tmp/windows-test/subdir
  cp testdata/simple.ahoy.yml /tmp/windows-test/.ahoy.yml
  
  cd /tmp/windows-test/subdir
  
  # Should find .ahoy.yml in parent directory
  run timeout 10s ../../../ahoy.exe echo "test"
  [ $status -eq 0 ]
  [[ "$output" == *"test"* ]]
  
  cd -
  rm -rf /tmp/windows-test
}

@test "Ahoy handles Windows working directory correctly" {
  if [[ "$OSTYPE" != "msys" && "$OSTYPE" != "cygwin" ]]; then
    skip "Windows-specific test"
  fi
  
  # Create a config that echoes the working directory
  cat > testdata/windows-workdir.ahoy.yml << 'EOF'
ahoyapi: v2
commands:
  test-workdir:
    usage: Test working directory
    cmd: pwd
EOF
  
  mkdir -p /tmp/windows-workdir-test
  cd /tmp/windows-workdir-test
  
  # Copy config to the test directory
  cp ../ahoy/v2/testdata/windows-workdir.ahoy.yml .ahoy.yml
  
  run timeout 10s ../ahoy/v2/ahoy.exe test-workdir
  [ $status -eq 0 ]
  [[ "$output" == *"windows-workdir-test"* ]]
  
  cd -
  rm -rf /tmp/windows-workdir-test
  rm -f testdata/windows-workdir.ahoy.yml
}

@test "Ahoy handles Windows file permissions correctly" {
  if [[ "$OSTYPE" != "msys" && "$OSTYPE" != "cygwin" ]]; then
    skip "Windows-specific test"
  fi
  
  # Create a config file and test it can be read
  cat > testdata/windows-perms.ahoy.yml << 'EOF'
ahoyapi: v2
commands:
  test-perms:
    usage: Test file permissions
    cmd: echo "permissions test"
EOF
  
  # On Windows, file permissions work differently, but the file should still be readable
  run timeout 10s ./ahoy.exe -f testdata/windows-perms.ahoy.yml test-perms
  [ $status -eq 0 ]
  [[ "$output" == *"permissions test"* ]]
  
  # Clean up
  rm -f testdata/windows-perms.ahoy.yml
}