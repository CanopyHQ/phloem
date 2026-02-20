@wip
Feature: First Run Experience
  As a new Phloem user
  I want a smooth first-run experience
  So that I can start using memory features within 2 minutes

  Background:
    Given Phloem is freshly installed via Homebrew
    And no previous Phloem data exists

  @smoke @critical
  Scenario: First run creates data directory
    When I run "phloem status"
    Then the directory "~/.phloem" should be created
    And the file "~/.phloem/memories.db" should be created
    And the output should show "Total Memories: 0"

  @smoke @critical
  Scenario: Doctor command on fresh install
    When I run "phloem doctor"
    Then the command should succeed
    And the output should show "Checking if phloem is in PATH"
    And the output should show "OK" for binary check
    And the output should show "OK" for data directory
    And the output should show "OK" for SQLite database

  @critical
  Scenario: Setup Cursor on fresh install
    Given Cursor IDE is installed
    When I run "phloem setup cursor"
    Then the command should succeed
    And the file "~/.cursor/mcp.json" should exist
    And the MCP config should contain "phloem" server
    And the output should show "Phloem is now configured for Cursor"

  @critical
  Scenario: Setup Windsurf on fresh install
    Given Windsurf IDE is installed
    When I run "phloem setup windsurf"
    Then the command should succeed
    And the file "~/.windsurf/mcp_config.json" should exist
    And the MCP config should contain "phloem" server
    And the output should show "Phloem is now configured for Windsurf"

  Scenario: Setup preserves existing MCP servers
    Given Cursor IDE is installed
    And "~/.cursor/mcp.json" exists with other MCP servers
    When I run "phloem setup cursor"
    Then the command should succeed
    And the existing MCP servers should be preserved
    And "phloem" server should be added

  @critical
  Scenario: Complete first-run flow
    When I run "phloem version"
    Then I should see the version number
    When I run "phloem doctor"
    Then all checks should pass or show warnings
    When I run "phloem setup cursor"
    Then Cursor should be configured
    When I run "phloem status"
    Then I should see memory statistics

  # Unhappy Paths

  Scenario: Setup without IDE installed
    Given Cursor IDE is not installed
    When I run "phloem setup cursor"
    Then the command should show a warning
    And the output should indicate Cursor not found
    And the config file should still be created

  Scenario: Setup with invalid existing config
    Given "~/.cursor/mcp.json" exists with invalid JSON
    When I run "phloem setup cursor"
    Then the command should handle the error gracefully
    And the output should indicate config was reset or backed up

  Scenario: Doctor with missing binary in PATH
    Given the phloem binary is not in PATH
    When I run "/full/path/to/phloem doctor"
    Then the output should show a warning for PATH
    And the output should suggest adding to PATH

  Scenario: First run with read-only home directory
    Given the home directory is read-only
    When I run "phloem status"
    Then the command should fail gracefully
    And the error should mention permission denied
    And the error should suggest checking permissions
