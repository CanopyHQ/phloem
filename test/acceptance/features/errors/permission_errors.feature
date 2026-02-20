@wip
Feature: Permission Error Handling
  As a Phloem user
  I want clear errors when permissions are wrong
  So that I can fix access issues

  Background:
    Given Phloem is installed

  @critical
  Scenario: Data directory not writable
    Given "~/.phloem" is not writable
    When I run "phloem remember 'test'"
    Then the command should fail
    And the error should contain "permission denied"

  @critical
  Scenario: Database file read-only for write
    Given the database file is read-only
    When I run "phloem remember 'new memory'"
    Then the command should fail
    And the error should mention permission or read-only

  Scenario: Export to unwritable path
    When I run "phloem export json /root/memories.json"
    Then the command should fail
    And the error should mention permission or denied

  Scenario: Doctor reports permission issues
    Given "~/.phloem" has incorrect permissions
    When I run "phloem doctor"
    Then the check should show ERROR
    And the output should suggest fixing permissions
