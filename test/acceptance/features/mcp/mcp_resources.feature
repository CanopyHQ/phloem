@wip
Feature: MCP Resources Discovery and Read
  As an MCP client
  I want to discover and read Phloem resources
  So that I can access memory content via resource URIs

  Background:
    Given the Phloem MCP server is running
    And the memory store is initialized

  @smoke @critical
  Scenario: List resources includes phloem schema
    When I request the list of available MCP resources
    Then I should receive a list containing "phloem" or "memories"

  @critical
  Scenario: Read recent memories resource
    Given I have stored 3 memories
    When I read the MCP resource "phloem://memories/recent"
    Then I should receive valid content
    And the response should be valid JSON

  Scenario: Read memory stats resource
    When I read the MCP resource "phloem://memories/stats"
    Then I should receive valid content
    And the response should contain total_memories or database_size

  Scenario: Invalid resource URI returns error
    When I read the MCP resource "phloem://invalid/path"
    Then I should receive an error or empty response

  Scenario: Resources list is non-empty when memories exist
    Given I have stored 1 memory
    When I request the list of available MCP resources
    Then I should receive at least one resource URI or template
