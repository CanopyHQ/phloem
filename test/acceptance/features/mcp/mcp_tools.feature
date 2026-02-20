@wip
Feature: MCP Tools Discovery and Invocation
  As an MCP client
  I want to discover and call Phloem tools
  So that I can store and recall memories

  Background:
    Given the Phloem MCP server is running
    And the memory store is initialized

  @smoke @critical
  Scenario: List tools includes remember and recall
    When I request the list of available MCP tools
    Then I should receive a list containing "remember"
    And I should receive a list containing "recall"

  @critical
  Scenario: List tools includes core memory tools
    When I request the list of available MCP tools
    Then I should receive a list containing "forget"
    And I should receive a list containing "list_memories"
    And I should receive a list containing "memory_stats"

  @critical
  Scenario: Remember tool stores content
    When I call the MCP tool "remember" with content "Test memory for tools"
    Then I should receive a success response
    And the response should contain a memory ID

  @critical
  Scenario: Recall tool returns matching memories
    Given I have stored a memory with content "Unique searchable content"
    When I call the MCP tool "recall" with query "searchable"
    Then the results should contain "Unique searchable content"

  Scenario: Memory_stats returns totals
    Given I have stored 5 memories
    When I call the MCP tool "memory_stats"
    Then I should receive memory statistics
    And the response should contain total_memories

  Scenario: List_memories returns recent items
    Given I have stored 10 memories
    When I call the MCP tool "list_memories" with limit 3
    Then I should receive a list of recent memories
    And I should receive exactly 3 or fewer memories

  # Error cases

  Scenario: Invalid tool name returns error
    When I call the MCP tool "nonexistent_tool"
    Then I should receive an error response

  Scenario: Remember with empty content
    When I call the MCP tool "remember" with content ""
    Then I should receive an error response or validation message
