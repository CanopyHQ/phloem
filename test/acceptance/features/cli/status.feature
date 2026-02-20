@wip
Feature: Canopy Status Command
  As a Canopy user
  I want to view memory statistics
  So that I can understand my memory usage

  Background:
    Given Canopy is installed
    And the memory store is initialized

  @smoke @critical
  Scenario: Status with empty database
    Given no memories have been stored
    When I run "canopy status"
    Then the command should succeed
    And the output should show "Total Memories: 0"

  @smoke @critical
  Scenario: Status with memories
    Given I have stored 10 memories
    When I run "canopy status"
    Then the command should succeed
    And the output should show "Total Memories: 10"
    And the output should show "Accessible:"
    And the output should show "Last Activity:"

  Scenario: Status shows database size
    Given I have stored 100 memories
    When I run "canopy status"
    Then the output should show database size in KB or MB

  Scenario: Status shows gated memories (free tier)
    Given I am on the free tier
    And I have stored 500 memories over 30 days
    When I run "canopy status"
    Then the output should show "Accessible:" with a count
    And the output should show "Gated:" with memories older than 14 days

  Scenario: Status shows pending cloud uploads
    Given I have 5 memories pending cloud sync
    When I run "canopy status"
    Then the output should show "Pending Cloud Uploads: 5"

  Scenario: Status with Pro license
    Given I have an active Pro license
    And I have stored 500 memories
    When I run "canopy status"
    Then the output should show "Accessible: 500 (100%)"
    And the output should show "Gated: 0"

  # Unhappy Paths

  @critical
  Scenario: Status with locked database
    Given another process has locked the database
    When I run "canopy status"
    Then the command should fail
    And the error should mention "database is locked"

  Scenario: Status with corrupted database
    Given the SQLite database is corrupted
    When I run "canopy status"
    Then the command should fail
    And the error should mention database error
    And the error should suggest running "canopy doctor"

  Scenario: Status with missing data directory
    Given "~/.phloem" does not exist
    When I run "canopy status"
    Then the data directory should be created
    And the output should show "Total Memories: 0"

  Scenario: Status with read-only database
    Given the database file is read-only
    When I run "canopy status"
    Then the command should succeed for reading
    And the output should show memory statistics
