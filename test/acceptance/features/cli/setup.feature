@wip
Feature: Canopy Setup Command
  As a Canopy user
  I want to configure my IDE for MCP
  So that Canopy memory works in my editor

  Background:
    Given Canopy is installed

  @smoke @critical
  Scenario: Setup Cursor when config does not exist
    Given no MCP config exists
    When I run "canopy setup cursor"
    Then the command should succeed
    And the Cursor MCP config should include Canopy

  @critical
  Scenario: Setup Cursor when config exists
    Given Cursor MCP config exists
    When I run "canopy setup cursor"
    Then the command should succeed
    And the existing config should be preserved or merged

  Scenario: Setup Windsurf
    Given no MCP config exists for Windsurf
    When I run "canopy setup windsurf"
    Then the command should succeed
    And the Windsurf MCP config should include Canopy

  Scenario: Setup shows warning when IDE not found
    Given Cursor is not installed
    When I run "canopy setup cursor"
    Then the command should show a warning
    And the output should suggest installing Cursor or checking PATH

  # Unhappy paths

  Scenario: Setup with invalid IDE name
    When I run "canopy setup unknown-ide"
    Then the command should fail
    And the error should mention unsupported or unknown

  Scenario: Setup with permission denied on config dir
    Given I do not have write access to the MCP config directory
    When I run "canopy setup cursor"
    Then the command should fail
    And the error should mention permission
