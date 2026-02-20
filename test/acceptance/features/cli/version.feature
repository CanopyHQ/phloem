@wip
Feature: Canopy Version Command
  As a Canopy user
  I want to see the installed version
  So that I can verify installation and report bugs

  Background:
    Given Canopy is installed

  @smoke @critical
  Scenario: Version shows semantic version
    When I run "canopy version"
    Then the command should succeed
    And the output should match pattern "canopy v\\d+\\.\\d+\\.\\d+"

  @critical
  Scenario: Version shows commit when built from source
    Given Canopy was built from source
    When I run "canopy version"
    Then the output should contain "commit" or "build"

  Scenario: Version is parseable
    When I run "canopy version"
    Then the command should succeed
    And the output should not be empty
    And the output should contain "canopy"

  Scenario: Version works without data directory
    Given "~/.phloem" does not exist
    When I run "canopy version"
    Then the command should succeed
    And the output should show version information
