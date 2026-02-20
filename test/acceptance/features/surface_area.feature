@surface
Feature: Surface Area Coverage
  As a release engineer
  I want every MCP tool and resource exercised
  So that no endpoint ships untested

  Background:
    Given the Phloem MCP server is running
    And the memory store is initialized

  @surface @critical
  Scenario: All 14 MCP tools respond successfully
    # 1. remember
    When I call the MCP tool "remember" with content "surface area test memory"
    Then I should receive a success response
    And the response should contain a memory ID

    # 2. recall
    When I call the MCP tool "recall" with query "surface area test"
    Then I should receive a success response

    # 3. list_memories
    When I call the MCP tool "list_memories"
    Then I should receive a success response

    # 4. memory_stats
    When I call the MCP tool "memory_stats"
    Then I should receive a success response

    # 5. session_context
    When I call the MCP tool "session_context"
    Then I should receive a success response

    # 6. add_citation
    When I call the MCP tool "add_citation" with arguments:
      """
      {"memory_id":"placeholder","file_path":"/test/file.go","start_line":1,"end_line":10}
      """
    Then I should receive a success response

    # 7. get_citations
    When I call the MCP tool "get_citations" with arguments:
      """
      {"memory_id":"placeholder"}
      """
    Then I should receive a success response

    # 8. verify_citation
    When I call the MCP tool "verify_citation" with arguments:
      """
      {"citation_id":"placeholder"}
      """
    Then I should receive a success response

    # 9. verify_memory
    When I call the MCP tool "verify_memory" with arguments:
      """
      {"memory_id":"placeholder"}
      """
    Then I should receive a success response

    # 10. causal_query
    When I call the MCP tool "causal_query" with arguments:
      """
      {"memory_id":"placeholder","query_type":"neighbors"}
      """
    Then I should receive a success response

    # 11. compose
    When I call the MCP tool "compose" with arguments:
      """
      {"query_a":"test","query_b":"surface"}
      """
    Then I should receive a success response

    # 12. prefetch
    When I call the MCP tool "prefetch" with arguments:
      """
      {"context_hint":"test"}
      """
    Then I should receive a success response

    # 13. prefetch_suggest
    When I call the MCP tool "prefetch_suggest" with arguments:
      """
      {"context":"test file"}
      """
    Then I should receive a success response

    # 14. forget
    When I call the MCP tool "forget" with arguments:
      """
      {"id":"placeholder"}
      """
    Then I should receive a success response

  @surface
  Scenario: All 3 MCP resources respond
    When I request the list of available MCP resources
    Then I should receive a list containing "phloem://memories/recent"
    And I should receive a list containing "phloem://memories/stats"
    And I should receive a list containing "phloem://context/session"

    When I read the MCP resource "phloem://memories/recent"
    Then I should receive a success response

    When I read the MCP resource "phloem://memories/stats"
    Then I should receive a success response

    When I read the MCP resource "phloem://context/session"
    Then I should receive a success response

  @surface
  Scenario: CLI commands work
    Given Phloem is installed
    When I run "phloem version"
    Then the command should succeed
    When I run "phloem help"
    Then the command should succeed
    When I run "phloem status"
    Then the command should succeed
    When I run "phloem doctor"
    Then the command should succeed
    When I run "phloem audit"
    Then the command should succeed
