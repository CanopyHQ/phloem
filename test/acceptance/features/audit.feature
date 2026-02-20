@critical
Feature: Privacy Audit
  As a privacy-conscious user
  I want to audit my Phloem installation
  So that I can verify no data leaves my machine

  Scenario: Audit shows data inventory
    Given phloem is installed
    When I run "phloem audit"
    Then the output should contain "Data Inventory"

  Scenario: Audit shows database schema
    Given phloem is installed
    And I have stored memories
    When I run "phloem audit"
    Then the output should contain table counts
