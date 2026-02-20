@wip
Feature: Memory Tags
  As an AI assistant
  I want to tag memories and filter by tags
  So that I can organize and narrow recall

  Background:
    Given the Phloem MCP server is running
    And the memory store is initialized

  @critical
  Scenario: Store memory with tags
    When I call the MCP tool "remember" with:
      | content | Tagged decision about architecture |
      | tags   | decision, architecture             |
    Then I should receive a success response
    And the memory should be tagged with "decision"
    And the memory should be tagged with "architecture"

  @critical
  Scenario: Filter recall by tags
    Given I have stored memories with various tags:
      | content                 | tags           |
      | Architecture decision  | decision, arch |
      | Bug fix for login       | bugfix, auth   |
    When I call the MCP tool "recall" with query "decision" and tags "decision"
    Then I should only receive memories tagged with "decision"
    And the results should contain "Architecture"

  Scenario: Export graft with tag filter
    Given I have stored 5 memories with tags "architecture"
    And I have stored 3 memories with tags "bugfix"
    When I export a graft with tags "architecture" to "/tmp/arch.graft"
    Then the graft file should be created
    And the graft should contain 5 memories

  Scenario: Multiple tags on one memory
    When I call the MCP tool "remember" with:
      | content | Multi-tag note |
      | tags    | a, b, c        |
    Then I should receive a success response
    And the memory should be tagged with "a"
    And the memory should be tagged with "b"

  # Edge cases

  Scenario: Recall with non-existent tag
    Given I have stored memories with tags "existing"
    When I call the MCP tool "recall" with query "anything" and tags "nonexistent"
    Then I should receive empty or zero results
    And I should receive a success response
