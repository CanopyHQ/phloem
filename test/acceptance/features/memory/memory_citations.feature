@wip
Feature: Memory Citations
  As an AI assistant
  I want memories to support citations and provenance
  So that I can attribute and verify sources

  Background:
    Given the Phloem MCP server is running
    And the memory store is initialized

  @critical
  Scenario: Store memory with citation context
    Given I have stored memories with citations
    When I call the MCP tool "recall" with query "cited"
    Then I should receive a success response
    And the response may include citation or source metadata

  @critical
  Scenario: Export graft preserves citations
    Given I have stored memories with citations
    When I export a graft including citations
    Then the graft should contain citation data
    When I import the graft
    Then citations should be preserved on import

  Scenario: Recall result can include citation
    Given I have stored a memory with content "Source document says X"
    And the memory has citation metadata
    When I call the MCP tool "recall" with query "Source document"
    Then the results should contain "Source document says X"
    And the response may include citation or source

  Scenario: Graft manifest includes citation info
    Given I have stored memories with citations
    When I export a graft including citations
    Then the graft file should be created
    And the graft should contain citation data

  # Edge cases

  Scenario: Memory without citation still works
    When I call the MCP tool "remember" with content "Plain memory"
    Then I should receive a success response
    And recall should return it without requiring citation
