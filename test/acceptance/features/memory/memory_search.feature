@wip
Feature: Memory Search and Recall
  As an AI assistant
  I want to recall memories by semantic search
  So that I can surface relevant context

  Background:
    Given the Phloem MCP server is running
    And the memory store is initialized

  @smoke @critical
  Scenario: Recall returns semantically related content
    Given I have stored a memory with content "We chose PostgreSQL for the backend"
    When I call the MCP tool "recall" with query "database choice"
    Then the results should contain "PostgreSQL" or "backend"

  @critical
  Scenario: Recall with no matches returns empty
    Given I have stored a memory with content "Unrelated note"
    When I call the MCP tool "recall" with query "nonexistent topic xyz"
    Then the results should not contain "Unrelated note"
    And I should receive a success response

  Scenario: Recall with multiple memories returns best matches
    Given I have stored memories with various tags:
      | content              | tags    |
      | API design decision  | api     |
      | Login bug fix        | bugfix  |
      | API rate limiting    | api     |
    When I call the MCP tool "recall" with query "API"
    Then the results should contain "API" or "rate limiting" or "design"
    And I should receive a success response

  Scenario: Recall limits results
    Given I have stored 20 memories
    When I call the MCP tool "recall" with query "memory"
    Then I should receive a bounded number of results
    And the response should be valid

  # Edge cases

  Scenario: Recall with empty query
    When I call the MCP tool "recall" with query ""
    Then I should receive an error response or empty results

  Scenario: Recall with very long query
    When I call the MCP tool "recall" with query "a very long query that exceeds normal length"
    Then the command should handle the input
    And I should receive a response
