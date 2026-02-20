Feature: Memory Storage and Recall
  As an AI assistant
  I want to store and recall memories
  So that I can maintain context across sessions

  Background:
    Given the Phloem MCP server is running
    And the memory store is initialized

  @smoke @critical @brew_gate
  Scenario: Store a simple memory
    When I call the MCP tool "remember" with content "Test memory content"
    Then I should receive a success response
    And the response should contain a memory ID

  @smoke @critical @brew_gate
  Scenario: Recall a stored memory
    Given I have stored a memory with content "The sky is blue"
    When I call the MCP tool "recall" with query "sky color"
    Then the results should contain "The sky is blue"

  @critical @wip
  Scenario: Store memory with tags
    When I call the MCP tool "remember" with:
      | content | Important decision about architecture |
      | tags    | decision, architecture               |
      | context | Project planning session             |
    Then I should receive a success response
    And the memory should be tagged with "decision"
    And the memory should be tagged with "architecture"

  @critical @wip
  Scenario: Filter recall by tags
    Given I have stored memories with various tags:
      | content                    | tags              |
      | Architecture decision      | decision, arch    |
      | Bug fix for login          | bugfix, auth      |
      | Feature planning           | planning, feature |
    When I call the MCP tool "recall" with query "decision" and tags "decision"
    Then I should only receive memories tagged with "decision"

  @wip
  Scenario: List recent memories
    Given I have stored 5 memories
    When I call the MCP tool "list_memories" with limit 3
    Then I should receive exactly 3 memories
    And they should be ordered by creation date descending

  @wip
  Scenario: Forget a memory
    Given I have stored a memory with content "Temporary note"
    And I have the memory ID
    When I call the MCP tool "forget" with the memory ID
    Then I should receive a success response
    When I call the MCP tool "recall" with query "Temporary note"
    Then the results should not contain the forgotten memory

  @smoke @wip
  Scenario: Get memory statistics
    Given I have stored 10 memories
    When I call the MCP tool "memory_stats"
    Then I should receive statistics including:
      | field           |
      | total_memories  |
      | database_size   |
      | last_activity   |

  @critical @wip
  Scenario: Semantic similarity search
    Given I have stored memories:
      | content                                          |
      | The quick brown fox jumps over the lazy dog      |
      | Python is a programming language                 |
      | Machine learning uses neural networks            |
    When I call the MCP tool "recall" with query "AI and deep learning"
    Then the top result should be about "neural networks"
    And the similarity score should be greater than 0.1

  @wip
  Scenario: Handle empty recall results
    When I call the MCP tool "recall" with query "xyzzy nonexistent gibberish"
    Then I should receive an empty results list
    And the response should indicate 0 memories found

  @critical @wip
  Scenario: Session context preload
    Given I have stored memories with tags:
      | content                      | tags                    |
      | Critical security decision   | critical, decision      |
      | Architecture overview        | architecture            |
      | Yesterday's conversation     | conversation            |
    When I call the MCP tool "session_context"
    Then I should receive formatted context
    And the context should include critical items
    And the context should include recent activity
