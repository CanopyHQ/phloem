Feature: Phloem Grafts - Shareable Memory Bundles
  As a Phloem user
  I want to export and import memory grafts
  So that I can share context with others and create viral pull

  Background:
    Given the Phloem system is initialized
    And the memory store is initialized

  @smoke @critical @brew_gate
  Scenario: Export memories as graft
    Given I have stored 5 memories with tags "architecture,patterns"
    When I export a graft with tags "architecture,patterns" to "test-arch.graft"
    Then the graft file should be created
    And the graft should contain 5 memories
    And the graft manifest should have name "Test Export"

  @smoke @critical @brew_gate
  Scenario: Import graft file
    Given I have a graft file "test-arch.graft" with 5 memories
    When I import the graft file "test-arch.graft"
    Then 5 memories should be imported
    And the imported memories should have tag "architecture"

  @critical @brew_gate
  Scenario: Inspect graft without importing
    Given I have a graft file "test-arch.graft" with 10 memories
    When I inspect the graft file "test-arch.graft"
    Then I should see the graft manifest
    And the manifest should show 10 memories
    And no memories should be imported

  @critical @brew_gate
  Scenario: Graft deduplication on import
    Given I have stored a memory with content "Duplicate test"
    And I have a graft file "duplicate.graft" containing "Duplicate test"
    When I import the graft file "duplicate.graft"
    Then the duplicate memory should not be created
    And only unique memories should be imported

  @wip
  Scenario: Export graft with citations
    Given I have stored memories with citations
    When I export a graft including citations
    Then the graft should contain citation data
    And citations should be preserved on import

  @wip
  Scenario: Graft file format validation
    Given I have an invalid graft file "invalid.graft"
    When I try to import "invalid.graft"
    Then I should receive an error
    And the error should indicate invalid format
