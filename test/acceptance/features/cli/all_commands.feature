@wip
Feature: All CLI Commands
  As a Phloem user
  I want all CLI commands to work correctly
  So that I can use the full feature set

  Background:
    Given Phloem is installed
    And the memory store is initialized

  # ============================================
  # CORE MCP COMMANDS
  # ============================================

  @smoke @critical
  Scenario: phloem serve - Start MCP server
    When I run "phloem serve" in background
    And I send an MCP initialize request
    Then I should receive a valid MCP response
    And the server should be running

  Scenario: phloem serve - Port already in use
    Given another process is using the MCP stdin/stdout
    When I run "phloem serve"
    Then the command should handle the conflict gracefully

  # ============================================
  # MEMORY MANAGEMENT COMMANDS
  # ============================================

  @smoke @critical
  Scenario: phloem export - Export as JSON
    Given I have stored 10 memories
    When I run "phloem export json /tmp/memories.json"
    Then the command should succeed
    And the file "/tmp/memories.json" should exist
    And the file should contain valid JSON with 10 memories

  Scenario: phloem export - Export as Markdown
    Given I have stored 5 memories
    When I run "phloem export markdown /tmp/memories.md"
    Then the command should succeed
    And the file "/tmp/memories.md" should exist
    And the file should contain markdown formatted memories

  Scenario: phloem export - Invalid format
    When I run "phloem export xml /tmp/memories.xml"
    Then the command should fail
    And the error should mention unsupported format

  Scenario: phloem export - Permission denied
    When I run "phloem export json /root/memories.json"
    Then the command should fail
    And the error should mention permission denied

  @critical
  Scenario: phloem verify - Verify citations
    Given I have a memory with citations
    When I run "phloem verify <memory_id>"
    Then the command should succeed
    And the output should show citation verification results

  Scenario: phloem verify - Invalid memory ID
    When I run "phloem verify nonexistent123"
    Then the command should fail
    And the error should mention memory not found

  Scenario: phloem verify - Memory without citations
    Given I have a memory without citations
    When I run "phloem verify <memory_id>"
    Then the command should succeed
    And the output should indicate no citations to verify

  Scenario: phloem decay - Decay confidence scores
    Given I have memories with citations
    When I run "phloem decay"
    Then the command should succeed
    And citation confidence scores should be reduced based on age

  Scenario: phloem decay - Empty database
    Given no memories exist
    When I run "phloem decay"
    Then the command should succeed
    And the output should indicate nothing to decay

  # ============================================
  # GRAFT COMMANDS
  # ============================================

  @smoke @critical
  Scenario: phloem graft package - Create graft
    Given I have 5 memories tagged "architecture"
    When I run "phloem graft package architecture /tmp/arch.graft"
    Then the command should succeed
    And the file "/tmp/arch.graft" should exist

  @smoke @critical
  Scenario: phloem graft unpack - Import graft
    Given I have a valid graft file "/tmp/test.graft"
    When I run "phloem graft unpack /tmp/test.graft"
    Then the command should succeed
    And the memories should be imported

  Scenario: phloem graft inspect - View graft contents
    Given I have a valid graft file "/tmp/test.graft"
    When I run "phloem graft inspect /tmp/test.graft"
    Then the command should succeed
    And the output should show graft metadata
    And no memories should be imported

  Scenario: phloem graft - Invalid file
    When I run "phloem graft unpack /tmp/invalid.graft"
    Then the command should fail
    And the error should mention invalid graft format

  Scenario: phloem graft - File not found
    When I run "phloem graft unpack /tmp/nonexistent.graft"
    Then the command should fail
    And the error should mention file not found

  # ============================================
  # IMPORT COMMANDS
  # ============================================

  @critical
  Scenario: phloem import chatgpt - Import ChatGPT history
    Given I have a valid ChatGPT export file "conversations.json"
    When I run "phloem import chatgpt conversations.json"
    Then the command should succeed
    And conversations should be imported as memories

  Scenario: phloem import chatgpt - Invalid JSON
    Given I have an invalid JSON file "invalid.json"
    When I run "phloem import chatgpt invalid.json"
    Then the command should fail
    And the error should mention invalid JSON

  Scenario: phloem import chatgpt - Wrong format
    Given I have a JSON file with wrong structure
    When I run "phloem import chatgpt wrong.json"
    Then the command should fail
    And the error should mention unexpected format

  @critical
  Scenario: phloem import claude - Import Claude history
    Given I have a valid Claude export directory
    When I run "phloem import claude ./claude-export/"
    Then the command should succeed
    And conversations should be imported as memories

  Scenario: phloem import claude - Invalid directory
    When I run "phloem import claude /nonexistent/"
    Then the command should fail
    And the error should mention directory not found

  # ============================================
  # SUPPORT COMMANDS
  # ============================================

  @smoke @critical
  Scenario: phloem version - Show version
    When I run "phloem version"
    Then the command should succeed
    And the output should match "phloem v\d+\.\d+\.\d+"

  Scenario: phloem help - Show help
    When I run "phloem help"
    Then the command should succeed
    And the output should list available commands

  Scenario: phloem help <command> - Show command help
    When I run "phloem help status"
    Then the command should succeed
    And the output should show status command usage

  # ============================================
  # ERROR HANDLING
  # ============================================

  @critical
  Scenario: Unknown command
    When I run "phloem unknowncommand"
    Then the command should fail
    And the error should mention unknown command
    And the error should suggest "phloem help"

  @critical
  Scenario: Unknown subcommand
    When I run "phloem graft unknownsub"
    Then the command should fail
    And the error should mention unknown subcommand

  @critical
  Scenario: Missing required argument
    When I run "phloem export"
    Then the command should fail
    And the error should mention missing argument

  Scenario: Invalid flag
    When I run "phloem status --invalid-flag"
    Then the command should fail
    And the error should mention unknown flag
