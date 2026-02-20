@wip
Feature: Phloem Version Command
  As a Phloem user
  I want to see the installed version
  So that I can verify installation and report bugs

  Background:
    Given Phloem is installed

  @smoke @critical
  Scenario: Version shows semantic version
    When I run "phloem version"
    Then the command should succeed
    And the output should match pattern "phloem v\\d+\\.\\d+\\.\\d+"

  @critical
  Scenario: Version shows commit when built from source
    Given Phloem was built from source
    When I run "phloem version"
    Then the output should contain "commit" or "build"

  Scenario: Version is parseable
    When I run "phloem version"
    Then the command should succeed
    And the output should not be empty
    And the output should contain "phloem"

  Scenario: Version works without data directory
    Given "~/.phloem" does not exist
    When I run "phloem version"
    Then the command should succeed
    And the output should show version information
