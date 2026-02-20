@wip
Feature: Homebrew Installation
  As a new user
  I want to install Canopy via Homebrew
  So that I can start using AI memory features quickly

  @smoke @critical
  Scenario: Fresh install on macOS
    Given I have Homebrew installed
    And I have not previously installed Canopy
    When I run "brew tap canopyhq/tap"
    Then the command should succeed
    When I run "brew install canopy"
    Then the command should succeed
    And the "canopy" binary should be in my PATH
    And running "canopy version" should output version information

  @critical
  Scenario: Verify binary is executable
    Given Canopy is installed via Homebrew
    When I run "canopy version"
    Then the command should succeed
    And the output should match pattern "canopy v\d+\.\d+\.\d+"

  @critical
  Scenario: Verify data directory creation
    Given Canopy is installed via Homebrew
    When I run "canopy status"
    Then the directory "~/.phloem" should exist
    And the file "~/.phloem/memories.db" should exist

  Scenario: Upgrade from previous version
    Given Canopy version "1.0.0" is installed
    And version "1.0.1" is available in the tap
    When I run "brew upgrade canopy"
    Then the command should succeed
    And running "canopy version" should show "1.0.1"
    And my existing memories should be preserved

  Scenario: Reinstall after uninstall
    Given Canopy was previously installed and uninstalled
    And "~/.phloem" directory still exists with data
    When I run "brew install canopy"
    Then the command should succeed
    And my existing memories should be accessible

  @critical
  Scenario: Install on Apple Silicon (arm64)
    Given I am on macOS with Apple Silicon
    When I install Canopy via Homebrew
    Then the binary should be native arm64
    And running "file $(which canopy)" should show "arm64"

  @critical
  Scenario: Install on Intel Mac (amd64)
    Given I am on macOS with Intel processor
    When I install Canopy via Homebrew
    Then the binary should be native x86_64
    And running "file $(which canopy)" should show "x86_64"

  # Unhappy Paths

  @critical
  Scenario: Install without tapping first
    Given I have not tapped canopyhq/tap
    When I run "brew install canopy"
    Then the command should fail
    And the error should mention "No available formula"

  Scenario: Install with network error
    Given the network is unavailable
    When I run "brew install canopyhq/tap/canopy"
    Then the command should fail
    And the error should mention network or connection

  Scenario: Install with insufficient disk space
    Given disk space is below 100MB
    When I run "brew install canopy"
    Then the command should fail
    And the error should mention disk space

  Scenario: Homebrew not installed
    Given Homebrew is not installed
    When I try to run "brew install canopy"
    Then the command should fail with "command not found"
