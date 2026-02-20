@wip
Feature: Phloem Doctor Command
  As a Phloem user
  I want to diagnose setup issues
  So that I can quickly identify and fix problems

  Background:
    Given Phloem is installed

  @smoke @critical
  Scenario: Doctor with healthy setup
    Given the phloem binary is in PATH
    And the data directory exists
    And the SQLite database is valid
    When I run "phloem doctor"
    Then the command should succeed
    And the output should show all checks passing

  @critical
  Scenario: Doctor checks binary in PATH
    When I run "phloem doctor"
    Then the output should contain "Checking if phloem is in PATH"
    And the check should show the binary location

  @critical
  Scenario: Doctor checks binary permissions
    When I run "phloem doctor"
    Then the output should contain "Checking binary permissions"
    And the check should verify executable permission

  @critical
  Scenario: Doctor checks data directory
    When I run "phloem doctor"
    Then the output should contain "Checking data directory"
    And the check should verify "~/.phloem" exists

  @critical
  Scenario: Doctor checks MCP configuration
    Given Cursor MCP config exists
    When I run "phloem doctor"
    Then the output should contain "Checking Cursor MCP configuration"
    And the check should show OK

  Scenario: Doctor warns about missing MCP config
    Given no MCP config exists
    When I run "phloem doctor"
    Then the output should contain "WARNING"
    And the output should suggest running "phloem setup cursor"

  @critical
  Scenario: Doctor checks SQLite database
    When I run "phloem doctor"
    Then the output should contain "Checking SQLite database"
    And the check should verify database integrity

  @critical
  Scenario: Doctor tests MCP server startup
    When I run "phloem doctor"
    Then the output should contain "Testing MCP server startup"
    And the check should verify server starts successfully

  Scenario: Doctor checks environment
    When I run "phloem doctor"
    Then the output should contain "Checking environment"
    And the output should show OS and architecture
    And on Apple Silicon it should show "Apple Silicon native"

  Scenario: Doctor detects Rosetta emulation
    Given I am running under Rosetta emulation
    When I run "phloem doctor"
    Then the output should show a warning about Rosetta
    And the output should suggest installing native arm64 binary

  # Unhappy Paths

  Scenario: Doctor with corrupted database
    Given the SQLite database is corrupted
    When I run "phloem doctor"
    Then the database check should show ERROR
    And the output should suggest recovery steps

  Scenario: Doctor with locked database
    Given another process has locked the database
    When I run "phloem doctor"
    Then the database check should show WARNING
    And the output should mention database lock

  Scenario: Doctor with missing data directory
    Given "~/.phloem" does not exist
    When I run "phloem doctor"
    Then the data directory check should show WARNING
    And the output should suggest running any phloem command to create it

  Scenario: Doctor with permission issues
    Given "~/.phloem" has incorrect permissions
    When I run "phloem doctor"
    Then the check should show ERROR
    And the output should suggest fixing permissions

  Scenario: Doctor summary with issues
    Given there are 2 warnings and 1 error
    When I run "phloem doctor"
    Then the summary should show "Found 1 error(s) and 2 warning(s)"
    And the exit code should be non-zero
