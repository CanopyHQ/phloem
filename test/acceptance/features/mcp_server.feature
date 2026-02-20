Feature: MCP Server Protocol Compliance
  As a developer using Cursor IDE
  I want Phloem to expose memory features via MCP
  So that I can store and recall memories directly in Cursor

  Background:
    Given the Phloem MCP server is running

  @smoke @critical @brew_gate
  Scenario: MCP server initialization
    When I send an initialize request to the MCP server
    Then I should receive a valid initialization response
    And the response should contain protocol version "2024-11-05"
    And the response should contain server name "phloem"

  @smoke @critical @brew_gate
  Scenario: List available MCP tools
    When I request the list of available MCP tools
    Then I should receive a list containing "remember"
    And I should receive a list containing "recall"
    And I should receive a list containing "forget"
    And I should receive a list containing "list_memories"
    And I should receive a list containing "memory_stats"
    And I should receive a list containing "session_context"

  @smoke @brew_gate
  Scenario: List available MCP resources
    When I request the list of available MCP resources
    Then I should receive a list containing "phloem://memories/recent"
    And I should receive a list containing "phloem://memories/stats"
    And I should receive a list containing "phloem://context/session"

  @wip
  Scenario: Read recent memories resource
    Given I have stored 5 memories
    When I read the MCP resource "phloem://memories/recent"
    Then I should receive a list of recent memories
    And the response should be valid JSON

  @wip
  Scenario: Read stats resource
    When I read the MCP resource "phloem://memories/stats"
    Then I should receive memory statistics
    And the response should contain total_memories
    And the response should contain database_size

  @wip
  Scenario: Read session context resource
    Given I have stored memories with various tags
    When I read the MCP resource "phloem://context/session"
    Then I should receive formatted markdown context
    And the context should include recent memories
    And the context should include tagged sections

  @critical @wip
  Scenario: Handle invalid tool call
    When I call the MCP tool "nonexistent_tool"
    Then I should receive an error response
    And the error code should be -32602
    And the error should indicate unknown tool

  @critical @wip
  Scenario: Handle invalid resource
    When I read the MCP resource "phloem://invalid/resource"
    Then I should receive an error response
    And the error should indicate unknown resource

  @wip
  Scenario: Tool input schema validation
    When I request the list of available MCP tools
    Then the "remember" tool should have required parameter "content"
    And the "recall" tool should have required parameter "query"
    And the "forget" tool should have required parameter "id"

  @critical @wip
  Scenario: JSON-RPC error handling
    When I send a malformed JSON-RPC request
    Then I should receive error code -32700 (Parse error)

  @wip
  Scenario: Concurrent MCP requests
    When I send 10 concurrent remember requests
    Then all requests should complete successfully
    And all memories should be stored correctly
