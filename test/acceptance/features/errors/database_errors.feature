@wip
Feature: Database Error Handling
  As a Phloem user
  I want graceful handling of database errors
  So that I don't lose data and can recover from issues

  Background:
    Given Phloem is installed

  @critical
  Scenario: Database locked by another process
    Given the database is locked by another phloem process
    When I run "phloem status"
    Then the command should fail with exit code 1
    And the error should contain "database is locked"
    And the error should suggest closing other Phloem instances

  @critical
  Scenario: Database file corrupted
    Given the database file is corrupted
    When I run "phloem status"
    Then the command should fail
    And the error should contain "database"
    And the error should suggest running "phloem doctor"

  Scenario: Database file missing
    Given the database file does not exist
    When I run "phloem status"
    Then a new database should be created
    And the command should succeed
    And the output should show "Total Memories: 0"

  Scenario: Database directory not writable
    Given "~/.phloem" is not writable
    When I run "phloem remember 'test'"
    Then the command should fail
    And the error should contain "permission denied"

  Scenario: Disk full during write
    Given the disk is full
    When I try to store a memory
    Then the command should fail
    And the error should mention disk space
    And no partial data should be written

  Scenario: Database schema migration
    Given I have an old database schema version
    When I run "phloem status"
    Then the database should be migrated automatically
    And the command should succeed
    And existing data should be preserved

  Scenario: Concurrent write operations
    Given I run 10 concurrent remember operations
    Then all operations should complete
    And no data should be lost
    And no database corruption should occur

  Scenario: Recovery from WAL corruption
    Given the SQLite WAL file is corrupted
    When I run "phloem status"
    Then the command should attempt recovery
    And the error should provide recovery instructions

  Scenario: Database backup before migration
    Given I have an old database schema version
    When migration is triggered
    Then a backup should be created at "~/.phloem/memories.db.backup"
    And the migration should proceed
