@wip
Feature: Input Validation and Error Messages
  As a Phloem user
  I want clear error messages for invalid input
  So that I can quickly fix my mistakes

  Background:
    Given Phloem is installed

  # ============================================
  # COMMAND LINE ARGUMENT VALIDATION
  # ============================================

  @critical
  Scenario: Missing required argument
    When I run "phloem export"
    Then the command should fail with exit code 1
    And the error should mention "missing" or "required"
    And the error should show usage information

  @critical
  Scenario: Too many arguments
    When I run "phloem status extra unexpected args"
    Then the command should fail
    And the error should mention "unexpected argument"

  @critical
  Scenario: Invalid flag
    When I run "phloem status --nonexistent-flag"
    Then the command should fail
    And the error should mention "unknown flag"
    And the error should suggest valid flags

  Scenario: Flag without value
    When I run "phloem export --format"
    Then the command should fail
    And the error should mention "flag needs an argument"

  # ============================================
  # FILE PATH VALIDATION
  # ============================================

  @critical
  Scenario: File not found
    When I run "phloem graft unpack /nonexistent/file.graft"
    Then the command should fail
    And the error should mention "not found" or "no such file"

  Scenario: Directory instead of file
    When I run "phloem graft unpack /tmp/"
    Then the command should fail
    And the error should mention "is a directory"

  Scenario: Path with special characters
    When I run "phloem export json '/tmp/file with spaces.json'"
    Then the command should succeed
    And the file should be created correctly

  Scenario: Relative path handling
    When I run "phloem export json ./memories.json"
    Then the command should succeed
    And the file should be created in current directory

  Scenario: Home directory expansion
    When I run "phloem export json ~/memories.json"
    Then the command should succeed
    And the file should be created in home directory

  # ============================================
  # MEMORY ID VALIDATION
  # ============================================

  @critical
  Scenario: Invalid memory ID format
    When I run "phloem verify not-a-valid-id!!!"
    Then the command should fail
    And the error should mention "invalid" and "ID"

  Scenario: Memory ID not found
    When I run "phloem verify abc123def456"
    Then the command should fail
    And the error should mention "not found"

  Scenario: Empty memory ID
    When I run "phloem verify ''"
    Then the command should fail
    And the error should mention "required" or "empty"

  # ============================================
  # CONTENT VALIDATION
  # ============================================

  Scenario: Empty content for remember
    When I call MCP tool "remember" with empty content
    Then the tool should return an error
    And the error should mention "content" and "required"

  Scenario: Very long content (>1MB)
    When I call MCP tool "remember" with 2MB of content
    Then the tool should return an error
    And the error should mention "too large" or "limit"

  Scenario: Content with null bytes
    When I call MCP tool "remember" with content containing null bytes
    Then the null bytes should be stripped or escaped
    And the memory should be stored successfully

  Scenario: Unicode content
    When I call MCP tool "remember" with content "日本語 emoji 中文"
    Then the memory should be stored successfully
    And recall should return the exact content

  # ============================================
  # TAG VALIDATION
  # ============================================

  Scenario: Invalid tag format
    When I call MCP tool "remember" with tags "invalid tag with spaces"
    Then the tag should be normalized or rejected
    And the error should explain valid tag format

  Scenario: Too many tags
    When I call MCP tool "remember" with 100 tags
    Then the tool should return an error
    And the error should mention "too many tags"

  Scenario: Empty tag in list
    When I call MCP tool "remember" with tags "valid,,also-valid"
    Then empty tags should be ignored
    And valid tags should be applied

  Scenario: Tag with special characters
    When I call MCP tool "remember" with tags "c++,c#,node.js"
    Then special characters should be handled correctly

  # ============================================
  # JSON INPUT VALIDATION
  # ============================================

  @critical
  Scenario: Invalid JSON in ChatGPT import
    Given I have a file with invalid JSON
    When I run "phloem import chatgpt invalid.json"
    Then the command should fail
    And the error should mention "JSON" and "parse"
    And the error should show line/position if possible

  Scenario: Valid JSON but wrong structure
    Given I have a JSON file with wrong structure
    When I run "phloem import chatgpt wrong-structure.json"
    Then the command should fail
    And the error should mention "unexpected format"
    And the error should describe expected format

  # ============================================
  # GRAFT FILE VALIDATION
  # ============================================

  @critical
  Scenario: Invalid graft file format
    Given I have a file that is not a valid graft
    When I run "phloem graft unpack invalid.graft"
    Then the command should fail
    And the error should mention "invalid" and "graft"

  Scenario: Corrupted graft file
    Given I have a corrupted graft file
    When I run "phloem graft unpack corrupted.graft"
    Then the command should fail
    And the error should mention "corrupted" or "invalid"

  Scenario: Graft file with unsupported version
    Given I have a graft file with version 99.0
    When I run "phloem graft unpack future.graft"
    Then the command should fail
    And the error should mention "version" and "unsupported"
