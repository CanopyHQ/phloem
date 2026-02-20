@wip
Feature: MCP Protocol Error Handling
  As an MCP client (Cursor/Windsurf)
  I want proper error responses from the MCP server
  So that I can handle errors gracefully

  Background:
    Given the Phloem MCP server is running

  # ============================================
  # JSON-RPC ERRORS
  # ============================================

  @critical
  Scenario: Parse error - Invalid JSON
    When I send invalid JSON to the MCP server
    Then I should receive error code -32700
    And the error message should mention "Parse error"

  @critical
  Scenario: Invalid request - Missing jsonrpc field
    When I send a request without "jsonrpc" field
    Then I should receive error code -32600
    And the error message should mention "Invalid Request"

  @critical
  Scenario: Invalid request - Wrong jsonrpc version
    When I send a request with "jsonrpc": "1.0"
    Then I should receive error code -32600
    And the error message should mention "Invalid Request"

  @critical
  Scenario: Method not found
    When I send a request with method "nonexistent/method"
    Then I should receive error code -32601
    And the error message should mention "Method not found"

  @critical
  Scenario: Invalid params - Wrong type
    When I call "tools/call" with params as a string instead of object
    Then I should receive error code -32602
    And the error message should mention "Invalid params"

  Scenario: Invalid params - Missing required field
    When I call "tools/call" without the "name" field
    Then I should receive error code -32602
    And the error message should mention required field

  # ============================================
  # TOOL ERRORS
  # ============================================

  @critical
  Scenario: Unknown tool
    When I call the MCP tool "unknown_tool_name"
    Then I should receive an error response
    And the error should indicate unknown tool

  @critical
  Scenario: Tool with missing required argument
    When I call the MCP tool "remember" without content
    Then I should receive an error response
    And the error should mention "content" is required

  Scenario: Tool with invalid argument type
    When I call the MCP tool "list_memories" with limit as string "ten"
    Then I should receive an error response
    And the error should mention invalid type

  Scenario: Tool with out-of-range argument
    When I call the MCP tool "list_memories" with limit -5
    Then I should receive an error response
    And the error should mention invalid value

  Scenario: Remember tool - Empty content
    When I call the MCP tool "remember" with content ""
    Then I should receive an error response
    And the error should mention content cannot be empty

  Scenario: Recall tool - Empty query
    When I call the MCP tool "recall" with query ""
    Then I should receive an error response
    And the error should mention query cannot be empty

  Scenario: Forget tool - Invalid ID
    When I call the MCP tool "forget" with id "not-a-valid-id"
    Then I should receive an error response
    And the error should mention invalid or not found

  Scenario: Verify tool - Memory not found
    When I call the MCP tool "verify" with id "nonexistent123"
    Then I should receive an error response
    And the error should mention memory not found

  # ============================================
  # RESOURCE ERRORS
  # ============================================

  @critical
  Scenario: Unknown resource
    When I read the MCP resource "phloem://unknown/resource"
    Then I should receive an error response
    And the error should indicate unknown resource

  Scenario: Resource with invalid URI
    When I read the MCP resource "not-a-valid-uri"
    Then I should receive an error response
    And the error should mention invalid URI

  Scenario: Resource with wrong scheme
    When I read the MCP resource "http://memories/recent"
    Then I should receive an error response
    And the error should mention unsupported scheme

  # ============================================
  # INTERNAL ERRORS
  # ============================================

  Scenario: Database error during tool call
    Given the database is locked
    When I call the MCP tool "remember" with content "test"
    Then I should receive error code -32603
    And the error message should mention "Internal error"

  Scenario: Disk full during remember
    Given the disk is full
    When I call the MCP tool "remember" with content "test"
    Then I should receive an error response
    And the error should mention storage or disk

  # ============================================
  # CONCURRENT REQUEST HANDLING
  # ============================================

  Scenario: Concurrent requests with same ID
    When I send two requests with the same ID simultaneously
    Then both should receive responses
    And responses should be correctly matched to requests

  Scenario: Request timeout
    When I send a request that takes too long
    Then the client should be able to timeout
    And the server should handle the abandoned request gracefully

  Scenario: Rapid fire requests
    When I send 100 requests in rapid succession
    Then all requests should receive responses
    And no requests should be dropped

  # ============================================
  # MALFORMED REQUESTS
  # ============================================

  Scenario: Request with extra fields
    When I send a request with extra unknown fields
    Then the request should be processed normally
    And extra fields should be ignored

  Scenario: Request with null ID
    When I send a request with "id": null
    Then the request should be treated as notification
    And no response should be sent

  Scenario: Request with missing ID
    When I send a request without "id" field
    Then the request should be treated as notification
    And no response should be sent

  Scenario: Batch request (array of requests)
    When I send a batch of 3 requests as JSON array
    Then I should receive a batch of 3 responses
    And each response should match its request ID

  # ============================================
  # EDGE CASES
  # ============================================

  Scenario: Very large request payload
    When I send a request with 10MB of content
    Then I should receive an error response
    And the error should mention size limit

  Scenario: Request with binary data
    When I send a request containing binary data
    Then the server should handle it gracefully
    And return appropriate error if invalid

  Scenario: Request with unicode edge cases
    When I send a request with content containing:
      | character type        |
      | Zero-width joiner     |
      | Right-to-left marks   |
      | Combining characters  |
      | Emoji sequences       |
    Then the request should be processed correctly
    And the content should be preserved exactly
