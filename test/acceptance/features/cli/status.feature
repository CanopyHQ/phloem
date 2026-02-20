@wip
Feature: Phloem Status Command
  As a Phloem user
  I want to view memory statistics
  So that I can understand my memory usage

  Background:
    Given Phloem is installed
    And the memory store is initialized

  @smoke @critical
  Scenario: Status with empty database
    Given no memories have been stored
    When I run "phloem status"
    Then the command should succeed
    And the output should show "Total Memories: 0"

  @smoke @critical
  Scenario: Status with memories
    Given I have stored 10 memories
    When I run "phloem status"
    Then the command should succeed
    And the output should show "Total Memories: 10"
    And the output should show "Accessible:"
    And the output should show "Last Activity:"

  Scenario: Status shows database size
    Given I have stored 100 memories
    When I run "phloem status"
    Then the output should show database size in KB or MB

  # Unhappy Paths

  @critical
  Scenario: Status with locked database
    Given another process has locked the database
    When I run "phloem status"
    Then the command should fail
    And the error should mention "database is locked"

  Scenario: Status with corrupted database
    Given the SQLite database is corrupted
    When I run "phloem status"
    Then the command should fail
    And the error should mention database error
    And the error should suggest running "phloem doctor"

  Scenario: Status with missing data directory
    Given "~/.phloem" does not exist
    When I run "phloem status"
    Then the data directory should be created
    And the output should show "Total Memories: 0"

  Scenario: Status with read-only database
    Given the database file is read-only
    When I run "phloem status"
    Then the command should succeed for reading
    And the output should show memory statistics
