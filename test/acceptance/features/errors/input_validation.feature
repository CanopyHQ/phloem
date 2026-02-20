@wip
Feature: Input Validation and Error Messages
  As a Canopy user
  I want clear error messages for invalid input
  So that I can quickly fix my mistakes

  Background:
    Given Canopy is installed

  # ============================================
  # COMMAND LINE ARGUMENT VALIDATION
  # ============================================

  @critical
  Scenario: Missing required argument
    When I run "canopy export"
    Then the command should fail with exit code 1
    And the error should mention "missing" or "required"
    And the error should show usage information

  @critical
  Scenario: Too many arguments
    When I run "canopy status extra unexpected args"
    Then the command should fail
    And the error should mention "unexpected argument"

  @critical
  Scenario: Invalid flag
    When I run "canopy status --nonexistent-flag"
    Then the command should fail
    And the error should mention "unknown flag"
    And the error should suggest valid flags

  Scenario: Flag without value
    When I run "canopy export --format"
    Then the command should fail
    And the error should mention "flag needs an argument"

  Scenario: Invalid flag value
    When I run "canopy ingest --duration=invalid"
    Then the command should fail
    And the error should mention "invalid" and "duration"

  # ============================================
  # FILE PATH VALIDATION
  # ============================================

  @critical
  Scenario: File not found
    When I run "canopy graft unpack /nonexistent/file.graft"
    Then the command should fail
    And the error should mention "not found" or "no such file"

  Scenario: Directory instead of file
    When I run "canopy graft unpack /tmp/"
    Then the command should fail
    And the error should mention "is a directory"

  Scenario: File instead of directory
    When I run "canopy import claude /tmp/somefile.txt"
    Then the command should fail
    And the error should mention "not a directory"

  Scenario: Path with special characters
    When I run "canopy export json '/tmp/file with spaces.json'"
    Then the command should succeed
    And the file should be created correctly

  Scenario: Relative path handling
    When I run "canopy export json ./memories.json"
    Then the command should succeed
    And the file should be created in current directory

  Scenario: Home directory expansion
    When I run "canopy export json ~/memories.json"
    Then the command should succeed
    And the file should be created in home directory

  # ============================================
  # MEMORY ID VALIDATION
  # ============================================

  @critical
  Scenario: Invalid memory ID format
    When I run "canopy verify not-a-valid-id!!!"
    Then the command should fail
    And the error should mention "invalid" and "ID"

  Scenario: Memory ID not found
    When I run "canopy verify abc123def456"
    Then the command should fail
    And the error should mention "not found"

  Scenario: Empty memory ID
    When I run "canopy verify ''"
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
    When I call MCP tool "remember" with content "日本語 emoji 🎉 中文"
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
  # LICENSE KEY VALIDATION
  # ============================================

  @critical
  Scenario: Malformed license key
    When I run "canopy activate not-a-valid-key"
    Then the command should fail
    And the error should mention "invalid" and "license"

  Scenario: License key with wrong signature
    When I run "canopy activate <tampered_key>"
    Then the command should fail
    And the error should mention "signature" or "invalid"

  Scenario: Empty license key
    When I run "canopy activate ''"
    Then the command should fail
    And the error should mention "required"

  # ============================================
  # API KEY VALIDATION
  # ============================================

  Scenario: Malformed API key
    Given PHLOEM_API_KEY is set to "not-valid"
    When I run "canopy sync"
    Then the command should fail
    And the error should mention "invalid" and "API key"

  Scenario: Empty API key
    Given PHLOEM_API_KEY is set to ""
    When I run "canopy sync"
    Then the command should fail
    And the error should mention "API key" and "required"

  # ============================================
  # JSON INPUT VALIDATION
  # ============================================

  @critical
  Scenario: Invalid JSON in ChatGPT import
    Given I have a file with invalid JSON
    When I run "canopy import chatgpt invalid.json"
    Then the command should fail
    And the error should mention "JSON" and "parse"
    And the error should show line/position if possible

  Scenario: Valid JSON but wrong structure
    Given I have a JSON file with wrong structure
    When I run "canopy import chatgpt wrong-structure.json"
    Then the command should fail
    And the error should mention "unexpected format"
    And the error should describe expected format

  # ============================================
  # GRAFT FILE VALIDATION
  # ============================================

  @critical
  Scenario: Invalid graft file format
    Given I have a file that is not a valid graft
    When I run "canopy graft unpack invalid.graft"
    Then the command should fail
    And the error should mention "invalid" and "graft"

  Scenario: Corrupted graft file
    Given I have a corrupted graft file
    When I run "canopy graft unpack corrupted.graft"
    Then the command should fail
    And the error should mention "corrupted" or "invalid"

  Scenario: Graft file with unsupported version
    Given I have a graft file with version 99.0
    When I run "canopy graft unpack future.graft"
    Then the command should fail
    And the error should mention "version" and "unsupported"

  # ============================================
  # DURATION VALIDATION
  # ============================================

  Scenario: Valid duration formats
    When I run "canopy ingest 24h"
    Then the command should succeed
    When I run "canopy ingest 7d"
    Then the command should succeed
    When I run "canopy ingest 2w"
    Then the command should succeed

  Scenario: Invalid duration format
    When I run "canopy ingest 5x"
    Then the command should fail
    And the error should mention "invalid duration"
    And the error should show valid formats

  Scenario: Negative duration
    When I run "canopy ingest -1d"
    Then the command should fail
    And the error should mention "positive" or "invalid"

  Scenario: Zero duration
    When I run "canopy ingest 0h"
    Then the command should fail
    And the error should mention "must be greater than zero"
