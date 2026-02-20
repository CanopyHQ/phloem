@critical
Feature: IDE Setup
  As a developer
  I want to configure Phloem for my IDE
  So that my AI assistant has persistent memory

  Scenario: Setup auto-detects installed IDEs
    Given phloem is installed
    When I run "phloem setup"
    Then it should detect installed IDEs

  Scenario: Setup configures Cursor
    Given phloem is installed
    When I run "phloem setup cursor"
    Then Cursor MCP config should contain phloem

  Scenario: Setup configures Windsurf
    Given phloem is installed
    When I run "phloem setup windsurf"
    Then Windsurf MCP config should contain phloem
