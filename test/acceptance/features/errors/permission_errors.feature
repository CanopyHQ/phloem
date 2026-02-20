@wip
Feature: Permission Error Handling
  As a Canopy user
  I want clear errors when permissions are wrong
  So that I can fix access issues

  Background:
    Given Canopy is installed

  @critical
  Scenario: Data directory not writable
    Given "~/.phloem" is not writable
    When I run "canopy remember 'test'"
    Then the command should fail
    And the error should contain "permission denied"

  @critical
  Scenario: Database file read-only for write
    Given the database file is read-only
    When I run "canopy remember 'new memory'"
    Then the command should fail
    And the error should mention permission or read-only

  Scenario: Export to unwritable path
    When I run "canopy export json /root/memories.json"
    Then the command should fail
    And the error should mention permission or denied

  Scenario: Install native manifest without write access
    Given I do not have write access to the native messaging config directory
    When I run "canopy install-native"
    Then the command should fail
    And the error should suggest running with appropriate permissions

  Scenario: Doctor reports permission issues
    Given "~/.phloem" has incorrect permissions
    When I run "canopy doctor"
    Then the check should show ERROR
    And the output should suggest fixing permissions
