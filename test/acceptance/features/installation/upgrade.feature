@wip
Feature: Canopy Upgrade
  As a Canopy user
  I want to upgrade to a newer version
  So that I get new features and fixes without losing data

  Background:
    Given Canopy is installed

  @critical
  Scenario: Upgrade from previous version via Homebrew
    Given Canopy version "1.0.0" is installed
    And version "1.0.1" is available in the tap
    When I run "brew upgrade canopy"
    Then the command should succeed
    And running "canopy version" should show "1.0.1"
    And my existing memories should be preserved

  Scenario: Upgrade when already on latest
    Given I have the latest Canopy version installed
    When I run "brew upgrade canopy"
    Then the command should succeed
    And the output should contain "already up-to-date"

  Scenario: Upgrade preserves data directory
    Given I have stored 10 memories
    When I run "brew upgrade canopy"
    Then the command should succeed
    When I run "canopy status"
    Then the output should show "Total Memories: 10"

  # Unhappy paths

  Scenario: Upgrade fails when not installed via Homebrew
    Given Canopy was installed manually
    When I run "brew upgrade canopy"
    Then the command should fail
    And the error should mention "not installed"
